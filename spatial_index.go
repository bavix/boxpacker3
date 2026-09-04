package boxpacker3

const (
	gridAxisCells = 4
	gridCells     = gridAxisCells * gridAxisCells * gridAxisCells
)

type itemGrid struct {
	cell    Dimension
	start   []int
	entries []int
	stamp   []uint32
	filled  []int
	epoch   uint32
	covered int
}

func (c *Container) nearby(position Pivot, dimension Dimension, visit func(*piece) bool) {
	if len(c.items) == 0 {
		return
	}

	if len(c.items) < gridAxisCells {
		for _, item := range c.items {
			if !visit(item) {
				return
			}
		}

		return
	}

	grid := c.indexed()
	low, high := grid.cellRange(position, dimension)

	grid.turn()
	grid.walk(low, high, c.items, visit)
}

func (g *itemGrid) turn() {
	if g.epoch == ^uint32(0) {
		clear(g.stamp)

		g.epoch = 0
	}

	g.epoch++
}

func (g *itemGrid) walk(low, high [3]int, items []*piece, visit func(*piece) bool) {
	stop := false

	g.eachCell(low, high, func(cell int) {
		if stop {
			return
		}

		for _, index := range g.entries[g.start[cell]:g.start[cell+1]] {
			if g.stamp[index] == g.epoch {
				continue
			}

			g.stamp[index] = g.epoch

			if !visit(items[index]) {
				stop = true

				return
			}
		}
	})
}

func (c *Container) invalidateIndex() {
	if c.grid != nil {
		c.grid.covered = -1
	}
}

func (c *Container) indexed() *itemGrid {
	if c.grid == nil {
		c.grid = &itemGrid{
			cell:    Dimension{},
			start:   make([]int, gridCells+1),
			entries: nil,
			stamp:   nil,
			filled:  make([]int, gridCells),
			epoch:   0,
			covered: -1,
		}
	}

	if c.grid.covered != len(c.items) {
		c.grid.rebuild(c)
	}

	return c.grid
}

func (g *itemGrid) rebuild(c *Container) {
	for _, axis := range allAxes() {
		g.cell[axis] = max(c.boxDimension(axis), minDimension) / gridAxisCells
	}

	clear(g.start)

	for _, item := range c.items {
		low, high := g.cellRange(item.position, item.dimension())

		g.eachCell(low, high, func(cell int) {
			g.start[cell+1]++
		})
	}

	for cell := range gridCells {
		g.start[cell+1] += g.start[cell]
	}

	total := g.start[gridCells]
	if cap(g.entries) < total {
		g.entries = make([]int, total)
	}

	g.entries = g.entries[:total]

	filled := g.filled
	clear(filled)

	for index, item := range c.items {
		low, high := g.cellRange(item.position, item.dimension())

		g.eachCell(low, high, func(cell int) {
			g.entries[g.start[cell]+filled[cell]] = index
			filled[cell]++
		})
	}

	if cap(g.stamp) < len(c.items) {
		g.stamp = make([]uint32, len(c.items))
	}

	g.stamp = g.stamp[:len(c.items)]
	clear(g.stamp)

	g.epoch = 0
	g.covered = len(c.items)
}

func (g *itemGrid) eachCell(low, high [3]int, visit func(int)) {
	for width := low[WidthAxis]; width <= high[WidthAxis]; width++ {
		for height := low[HeightAxis]; height <= high[HeightAxis]; height++ {
			for depth := low[DepthAxis]; depth <= high[DepthAxis]; depth++ {
				visit(cellNumber(width, height, depth))
			}
		}
	}
}

func (g *itemGrid) cellRange(position Pivot, dimension Dimension) ([3]int, [3]int) {
	var low, high [3]int

	for _, axis := range allAxes() {
		low[axis] = clampCell(int(position[axis] / g.cell[axis]))
		high[axis] = clampCell(int((position[axis] + dimension[axis]) / g.cell[axis]))
	}

	return low, high
}

func clampCell(at int) int {
	return min(max(at, 0), gridAxisCells-1)
}

func cellNumber(width, height, depth int) int {
	return (width*gridAxisCells+height)*gridAxisCells + depth
}
