package boxpacker3

import (
	"math"
	"slices"
)

func allAxes() [3]Axis {
	return [3]Axis{WidthAxis, HeightAxis, DepthAxis}
}

func (b *Box) boxDimension(axis Axis) float64 {
	switch axis {
	case WidthAxis:
		return b.width
	case HeightAxis:
		return b.height
	case DepthAxis:
		return b.depth
	}

	return 0
}

func (c *Container) resetPoints() {
	c.points = append(c.points[:0], Pivot{})
}

func (c *Container) refreshPoints(placed *piece) {
	position := placed.position
	dimension := placed.dimension()

	c.points = slices.DeleteFunc(c.points, func(p Pivot) bool {
		return pointInside(p, position, dimension)
	})

	for _, axis := range allAxes() {
		corner := position
		corner[axis] += dimension[axis]

		if corner[axis] >= c.boxDimension(axis) {
			continue
		}

		c.addPoint(corner)

		for _, other := range allAxes() {
			if other == axis {
				continue
			}

			projected := corner
			projected[other] = c.project(corner, other)

			c.addPoint(projected)
		}
	}

	slices.SortFunc(c.points, comparePoints)
}

func (c *Container) addPoint(p Pivot) {
	for _, axis := range allAxes() {
		if p[axis] < 0 || p[axis] >= c.boxDimension(axis) {
			return
		}
	}

	for _, item := range c.items {
		if pointInside(p, item.position, item.dimension()) {
			return
		}
	}

	if slices.Contains(c.points, p) {
		return
	}

	c.points = append(c.points, p)
}

func (c *Container) project(point Pivot, axis Axis) float64 {
	first, second := otherAxes(axis)
	best := 0.0

	for _, item := range c.items {
		position := item.position
		dimension := item.dimension()

		if !spans(position[first], dimension[first], point[first]) ||
			!spans(position[second], dimension[second], point[second]) {
			continue
		}

		face := position[axis] + dimension[axis]
		if face <= point[axis] && face > best {
			best = face
		}
	}

	return best
}

func otherAxes(axis Axis) (Axis, Axis) {
	switch axis {
	case WidthAxis:
		return HeightAxis, DepthAxis
	case HeightAxis:
		return WidthAxis, DepthAxis
	case DepthAxis:
		return WidthAxis, HeightAxis
	}

	return WidthAxis, HeightAxis
}

func spans(start, size, value float64) bool {
	return start <= value && value < start+size
}

func pointInside(p Pivot, position Pivot, dimension Dimension) bool {
	for _, axis := range allAxes() {
		if !spans(position[axis], dimension[axis], p[axis]) {
			return false
		}
	}

	return true
}

func comparePoints(a, b Pivot) int {
	for _, axis := range [3]Axis{DepthAxis, HeightAxis, WidthAxis} {
		if diff := a[axis] - b[axis]; diff != 0 {
			if diff < 0 {
				return -1
			}

			return 1
		}
	}

	return 0
}

func overlaps(aPosition Pivot, aDimension Dimension, bPosition Pivot, bDimension Dimension) bool {
	for _, axis := range allAxes() {
		if aDimension[axis] <= minDimension || bDimension[axis] <= minDimension {
			return false
		}

		low := max(aPosition[axis], bPosition[axis])
		high := min(aPosition[axis]+aDimension[axis], bPosition[axis]+bDimension[axis])
		scale := max(aDimension[axis], bDimension[axis])

		if high-low <= scale*dimensionEpsilon {
			return false
		}
	}

	return true
}

func overlapLength(aStart, aSize, bStart, bSize float64) float64 {
	return max(0, min(aStart+aSize, bStart+bSize)-max(aStart, bStart))
}

func (c *Container) fitsAt(position Pivot, dimension Dimension) bool {
	for _, axis := range allAxes() {
		if position[axis] < 0 || !withinLimit(position[axis]+dimension[axis], c.boxDimension(axis)) {
			return false
		}
	}

	for _, item := range c.items {
		if overlaps(position, dimension, item.position, item.dimension()) {
			return false
		}
	}

	return true
}

func (c *Container) bearsTheLoad(item *piece, position Pivot, dimension Dimension) bool {
	if item == nil {
		return true
	}

	if !c.loadLimited && !item.nothingOnTop && item.maxLoadOnTop <= 0 {
		return true
	}

	if !c.carriesWhatIsAbove(item, position, dimension) {
		return false
	}

	for _, below := range c.items {
		if below == nil || !underneath(position, dimension, below.position, below.dimension()) {
			continue
		}

		if !c.takesOneMore(below, item.weight) {
			return false
		}
	}

	return true
}

func (c *Container) takesOneMore(below *piece, weight float64) bool {
	if below.nothingOnTop {
		return false
	}

	if below.maxLoadOnTop <= 0 {
		return true
	}

	return c.loadAbove(below.position, below.dimension(), below)+weight <= below.maxLoadOnTop
}

func (c *Container) carriesWhatIsAbove(item *piece, position Pivot, dimension Dimension) bool {
	if !item.nothingOnTop && item.maxLoadOnTop <= 0 {
		return true
	}

	if item.nothingOnTop {
		return !c.anythingAbove(position, dimension)
	}

	return c.loadAbove(position, dimension, nil) <= item.maxLoadOnTop
}

func (c *Container) anythingAbove(position Pivot, dimension Dimension) bool {
	for _, above := range c.items {
		if above != nil && underneath(above.position, above.dimension(), position, dimension) {
			return true
		}
	}

	return false
}

func underneath(position Pivot, dimension Dimension, other Pivot, otherDimension Dimension) bool {
	top := other[DepthAxis] + otherDimension[DepthAxis]
	if top > position[DepthAxis]+max(top, position[DepthAxis])*dimensionEpsilon {
		return false
	}

	return sharesGround(position, dimension, other, otherDimension)
}

func sharesGround(position Pivot, dimension Dimension, other Pivot, otherDimension Dimension) bool {
	return overlapLength(position[WidthAxis], dimension[WidthAxis],
		other[WidthAxis], otherDimension[WidthAxis]) > 0 &&
		overlapLength(position[HeightAxis], dimension[HeightAxis],
			other[HeightAxis], otherDimension[HeightAxis]) > 0
}

func (c *Container) loadAbove(position Pivot, dimension Dimension, skip *piece) float64 {
	load := 0.0

	for _, above := range c.items {
		if above == nil || above == skip {
			continue
		}

		if !underneath(above.position, above.dimension(), position, dimension) {
			continue
		}

		load += above.weight
	}

	return load
}

func (c *Container) contactArea(position Pivot, dimension Dimension) float64 {
	area := 0.0

	for _, axis := range allAxes() {
		first, second := otherAxes(axis)
		face := dimension[first] * dimension[second]

		if position[axis] <= c.boxDimension(axis)*dimensionEpsilon {
			area += face
		}

		far := position[axis] + dimension[axis]
		if math.Abs(far-c.boxDimension(axis)) <= c.boxDimension(axis)*dimensionEpsilon {
			area += face
		}
	}

	for _, item := range c.items {
		other := item.position
		otherDimension := item.dimension()

		if apart(position, dimension, other, otherDimension) {
			continue
		}

		area += touching(position, dimension, other, otherDimension)
	}

	return area
}

func apart(position Pivot, dimension Dimension, other Pivot, otherDimension Dimension) bool {
	for _, axis := range allAxes() {
		slack := dimensionEpsilon * max(dimension[axis], otherDimension[axis])

		if position[axis]-slack > other[axis]+otherDimension[axis] ||
			other[axis]-slack > position[axis]+dimension[axis] {
			return true
		}
	}

	return false
}

func wholeSurface(dimension Dimension) float64 {
	const faces = 2

	return faces * (dimension[WidthAxis]*dimension[HeightAxis] +
		dimension[WidthAxis]*dimension[DepthAxis] +
		dimension[HeightAxis]*dimension[DepthAxis])
}

func touching(position Pivot, dimension Dimension, other Pivot, otherDimension Dimension) float64 {
	for _, axis := range allAxes() {
		near := math.Abs(position[axis] + dimension[axis] - other[axis])
		far := math.Abs(other[axis] + otherDimension[axis] - position[axis])

		if near > dimensionEpsilon*max(dimension[axis], otherDimension[axis]) &&
			far > dimensionEpsilon*max(dimension[axis], otherDimension[axis]) {
			continue
		}

		first, second := otherAxes(axis)

		return overlapLength(position[first], dimension[first], other[first], otherDimension[first]) *
			overlapLength(position[second], dimension[second], other[second], otherDimension[second])
	}

	return 0
}

func (c *Container) residualGap(position Pivot, dimension Dimension, axis Axis) float64 {
	first, second := otherAxes(axis)
	start := position[axis] + dimension[axis]
	nearest := c.boxDimension(axis)

	for _, item := range c.items {
		itemPosition := item.position
		itemDimension := item.dimension()

		if itemPosition[axis] < start-max(start, itemPosition[axis])*dimensionEpsilon {
			continue
		}

		if overlapLength(position[first], dimension[first], itemPosition[first], itemDimension[first]) <= 0 ||
			overlapLength(position[second], dimension[second], itemPosition[second], itemDimension[second]) <= 0 {
			continue
		}

		if itemPosition[axis] < nearest {
			nearest = itemPosition[axis]
		}
	}

	return nearest - start
}

const stableTiltRadians = 0.261

func (c *Container) supportRatio(position Pivot, dimension Dimension) float64 {
	base := dimension[WidthAxis] * dimension[HeightAxis]
	if base <= minDimension {
		return 1
	}

	if position[DepthAxis] <= c.depth*dimensionEpsilon {
		return 1
	}

	supported := 0.0

	for _, item := range c.items {
		itemPosition := item.position
		itemDimension := item.dimension()
		top := itemPosition[DepthAxis] + itemDimension[DepthAxis]

		if math.Abs(top-position[DepthAxis]) > max(top, position[DepthAxis])*dimensionEpsilon {
			continue
		}

		supported += overlapLength(position[WidthAxis], dimension[WidthAxis], itemPosition[WidthAxis], itemDimension[WidthAxis]) *
			overlapLength(position[HeightAxis], dimension[HeightAxis], itemPosition[HeightAxis], itemDimension[HeightAxis])
	}

	return min(supported/base, 1)
}

func (b *Box) isStableOrientation(dimension Dimension) bool {
	depth := dimension[DepthAxis]
	if depth <= minDimension {
		return true
	}

	if math.Abs(depth-b.depth) <= b.depth*dimensionEpsilon {
		return true
	}

	return math.Atan(min(dimension[WidthAxis], dimension[HeightAxis])/depth) > stableTiltRadians
}

func (b *Box) hasStableOrientation(item *piece) bool {
	for _, rotation := range item.rotations {
		dimension := rotatedDimension(item.Item, rotation)

		if !b.dimensionsFit(dimension) {
			continue
		}

		if b.isStableOrientation(dimension) {
			return true
		}
	}

	return false
}

func (b *Box) dimensionsFit(dimension Dimension) bool {
	for _, axis := range allAxes() {
		if !withinLimit(dimension[axis], b.boxDimension(axis)) {
			return false
		}
	}

	return true
}

func (b *Box) couldEverHold(item *piece) bool {
	if !withinLimit(b.emptyWeight+item.weight, b.maxWeight) {
		return false
	}

	if len(b.accepts) > 0 && !slices.Contains(b.accepts, item.class) {
		return false
	}

	for _, rotation := range item.rotations {
		if b.dimensionsFit(rotatedDimension(item.Item, rotation)) {
			return true
		}
	}

	return false
}

func fitsNoBox(item *piece, boxes []*Box) bool {
	for _, box := range boxes {
		if box != nil && box.couldEverHold(item) {
			return false
		}
	}

	return true
}

func (b *Box) fitsFreelyTurned(item *piece) bool {
	all := []Orientation{
		OrientationWHD, OrientationHWD, OrientationHDW,
		OrientationDHW, OrientationDWH, OrientationWDH,
	}

	for _, orientation := range all {
		if b.dimensionsFit(rotatedDimension(item.Item, orientation)) {
			return true
		}
	}

	return false
}
