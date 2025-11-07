package boxpacker3

// Box represents a box that can hold items.
type Box struct {
	id string

	width  float64
	height float64
	depth  float64

	maxWeight float64
	volume    float64

	items []*Item

	maxLength float64

	itemsVolume float64
	itemsWeight float64
}

type boxSlice []*Box

func (bs boxSlice) Len() int {
	return len(bs)
}

func (bs boxSlice) Less(i, j int) bool {
	if bs[i] == nil {
		return false
	}

	if bs[j] == nil {
		return true
	}

	return bs[i].volume < bs[j].volume
}

func (bs boxSlice) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

// NewBox creates a new Box with the given id, dimensions, and maximum weight.
func NewBox(id string, w, h, d, mw float64) *Box {
	//nolint:exhaustruct
	return &Box{
		id:        id,
		width:     w,
		height:    h,
		depth:     d,
		maxWeight: mw,
		maxLength: max(w, h, d),
		volume:    w * h * d,
		items:     make([]*Item, 0, 1),
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

// GetItems returns a copy of the items slice.
func (b *Box) GetItems() []*Item {
	return append([]*Item(nil), b.items...)
}

func (b *Box) GetRemainingVolume() float64 {
	return b.volume - b.itemsVolume
}

func (b *Box) PutItem(item *Item, p Pivot) bool {
	if item == nil {
		return false
	}

	if !b.canQuota(item) {
		return false
	}

	item.position = p

	for rt := RotationTypeWhd; rt <= RotationTypeWdh; rt++ {
		matrix := rotationMatrix[rt]

		itemWidth := item.whd[matrix[WidthAxis]]
		itemHeight := item.whd[matrix[HeightAxis]]
		itemDepth := item.whd[matrix[DepthAxis]]

		if b.width < p[WidthAxis]+itemWidth ||
			b.height < p[HeightAxis]+itemHeight ||
			b.depth < p[DepthAxis]+itemDepth {
			continue
		}

		item.setRotationType(rt)

		if b.itemsIntersect(item) {
			continue
		}

		b.insert(item)

		return true
	}

	return false
}

func (b *Box) itemsIntersect(item *Item) bool {
	if item == nil {
		return false
	}

	for _, ib := range b.items {
		if ib != nil && ib.Intersect(item) {
			return true
		}
	}

	return false
}

func (b *Box) canQuota(item *Item) bool {
	if item == nil {
		return false
	}

	return b.canFitVolume(item) && b.canFitWeight(item)
}

func (b *Box) canFitVolume(item *Item) bool {
	if item == nil {
		return false
	}

	return b.itemsVolume+item.volume <= b.volume
}

func (b *Box) canFitWeight(item *Item) bool {
	if item == nil {
		return false
	}

	return b.itemsWeight+item.weight <= b.maxWeight
}

//nolint:ireturn
func (b *Box) clone() cloner {
	copyBox := &Box{
		id:          b.id,
		width:       b.width,
		height:      b.height,
		depth:       b.depth,
		maxWeight:   b.maxWeight,
		volume:      b.volume,
		maxLength:   b.maxLength,
		itemsVolume: b.itemsVolume,
		itemsWeight: b.itemsWeight,
	}
	if b.items != nil {
		copyBox.items = make([]*Item, len(b.items), cap(b.items))
		copy(copyBox.items, b.items)
	} else {
		copyBox.items = nil
	}

	return copyBox
}

func (b *Box) insert(item *Item) {
	b.items = append(b.items, item)
	b.itemsVolume += item.volume
	b.itemsWeight += item.weight
}

func (b *Box) Reset() {
	b.items = b.items[:0]
	b.itemsVolume = 0
	b.itemsWeight = 0
}
