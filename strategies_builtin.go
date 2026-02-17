package boxpacker3

import (
	"cmp"
	"context"
	"slices"
	"strings"
)

const (
	perfectFitRatio = 1e-6
)

func compact[T any](values []*T) []*T {
	kept := make([]*T, 0, len(values))

	for _, value := range values {
		if value != nil {
			kept = append(kept, value)
		}
	}

	return kept
}

func expandQuantities(boxes []*Container, itemCount int) []*Container {
	expanded := make([]*Container, 0, len(boxes))

	for _, box := range boxes {
		copies := min(max(box.quantity, 1), max(itemCount, 1))

		expanded = append(expanded, box)

		for index := 1; index < copies; index++ {
			copied := clonePtr(box)
			copied.index = index
			expanded = append(expanded, copied)
		}
	}

	return expanded
}

func isPerfectFit(box *Container) bool {
	if box == nil || box.volume <= 0 {
		return false
	}

	return box.remainingVolume() < box.volume*perfectFitRatio
}

func prepareData(inputBoxes []*Box, items []*piece, problem *Problem) ([]*Container, *Packing) {
	boxes := expandQuantities(problem.open(compact(inputBoxes)), len(items))

	slices.SortStableFunc(boxes, byBoxVolumeAsc)

	sortedBoxes := preferredSort(boxes, items)

	return sortedBoxes, newPacking(sortedBoxes, make([]*piece, 0, len(items))).on(problem.boxes).by(problem)
}

func checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func runFirstFit(ctx context.Context, boxes []*Box, items []*piece, problem *Problem) (*Packing, error) {
	sortedBoxes, result := prepareData(boxes, items, problem)
	remainingItems := items

	for _, box := range sortedBoxes {
		err := checkContext(ctx)
		if err != nil {
			return nil, err
		}

		if len(remainingItems) == 0 {
			break
		}

		var errPack error

		remainingItems, errPack = packToBox(ctx, box, remainingItems)
		if errPack != nil {
			return nil, errPack
		}
	}

	result.unfit = append(result.unfit, remainingItems...)

	return result, nil
}

func runFullestBox(ctx context.Context, boxes []*Box, items []*piece, problem *Problem) (*Packing, error) {
	sortedBoxes, result := prepareData(boxes, items, problem)
	remaining := items
	search := newFullestSearch(sortedBoxes)

	for len(remaining) > 0 {
		index, err := search.pick(ctx, sortedBoxes, remaining)
		if err != nil {
			return nil, err
		}

		if index < 0 {
			break
		}

		before := len(remaining)

		remaining, err = packToBox(ctx, sortedBoxes[index], remaining)
		if err != nil {
			return nil, err
		}

		if len(remaining) == before {
			break
		}
	}

	result.unfit = append(result.unfit, remaining...)

	return result, nil
}

type fullestSearch struct{}

type boxKind struct {
	id                     string
	width, height, depth   float64
	maxWeight, emptyWeight float64
	accepts                string
}

func kindOf(box *Container) boxKind {
	return boxKind{
		id:          box.id,
		width:       box.width,
		height:      box.height,
		depth:       box.depth,
		maxWeight:   box.maxWeight,
		emptyWeight: box.emptyWeight,
		accepts:     strings.Join(box.accepts, "\x00"),
	}
}

func newFullestSearch(_ []*Container) *fullestSearch {
	return &fullestSearch{}
}

func (s *fullestSearch) prefer(boxes []*Container, best, candidate int, bestVolume, measured float64) bool {
	if best < 0 {
		return true
	}

	if sameReading(measured, bestVolume) {
		return boxes[candidate].volume < boxes[best].volume
	}

	return measured > bestVolume
}

func (s *fullestSearch) pick(ctx context.Context, boxes []*Container, remaining []*piece) (int, error) {
	kinds, bound := emptyKindsByBound(boxes, remaining)
	if len(kinds) == 0 {
		return -1, nil
	}

	if len(kinds) == 1 {
		return kinds[0], nil
	}

	best, bestVolume := -1, 0.0

	for position, i := range kinds {
		err := checkContext(ctx)
		if err != nil {
			return -1, err
		}

		measured, err := s.measure(ctx, boxes[i], remaining)
		if err != nil {
			return -1, err
		}

		if measured > 0 && s.prefer(boxes, best, i, bestVolume, measured) {
			best, bestVolume = i, measured
		}

		if position+1 >= len(kinds) || bestVolume > bound[position+1] {
			break
		}
	}

	return best, nil
}

func (s *fullestSearch) measure(ctx context.Context, box *Container, remaining []*piece) (float64, error) {
	trial := clonePtr(box)
	trial.reset()

	placements := snapshot(remaining)

	_, err := ordinaryFill(ctx, trial, remaining)
	restore(remaining, placements)

	if err != nil {
		return 0, err
	}

	return trial.itemsVolume, nil
}

func holdableVolume(box *Container, remaining []*piece) float64 {
	volume := 0.0

	for _, item := range remaining {
		if item != nil && box.couldEverHold(item) {
			volume += item.volume
		}
	}

	return min(volume, box.volume)
}

func emptyKindsByBound(boxes []*Container, remaining []*piece) ([]int, []float64) {
	seen := make(map[boxKind]bool, len(boxes))
	kinds := make([]int, 0, len(boxes))
	bound := make(map[int]float64, len(boxes))

	for i, box := range boxes {
		if box == nil || len(box.items) > 0 || seen[kindOf(box)] {
			continue
		}

		seen[kindOf(box)] = true
		bound[i] = holdableVolume(box, remaining)

		kinds = append(kinds, i)
	}

	slices.SortStableFunc(kinds, func(a, b int) int {
		return cmp.Compare(bound[b], bound[a])
	})

	bounds := make([]float64, 0, len(kinds))
	for _, i := range kinds {
		bounds = append(bounds, bound[i])
	}

	return kinds, bounds
}

func runItemMajor(
	ctx context.Context,
	boxes []*Box,
	items []*piece,
	problem *Problem,
	selection BoxSelection,
) (*Packing, error) {
	sortedBoxes, result := prepareData(boxes, items, problem)
	unpacked := make([]*piece, 0, len(items))
	window := 0

	for _, item := range items {
		err := checkContext(ctx)
		if err != nil {
			return nil, err
		}

		if item == nil {
			continue
		}

		choice, found := selectBox(sortedBoxes[window:], item, selection)
		if !found {
			unpacked = append(unpacked, item)

			continue
		}

		index := window + choice.index
		sortedBoxes[index].place(item, choice.pivot, choice.rotation)

		if selection == SelectNextFit {
			window = index
		}
	}

	unpacked, err := recoverGaps(ctx, sortedBoxes, unpacked, selection, window)
	if err != nil {
		return nil, err
	}

	result.unfit = append(result.unfit, unpacked...)

	return result, nil
}

const gapRecoveryBudget = 24

func recoverGaps(
	ctx context.Context,
	boxes []*Container,
	unpacked []*piece,
	selection BoxSelection,
	window int,
) ([]*piece, error) {
	if len(unpacked) == 0 {
		return unpacked, nil
	}

	start := 0
	if selection == SelectNextFit {
		start = window
	}

	for i := start; i < len(boxes) && len(unpacked) > 0; i++ {
		if boxes[i] == nil || len(boxes[i].items) == 0 {
			continue
		}

		head, tail := unpacked, []*piece(nil)
		if len(head) > gapRecoveryBudget {
			head, tail = unpacked[:gapRecoveryBudget], unpacked[gapRecoveryBudget:]
		}

		recovered, err := fillGaps(ctx, boxes[i], head, true)

		stillUnpacked := make([]*piece, 0, len(recovered)+len(tail))
		stillUnpacked = append(stillUnpacked, recovered...)
		stillUnpacked = append(stillUnpacked, tail...)

		if err != nil {
			return stillUnpacked, err
		}

		unpacked = stillUnpacked
	}

	return unpacked, nil
}

type boxChoice struct {
	index     int
	pivot     Pivot
	rotation  Orientation
	remaining float64
	open      bool
}

func candidateBoxes(boxes []*Container, item *piece) []boxChoice {
	choices := make([]boxChoice, 0, len(boxes))

	for i, box := range boxes {
		if box == nil || !box.canQuota(item) {
			continue
		}

		pivot, rotation, found := bestPlacement(box, item)
		if !found {
			continue
		}

		choices = append(choices, boxChoice{
			index:     i,
			pivot:     pivot,
			rotation:  rotation,
			remaining: box.remainingVolume() - item.volume,
			open:      len(box.items) > 0,
		})
	}

	return choices
}

func openFirst(choices []boxChoice) []boxChoice {
	opened := make([]boxChoice, 0, len(choices))

	for _, choice := range choices {
		if choice.open {
			opened = append(opened, choice)
		}
	}

	if len(opened) > 0 {
		return opened
	}

	return choices
}

func selectBox(boxes []*Container, item *piece, selection BoxSelection) (boxChoice, bool) {
	choices := candidateBoxes(boxes, item)
	if len(choices) == 0 {
		return boxChoice{
			index: -1, pivot: Pivot{}, rotation: OrientationWHD, remaining: 0, open: false,
		}, false
	}

	switch selection {
	case SelectBestFit:
		return rankedByRemaining(openFirst(choices), true, 0), true
	case SelectWorstFit:
		return rankedByRemaining(openFirst(choices), false, 0), true
	case SelectAlmostWorstFit:
		return rankedByRemaining(openFirst(choices), false, 1), true

	case SelectFirstFit, SelectNextFit, SelectFullestBox:
	}

	return choices[0], true
}

func rankedByRemaining(choices []boxChoice, ascending bool, skip int) boxChoice {
	slices.SortStableFunc(choices, func(a, b boxChoice) int {
		if ascending {
			return cmp.Compare(a.remaining, b.remaining)
		}

		return cmp.Compare(b.remaining, a.remaining)
	})

	return choices[min(skip, len(choices)-1)]
}

func fitInSpecificBox(box *Container, item *piece) bool {
	if box == nil || item == nil || !box.canQuota(item) {
		return false
	}

	position, rotation, ok := bestPlacement(box, item)
	if !ok {
		return false
	}

	box.place(item, position, rotation)

	return true
}

func preferredSort(boxes []*Container, items []*piece) []*Container {
	index, found := wholeOrderBox(boxes, items)
	if !found {
		return boxes
	}

	result := make([]*Container, 0, len(boxes))
	result = append(result, boxes[index])

	for j, box := range boxes {
		if j != index {
			result = append(result, box)
		}
	}

	return result
}

type orderBounds struct {
	volume    float64
	weight    float64
	maxLength float64
}

func boundsOf(items []*piece) orderBounds {
	var bounds orderBounds

	for _, item := range items {
		bounds.volume += item.volume
		bounds.weight += item.weight
		bounds.maxLength = max(bounds.maxLength, item.maxLength)
	}

	return bounds
}

func shippableItems(boxes []*Container, items []*piece) []*piece {
	shippable := make([]*piece, 0, len(items))

	for _, item := range items {
		if item != nil && !fitsNoBoxHere(item, boxes) {
			shippable = append(shippable, item)
		}
	}

	return shippable
}

func wholeOrderBox(boxes []*Container, items []*piece) (int, bool) {
	shippable := shippableItems(boxes, items)
	if len(shippable) == 0 {
		return 0, false
	}

	bounds := boundsOf(shippable)

	for i, b := range boxes {
		if b == nil || b.volume < bounds.volume ||
			b.maxWeight < bounds.weight || b.maxLength < bounds.maxLength {
			continue
		}

		if holdsWholeOrder(b, shippable) {
			return i, true
		}
	}

	return 0, false
}

func holdsWholeOrder(b *Container, items []*piece) bool {
	trial := clonePtr(b)
	trial.reset()

	defer restore(items, snapshot(items))

	for _, item := range items {
		if item == nil {
			continue
		}

		if !fitInSpecificBox(trial, item) {
			return false
		}
	}

	return true
}

func packToBox(ctx context.Context, b *Container, items []*piece) ([]*piece, error) {
	if b.rules.Filler != nil {
		return fillThrough(ctx, b.rules.Filler, b, items)
	}

	return ordinaryFill(ctx, b, items)
}

func fillThrough(ctx context.Context, filler Filler, b *Container, items []*piece) ([]*piece, error) {
	byInstance := make(map[Instance]*piece, len(items))
	offered := make([]Instance, 0, len(items))

	for _, item := range items {
		instance := Instance{Item: item.Item, Index: item.index}
		byInstance[instance] = item
		offered = append(offered, instance)
	}

	left, err := filler.Fill(ctx, b, offered)
	if err != nil {
		return items, err
	}

	unpacked := make([]*piece, 0, len(left))

	for _, instance := range left {
		if item, known := byInstance[instance]; known {
			unpacked = append(unpacked, item)
		}
	}

	return unpacked, nil
}

func ordinaryFill(ctx context.Context, b *Container, items []*piece) ([]*piece, error) {
	unpacked, err := fillBox(ctx, b, items)
	if err != nil || len(unpacked) == 0 || len(b.items) == 0 {
		return unpacked, err
	}

	best, err := bestSeededFill(ctx, b, items)
	if err != nil {
		return unpacked, err
	}

	if best != nil && best.volume > b.itemsVolume {
		unpacked = best.applyTo(b, items)
	}

	return fillGaps(ctx, b, unpacked, false)
}

func fillGaps(ctx context.Context, b *Container, items []*piece, repack bool) ([]*piece, error) {
	remaining := items

	for len(remaining) > 0 {
		next := make([]*piece, 0, len(remaining))
		progress := false

		for i, item := range remaining {
			err := checkContext(ctx)
			if err != nil {
				return appendRest(next, remaining, i), err
			}

			if packItem(b, item, repack) {
				progress = true

				continue
			}

			next = append(next, item)
		}

		remaining = next

		if !progress {
			break
		}
	}

	return remaining, nil
}

func fillBox(ctx context.Context, b *Container, items []*piece) ([]*piece, error) {
	return fillBoxWith(ctx, b, items, true)
}

func fillBoxWith(ctx context.Context, b *Container, items []*piece, repack bool) ([]*piece, error) {
	unpacked := make([]*piece, 0, len(items))

	for i := range items {
		err := checkContext(ctx)
		if err != nil {
			return appendRest(unpacked, items, i), err
		}

		if !packItem(b, items[i], repack) {
			unpacked = append(unpacked, items[i])
		}
	}

	return unpacked, nil
}

type seededFill struct {
	placed   []bool
	position []Pivot
	rotation []Orientation
	volume   float64
}

func (f *seededFill) applyTo(b *Container, items []*piece) []*piece {
	b.reset()

	unpacked := make([]*piece, 0, len(items))

	for i, item := range items {
		if !f.placed[i] {
			unpacked = append(unpacked, item)

			continue
		}

		b.place(item, f.position[i], f.rotation[i])
	}

	return unpacked
}

func bestSeededFill(ctx context.Context, b *Container, items []*piece) (*seededFill, error) {
	var best *seededFill

	for _, rotation := range items[0].rotations {
		trial, err := seededFillAttempt(ctx, b, items, rotation)
		if err != nil {
			return nil, err
		}

		if trial != nil && (best == nil || trial.volume > best.volume) {
			best = trial
		}
	}

	return best, nil
}

func seededFillAttempt(ctx context.Context, b *Container, items []*piece, seed Orientation) (*seededFill, error) {
	trialBox := clonePtr(b)
	trialBox.reset()

	trialItems := cloneSlice(items)

	first := trialItems[0]
	allowUnstable := !trialBox.hasStableOrientation(first)

	if !trialBox.canQuota(first) ||
		!trialBox.acceptsPlacement(first, Pivot{}, rotatedDimension(first.Item, seed), allowUnstable) {
		return nil, nil //nolint:nilnil
	}

	trialBox.place(trialItems[0], Pivot{}, seed)

	_, err := fillBoxWith(ctx, trialBox, trialItems[1:], false)
	if err != nil {
		return nil, err
	}

	fill := &seededFill{
		placed:   make([]bool, len(items)),
		position: make([]Pivot, len(items)),
		rotation: make([]Orientation, len(items)),
		volume:   trialBox.itemsVolume,
	}

	inBox := make(map[*piece]bool, len(trialBox.items))
	for _, item := range trialBox.items {
		inBox[item] = true
	}

	for i, item := range trialItems {
		if inBox[item] {
			fill.placed[i] = true
			fill.position[i] = item.position
			fill.rotation[i] = item.orientation
		}
	}

	return fill, nil
}

func appendRest(unpacked []*piece, items []*piece, startIndex int) []*piece {
	for j := startIndex; j < len(items); j++ {
		if items[j] != nil {
			unpacked = append(unpacked, items[j])
		}
	}

	return unpacked
}

func packItem(b *Container, item *piece, repack bool) bool {
	if item == nil {
		return false
	}

	if fitInSpecificBox(b, item) {
		return true
	}

	if repack && b.canQuota(item) && len(b.items) > 0 {
		return attemptRepack(b, item)
	}

	return false
}

type placement struct {
	position Pivot
	rotation Orientation
}

func snapshot(items []*piece) []placement {
	saved := make([]placement, len(items))

	for i, item := range items {
		saved[i] = placement{position: item.position, rotation: item.orientation}
	}

	return saved
}

func restore(items []*piece, saved []placement) {
	for i, item := range items {
		item.position = saved[i].position
		item.orientation = saved[i].rotation
	}
}

func attemptRepack(b *Container, newItem *piece) bool {
	state := saveContainer(b)
	newItemPlacement := snapshot([]*piece{newItem})

	b.reset()

	if !fitInSpecificBox(b, newItem) {
		restore([]*piece{newItem}, newItemPlacement)
		state.rollback()

		return false
	}

	for _, item := range state.items {
		if fitInSpecificBox(b, item) {
			continue
		}

		restore([]*piece{newItem}, newItemPlacement)
		state.rollback()

		return false
	}

	return true
}

func fitsNoBoxHere(item *piece, boxes []*Container) bool {
	for _, box := range boxes {
		if box != nil && box.couldEverHold(item) {
			return false
		}
	}

	return true
}
