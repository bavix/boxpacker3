package boxpacker3

// Item represents an item that can be packed into a box.
type Item struct {
	id     string
	whd    [3]float64 // [width, height, depth] - stored as array for efficient matrix rotation
	weight float64
	volume float64

	maxLength    float64
	rotationType RotationType
	position     Pivot
}

type itemSlice []*Item

func (it itemSlice) Len() int {
	return len(it)
}

func (it itemSlice) Less(i, j int) bool {
	if it[i] == nil {
		return false
	}

	if it[j] == nil {
		return true
	}

	return it[i].volume < it[j].volume
}

func (it itemSlice) Swap(i, j int) {
	it[i], it[j] = it[j], it[i]
}

// NewItem creates a new item with the given parameters.
func NewItem(id string, w, h, d, wg float64) *Item {
	//nolint:exhaustruct
	return &Item{
		id:        id,
		whd:       [3]float64{w, h, d},
		weight:    wg,
		volume:    w * h * d,
		maxLength: max(w, h, d),
	}
}

// NewItem2D creates a new 2D item with the given parameters.
// The depth is set to 1, making it effectively 2D (width x height).
// This is useful for packing flat items like sheets, boards, or panels.
func NewItem2D(id string, w, h, wg float64) *Item {
	return NewItem(id, w, h, 1, wg)
}

func (i *Item) GetID() string {
	return i.id
}

func (i *Item) GetWidth() float64 {
	return i.whd[0]
}

func (i *Item) GetHeight() float64 {
	return i.whd[1]
}

func (i *Item) GetDepth() float64 {
	return i.whd[2]
}

func (i *Item) GetVolume() float64 {
	return i.volume
}

func (i *Item) GetWeight() float64 {
	return i.weight
}

func (i *Item) GetPosition() Pivot {
	return i.position
}

func (i *Item) setRotationType(rt RotationType) {
	i.rotationType = rt
}

func (i *Item) GetDimension() Dimension {
	matrix := rotationMatrix[i.rotationType]

	return Dimension{
		i.whd[matrix[0]],
		i.whd[matrix[1]],
		i.whd[matrix[2]],
	}
}

// Intersect tests for intersections between two items.
func (i *Item) Intersect(it *Item) bool {
	if i == nil || it == nil {
		return false
	}

	return i.intersect(it, WidthAxis, HeightAxis) &&
		i.intersect(it, HeightAxis, DepthAxis) &&
		i.intersect(it, WidthAxis, DepthAxis)
}

func (i *Item) intersect(it *Item, x, y Axis) bool {
	matrix1 := rotationMatrix[i.rotationType]
	matrix2 := rotationMatrix[it.rotationType]

	d1x := i.whd[matrix1[x]]

	d1y := i.whd[matrix1[y]]

	d2x := it.whd[matrix2[x]]

	d2y := it.whd[matrix2[y]]

	const minDimension = 1e-10

	if d1x <= minDimension || d1y <= minDimension || d2x <= minDimension || d2y <= minDimension {
		return false
	}

	cx1 := i.position[x] + d1x/2  //nolint:mnd
	cy1 := i.position[y] + d1y/2  //nolint:mnd
	cx2 := it.position[x] + d2x/2 //nolint:mnd
	cy2 := it.position[y] + d2y/2 //nolint:mnd

	ix := max(cx1, cx2) - min(cx1, cx2)
	iy := max(cy1, cy2) - min(cy1, cy2)

	return ix < (d1x+d2x)/2 && iy < (d1y+d2y)/2
}
