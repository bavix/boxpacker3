package boxpacker3

import "slices"

type Container struct {
	*Box

	index   int
	problem *Problem

	items       []*piece
	itemsVolume float64
	itemsWeight float64

	loadLimited bool

	separated bool

	points []Pivot
	spaces []space
	grid   *itemGrid
	rules  Rules
}

func newContainer(box *Box, index int, rules Rules, problem *Problem) *Container {
	return &Container{
		Box:         box,
		index:       index,
		problem:     problem,
		items:       make([]*piece, 0, 1),
		itemsVolume: 0,
		itemsWeight: 0,
		loadLimited: false,
		separated:   false,
		points:      []Pivot{{}},
		spaces:      []space{{origin: Pivot{}, size: Dimension{box.width, box.height, box.depth}}},
		rules:       rules,
	}
}

func (c *Container) Index() int {
	return c.index
}

func (c *Container) putItem(item *piece, p Pivot) bool {
	if item == nil || !c.canQuota(item) {
		return false
	}

	allowUnstable := !c.hasStableOrientation(item)

	for _, rotation := range item.rotations {
		if !c.acceptsPlacement(item, p, rotatedDimension(item.Item, rotation), allowUnstable) {
			continue
		}

		c.place(item, p, rotation)

		return true
	}

	return false
}

func (c *Container) canQuota(item *piece) bool {
	if item == nil {
		return false
	}

	return c.canFitVolume(item) && c.canFitWeight(item) && c.takesTheClass(item) && c.rules.admits(c, item)
}

func (c *Container) takesTheClass(item *piece) bool {
	if len(c.accepts) > 0 && !slices.Contains(c.accepts, item.class) {
		return false
	}

	if !c.separated && len(item.separateFrom) == 0 && len(c.rules.Pair) == 0 {
		return true
	}

	for _, packed := range c.items {
		if !item.travelsWith(packed.Item) || !c.rules.together(item, packed) {
			return false
		}
	}

	return true
}

func (c *Container) canFitVolume(item *piece) bool {
	if item == nil {
		return false
	}

	return withinLimit(c.itemsVolume+item.volume, c.volume)
}

func (c *Container) canFitWeight(item *piece) bool {
	if item == nil {
		return false
	}

	return withinLimit(c.emptyWeight+c.itemsWeight+item.weight, c.maxWeight)
}

//nolint:ireturn // the copy helper takes a cloner.
func (c *Container) clone() cloner {
	copied := &Container{
		Box:         c.Box,
		index:       c.index,
		problem:     c.problem,
		items:       nil,
		itemsVolume: c.itemsVolume,
		itemsWeight: c.itemsWeight,
		loadLimited: c.loadLimited,
		separated:   c.separated,
		points:      slices.Clone(c.points),
		spaces:      slices.Clone(c.spaces),
		grid:        nil,
		rules:       c.rules,
	}

	if c.items != nil {
		copied.items = make([]*piece, len(c.items), cap(c.items))
		copy(copied.items, c.items)
	}

	return copied
}

func (c *Container) insert(item *piece) {
	c.invalidateIndex()

	c.items = append(c.items, item)
	c.itemsVolume += item.volume
	c.itemsWeight += item.weight
	c.loadLimited = c.loadLimited || item.nothingOnTop || item.maxLoadOnTop > 0
	c.separated = c.separated || len(item.separateFrom) > 0
	c.refreshPoints(item)
	c.refreshSpaces(item)
}

func (c *Container) reset() {
	c.invalidateIndex()

	c.items = c.items[:0]
	c.itemsVolume = 0
	c.itemsWeight = 0
	c.loadLimited = false
	c.separated = false
	c.resetPoints()
	c.resetSpaces()
}

func (c *Container) remainingVolume() float64 {
	return c.volume - c.itemsVolume
}

func (c *Container) grossWeight() float64 {
	return c.emptyWeight + c.itemsWeight
}

func (c *Container) packed() PackedBox {
	items := make([]PackedItem, 0, len(c.items))
	for _, item := range c.items {
		items = append(items, item.packed())
	}

	return PackedBox{Box: c.Box, Index: c.index, Items: items, Stats: c.stats()}
}

func (c *Container) stats() BoxStats {
	var envelope Dimension

	for _, item := range c.items {
		dimension := item.dimension()

		for _, axis := range allAxes() {
			envelope[axis] = max(envelope[axis], item.position[axis]+dimension[axis])
		}
	}

	fill := 0.0
	if c.volume > 0 {
		fill = c.itemsVolume / c.volume
	}

	return BoxStats{
		ItemsVolume: c.itemsVolume,
		ItemsWeight: c.itemsWeight,
		GrossWeight: c.grossWeight(),
		Fill:        fill,
		Envelope:    envelope,
	}
}

func (c *Container) Contents() []PackedItem {
	items := make([]PackedItem, 0, len(c.items))
	for _, item := range c.items {
		items = append(items, item.packed())
	}

	return items
}

type containerState struct {
	container  *Container
	saved      *Container
	items      []*piece
	placements []placement
}

func saveContainer(c *Container) containerState {
	items := slices.Clone(c.items)

	return containerState{container: c, saved: clonePtr(c), items: items, placements: snapshot(items)}
}

func (s containerState) rollback() {
	restore(s.items, s.placements)
	*s.container = *s.saved
}
