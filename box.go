package boxpacker3

import (
	"golang.org/x/exp/slices"
)

type Box struct {
	id        string
	width     float64
	height    float64
	depth     float64
	maxWeight float64
	volume    float64
	items     []*Item

	maxLength   float64
	itemsVolume float64
	itemsWeight float64
}

type boxSlice []*Box

func (bs boxSlice) Len() int {
	return len(bs)
}

func (bs boxSlice) Less(i, j int) bool {
	return bs[i].volume < bs[j].volume
}

func (bs boxSlice) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func NewBox(id string, w, h, d, mw float64) *Box {
	//nolint:exhaustruct
	return &Box{
		id:        id,
		width:     w,
		height:    h,
		depth:     d,
		maxWeight: mw,
		maxLength: slices.Max([]float64{w, h, d}),
		volume:    w * h * d,
		items:     nil,
	}
}

func (b *Box) GetID() string {
	return b.id
}

func (b *Box) GetWidth() float64 {
	return b.width
}

func (b *Box) GetHeight() float64 {
	return b.height
}

func (b *Box) GetDepth() float64 {
	return b.depth
}

func (b *Box) GetVolume() float64 {
	return b.volume
}

func (b *Box) GetMaxWeight() float64 {
	return b.maxWeight
}

func (b *Box) GetItems() []*Item {
	return b.items
}

// PutItem Attempts to place an element at anchor point p of box b.
func (b *Box) PutItem(item *Item, p Pivot) bool {
	if !b.canQuota(item) {
		return false
	}

	item.position = p

	for rt := RotationTypeWhd; rt <= RotationTypeWdh; rt++ {
		item.rotationType = rt
		d := item.GetDimension()

		if b.width < p[WidthAxis]+d[WidthAxis] || b.height < p[HeightAxis]+d[HeightAxis] || b.depth < p[DepthAxis]+d[DepthAxis] {
			continue
		}

		for _, ib := range b.items {
			if ib.Intersect(item) {
				return false
			}
		}

		b.insert(item)

		return true
	}

	return false
}

func (b *Box) canQuota(item *Item) bool {
	return b.itemsVolume+item.volume <= b.volume && b.itemsWeight+item.weight <= b.maxWeight
}

func (b *Box) insert(item *Item) {
	b.items = append(b.items, item)
	b.itemsVolume += item.volume
	b.itemsWeight += item.weight
}

func (b *Box) purge() {
	b.items = make([]*Item, 0, len(b.items))
	b.itemsVolume = 0
	b.itemsWeight = 0
}
