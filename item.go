package boxpacker3

import (
	"math"
)

type Item struct {
	ID     string
	Width  float64
	Height float64
	Depth  float64
	Weight float64
	Volume float64

	RotationType RotationType
	Position     Pivot
}

type ItemSlice []*Item

func (is ItemSlice) Len() int {
	return len(is)
}

func (is ItemSlice) Less(i, j int) bool {
	return is[i].GetVolume() < is[j].GetVolume()
}

func (is ItemSlice) Swap(i, j int) {
	is[i], is[j] = is[j], is[i]
}

func NewItem(id string, w, h, d, wg float64) *Item {
	return &Item{
		ID:     id,
		Width:  w,
		Height: h,
		Depth:  d,
		Weight: wg,
		Volume: w * h * d,
	}
}

func (i *Item) GetID() string {
	return i.ID
}

func (i *Item) GetWidth() float64 {
	return i.Width
}

func (i *Item) GetHeight() float64 {
	return i.Height
}

func (i *Item) GetDepth() float64 {
	return i.Depth
}

func (i *Item) GetVolume() float64 {
	return i.Volume
}

func (i *Item) GetWeight() float64 {
	return i.Weight
}

//nolint:nonamedreturns
func (i *Item) GetDimension() (d Dimension) {
	switch i.RotationType {
	case RotationTypeWhd:
		d = Dimension{i.GetWidth(), i.GetHeight(), i.GetDepth()}
	case RotationTypeHwd:
		d = Dimension{i.GetHeight(), i.GetWidth(), i.GetDepth()}
	case RotationTypeHdw:
		d = Dimension{i.GetHeight(), i.GetDepth(), i.GetWidth()}
	case RotationTypeDhw:
		d = Dimension{i.GetDepth(), i.GetHeight(), i.GetWidth()}
	case RotationTypeDwh:
		d = Dimension{i.GetDepth(), i.GetWidth(), i.GetHeight()}
	case RotationTypeWdh:
		d = Dimension{i.GetWidth(), i.GetDepth(), i.GetHeight()}
	}

	return
}

// Intersect Проверяет пересечения между элементом i и элементом it.
func (i *Item) Intersect(it *Item) bool {
	d1 := i.GetDimension()
	d2 := it.GetDimension()

	return rectIntersect(d1, d2, i, it, WidthAxis, HeightAxis) &&
		rectIntersect(d1, d2, i, it, HeightAxis, DepthAxis) &&
		rectIntersect(d1, d2, i, it, WidthAxis, DepthAxis)
}

// rectIntersect Проверяет пересекаются ли два прямоугольника от осей x и y элементов i1 и i2.
func rectIntersect(d1, d2 Dimension, i1, i2 *Item, x, y Axis) bool {
	cx1 := i1.Position[x] + d1[x]/2 //nolint:gomnd
	cy1 := i1.Position[y] + d1[y]/2 //nolint:gomnd
	cx2 := i2.Position[x] + d2[x]/2 //nolint:gomnd
	cy2 := i2.Position[y] + d2[y]/2 //nolint:gomnd

	ix := math.Max(cx1, cx2) - math.Min(cx1, cx2)
	iy := math.Max(cy1, cy2) - math.Min(cy1, cy2)

	return ix < (d1[x]+d2[x])/2 && iy < (d1[y]+d2[y])/2
}
