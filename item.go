package boxpacker3

import (
	"math"
	"slices"
)

type Item struct {
	id     string
	width  float64
	height float64
	depth  float64
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
	return it[i].volume < it[j].volume
}

func (it itemSlice) Swap(i, j int) {
	it[i], it[j] = it[j], it[i]
}

func NewItem(id string, w, h, d, wg float64) *Item {
	//nolint:exhaustruct
	return &Item{
		id:        id,
		width:     w,
		height:    h,
		depth:     d,
		weight:    wg,
		volume:    w * h * d,
		maxLength: slices.Max([]float64{w, h, d}),
	}
}

func (i *Item) GetID() string {
	return i.id
}

func (i *Item) GetWidth() float64 {
	return i.width
}

func (i *Item) GetHeight() float64 {
	return i.height
}

func (i *Item) GetDepth() float64 {
	return i.depth
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

func (i *Item) GetDimension() Dimension {
	switch i.rotationType {
	case RotationTypeWhd:
		return Dimension{i.GetWidth(), i.GetHeight(), i.GetDepth()}
	case RotationTypeHwd:
		return Dimension{i.GetHeight(), i.GetWidth(), i.GetDepth()}
	case RotationTypeHdw:
		return Dimension{i.GetHeight(), i.GetDepth(), i.GetWidth()}
	case RotationTypeDhw:
		return Dimension{i.GetDepth(), i.GetHeight(), i.GetWidth()}
	case RotationTypeDwh:
		return Dimension{i.GetDepth(), i.GetWidth(), i.GetHeight()}
	case RotationTypeWdh:
		return Dimension{i.GetWidth(), i.GetDepth(), i.GetHeight()}
	default: // RotationTypeWhd
		return Dimension{i.GetWidth(), i.GetHeight(), i.GetDepth()}
	}
}

// Intersect Tests for intersections between the i element and the it element.
func (i *Item) Intersect(it *Item) bool {
	d1 := i.GetDimension()
	d2 := it.GetDimension()

	return i.intersect(d1, d2, it, WidthAxis, HeightAxis) &&
		i.intersect(d1, d2, it, HeightAxis, DepthAxis) &&
		i.intersect(d1, d2, it, WidthAxis, DepthAxis)
}

// intersect Checks if two rectangles intersect from the x and y axes of elements i1 and i2.
func (i *Item) intersect(d1, d2 Dimension, it *Item, x, y Axis) bool {
	cx1 := i.position[x] + d1[x]/2  //nolint:mnd
	cy1 := i.position[y] + d1[y]/2  //nolint:mnd
	cx2 := it.position[x] + d2[x]/2 //nolint:mnd
	cy2 := it.position[y] + d2[y]/2 //nolint:mnd

	ix := math.Max(cx1, cx2) - math.Min(cx1, cx2)
	iy := math.Max(cy1, cy2) - math.Min(cy1, cy2)

	return ix < (d1[x]+d2[x])/2 && iy < (d1[y]+d2[y])/2
}
