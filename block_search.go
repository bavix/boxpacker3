package boxpacker3

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"time"
)

type Search struct {
	nodes     int
	branching int
	order     ItemSorter
}

const (
	searchWidening = 1.4142135623730951

	searchRounding = 0.5

	searchStartWidth = 1

	searchMaxWidth = 32

	searchBlocksPerSpace = 2
)

func NewSearch(nodes, branching int) (*Search, error) {
	if nodes < 1 {
		return nil, fmt.Errorf("%w: search budget %d nodes", ErrInvalidSetting, nodes)
	}

	if branching < 1 {
		return nil, fmt.Errorf("%w: branching factor %d", ErrInvalidSetting, branching)
	}

	return &Search{nodes: nodes, branching: branching, order: OrderDecreasing}, nil
}

func (s *Search) Name() string {
	return fmt.Sprintf("Search(%d,%d)", s.nodes, s.branching)
}

func (s *Search) Nodes() int {
	return s.nodes
}

func (s *Search) Branching() int {
	return s.branching
}

func (s *Search) Pack(ctx context.Context, problem *Problem) (*Packing, error) {
	items := sortItems(compact(problem.pieces), s.order)
	sortedBoxes, packing := prepareData(problem.boxes, items, problem)
	remaining := items

	for _, box := range sortedBoxes {
		err := checkContext(ctx)
		if err != nil {
			return nil, err
		}

		if len(remaining) == 0 || box == nil {
			continue
		}

		var spent int

		remaining, spent = s.fillBox(ctx, problem, box, remaining)
		problem.spend(spent)
	}

	packing.unfit = append(packing.unfit, remaining...)

	return packing, nil
}

func (s *Search) fillBox(ctx context.Context, problem *Problem, box *Container, items []*piece) ([]*piece, int) {
	plain := plainFill(ctx, box, items)

	best, spent := s.search(ctx, problem, box, items)
	if best != nil {
		commitState(box, best)

		left, err := fillGaps(ctx, box, best.stock.remaining(), true)
		if err == nil && (plain == nil || box.itemsVolume >= plain.volume) {
			return left, spent
		}
	}

	if plain != nil {
		return plain.applyTo(box, items), spent
	}

	left, _ := fillGaps(ctx, box, items, true)

	return left, spent
}

type state struct {
	box *Container

	origins []*piece
	stock   *stock
	placed  []placedBlock
	score   float64
}

type placedBlock struct {
	block    block
	position Pivot
}

func (s *state) fork() *state {
	box := clonePtr(s.box)
	box.items = slices.Clone(box.items)

	return &state{
		box:     box,
		origins: slices.Clone(s.origins),
		stock:   s.stock.clone(),
		placed:  slices.Clone(s.placed),
		score:   s.score,
	}
}

func (s *stock) clone() *stock {
	left := make(map[blockKey][]*piece, len(s.left))
	for key, pieces := range s.left {
		left[key] = slices.Clone(pieces)
	}

	return &stock{kinds: s.kinds, sample: s.sample, left: left}
}

func (s *Search) search(ctx context.Context, problem *Problem, box *Container, items []*piece) (*state, int) {
	deadline, limited := problem.Deadline()
	spent := 0

	var best *state

	for width := searchStartWidth; width <= searchMaxWidth; width = wider(width) {
		if spent >= s.nodes || checkContext(ctx) != nil {
			break
		}

		if limited && !time.Now().Before(deadline) {
			break
		}

		found, used := s.run(ctx, box, items, width, s.nodes-spent)
		spent += used

		if found != nil && fuller(found, best) {
			best = found
		}
	}

	return best, spent
}

func wider(width int) int {
	next := int(float64(width)*searchWidening + searchRounding)
	if next <= width {
		return width + 1
	}

	return next
}

func (s *Search) run(ctx context.Context, box *Container, items []*piece, width, budget int) (*state, int) {
	fresh := clonePtr(box)
	fresh.reset()

	beam := []*state{{box: fresh, origins: nil, stock: stockOf(items), placed: nil, score: 0}}

	var (
		best  *state
		spent int
	)

	for len(beam) > 0 && spent < budget && checkContext(ctx) == nil {
		var next []*state

		next, best, spent = s.level(beam, best, spent, budget)
		beam = keepBest(next, width)
	}

	for _, leftover := range beam {
		if fuller(leftover, best) {
			best = leftover
		}
	}

	return best, spent
}

func (s *Search) level(beam []*state, best *state, spent, budget int) ([]*state, *state, int) {
	next := make([]*state, 0, len(beam)*s.branching)

	for _, current := range beam {
		if spent >= budget {
			break
		}

		spent++

		children := s.expand(current)
		if len(children) == 0 && fuller(current, best) {
			best = current
		}

		next = append(next, children...)
	}

	return next, best, spent
}

func fuller(current, best *state) bool {
	return best == nil || current.box.itemsVolume > best.box.itemsVolume
}

func keepBest(states []*state, width int) []*state {
	slices.SortStableFunc(states, func(a, b *state) int {
		return cmp.Or(
			cmp.Compare(b.score, a.score),
			cmp.Compare(b.box.itemsVolume, a.box.itemsVolume),
		)
	})

	return states[:min(width, len(states))]
}

func (s *Search) expand(current *state) []*state {
	for !current.stock.empty() {
		free, at, found := nextSpace(current.box)
		if !found {
			return nil
		}

		children := s.childrenIn(current, free)
		if len(children) > 0 {
			return children
		}

		current.box.dropSpace(at)
	}

	return nil
}

func (s *Search) childrenIn(current *state, free space) []*state {
	children := make([]*state, 0, s.branching)

	for _, candidate := range s.candidates(current, free) {
		child := current.fork()

		origins, laid := child.box.placeBlock(candidate.block, child.stock, candidate.position)
		if !laid {
			continue
		}

		child.origins = append(child.origins, origins...)
		child.placed = append(child.placed, candidate)
		child.score = completion(child)

		children = append(children, child)

		if len(children) >= s.branching {
			break
		}
	}

	return children
}

func (s *Search) candidates(current *state, free space) []placedBlock {
	found := make([]placedBlock, 0, len(current.stock.kinds)*searchBlocksPerSpace)

	for _, key := range current.stock.kinds {
		if current.stock.count(key) == 0 {
			continue
		}

		for _, candidate := range blocksAt(current.box, current.stock, key, free.origin, searchBlocksPerSpace) {
			if !free.holds(candidate.size()) {
				continue
			}

			found = append(found, placedBlock{block: candidate, position: free.origin})
		}
	}

	slices.SortStableFunc(found, func(a, b placedBlock) int {
		return cmp.Compare(b.block.volume(), a.block.volume())
	})

	return found
}

func nextSpace(box *Container) (space, int, bool) {
	var (
		best  space
		at    int
		found bool
	)

	for i, free := range box.spaces {
		if !found || nearerCorner(free, best) {
			best, at, found = free, i, true
		}
	}

	return best, at, found
}

func (c *Container) dropSpace(at int) {
	c.spaces = slices.Delete(c.spaces, at, at+1)
}

func nearerCorner(a, b space) bool {
	for _, axis := range [3]Axis{DepthAxis, HeightAxis, WidthAxis} {
		if a.origin[axis] != b.origin[axis] {
			return a.origin[axis] < b.origin[axis]
		}
	}

	return volumeOf(a) > volumeOf(b)
}

func volumeOf(free space) float64 {
	return free.size[WidthAxis] * free.size[HeightAxis] * free.size[DepthAxis]
}

func completion(current *state) float64 {
	scratch := current.fork()

	for !scratch.stock.empty() {
		free, at, found := nextSpace(scratch.box)
		if !found {
			break
		}

		if !fillSpace(scratch, free) {
			scratch.box.dropSpace(at)
		}
	}

	return scratch.box.itemsVolume
}

func fillSpace(current *state, free space) bool {
	for _, key := range current.stock.kinds {
		if current.stock.count(key) == 0 {
			continue
		}

		if placeBestStack(current, key, free) {
			return true
		}
	}

	return false
}

func placeBestStack(current *state, key blockKey, free space) bool {
	for _, candidate := range blocksAt(current.box, current.stock, key, free.origin, completionStacks) {
		if !free.holds(candidate.size()) {
			continue
		}

		if _, laid := current.box.placeBlock(candidate, current.stock, free.origin); laid {
			return true
		}
	}

	return false
}

const completionStacks = 3

func commitState(box *Container, best *state) {
	box.reset()

	for i, copied := range best.box.items {
		origin := best.origins[i]
		origin.position = copied.position
		origin.orientation = copied.orientation

		box.insert(origin)
	}
}

func plainFill(ctx context.Context, box *Container, items []*piece) *seededFill {
	trial := clonePtr(box)
	trial.reset()

	copies := cloneSlice(items)

	_, err := ordinaryFill(ctx, trial, copies)
	if err != nil {
		return nil
	}

	return layoutOf(trial, copies)
}

func layoutOf(trial *Container, copies []*piece) *seededFill {
	fill := &seededFill{
		placed:   make([]bool, len(copies)),
		position: make([]Pivot, len(copies)),
		rotation: make([]Orientation, len(copies)),
		volume:   trial.itemsVolume,
	}

	for i, copied := range copies {
		if !slices.Contains(trial.items, copied) {
			continue
		}

		fill.placed[i] = true
		fill.position[i] = copied.position
		fill.rotation[i] = copied.orientation
	}

	return fill
}

func (s *Search) Fill(ctx context.Context, box *Container, items []Instance) ([]Instance, error) {
	pieces := make([]*piece, 0, len(items))

	for _, instance := range items {
		if item := box.problem.pieceFor(instance); item != nil {
			pieces = append(pieces, item)
		}
	}

	err := checkContext(ctx)
	if err != nil {
		return items, err
	}

	left, spent := s.fillBox(ctx, box.problem, box, sortItems(pieces, s.order))
	box.problem.spend(spent)

	remaining := make([]Instance, 0, len(left))
	for _, item := range left {
		remaining = append(remaining, Instance{Item: item.Item, Index: item.index})
	}

	return remaining, nil
}
