package boxpacker3

type Box struct {
	ID        string
	Width     float64
	Height    float64
	Depth     float64
	MaxWeight float64
	Volume    float64
	Items     []*Item

	itemsVolume float64
	itemsWeight float64
}

type BoxSlice []*Box

func (bs BoxSlice) Len() int {
	return len(bs)
}

func (bs BoxSlice) Less(i, j int) bool {
	return bs[i].GetVolume() < bs[j].GetVolume()
}

func (bs BoxSlice) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func NewBox(id string, w, h, d, mw float64) *Box {
	return &Box{
		ID:        id,
		Width:     w,
		Height:    h,
		Depth:     d,
		MaxWeight: mw,
		Volume:    w * h * d,
		Items:     make([]*Item, 0),
	}
}

func (b *Box) GetID() string {
	return b.ID
}

func (b *Box) GetWidth() float64 {
	return b.Width
}

func (b *Box) GetHeight() float64 {
	return b.Height
}

func (b *Box) GetDepth() float64 {
	return b.Depth
}

func (b *Box) GetVolume() float64 {
	return b.Volume
}

func (b *Box) GetMaxWeight() float64 {
	return b.MaxWeight
}

// PutItem Пытается поместить элемент в опорную точку p коробки b.
func (b *Box) PutItem(item *Item, p Pivot) bool {
	fit := false

	if b.itemsVolume+item.GetVolume() > b.GetVolume() {
		return false
	}

	if b.itemsWeight+item.GetWeight() > b.GetMaxWeight() {
		return false
	}

	item.Position = p
	for rt := RotationTypeWhd; rt <= RotationTypeWdh; rt++ {
		item.RotationType = rt
		d := item.GetDimension()

		if b.GetWidth() < p[WidthAxis]+d[WidthAxis] || b.GetHeight() < p[HeightAxis]+d[HeightAxis] || b.GetDepth() < p[DepthAxis]+d[DepthAxis] {
			continue
		}

		fit = true

		for _, ib := range b.Items {
			if ib.Intersect(item) {
				fit = false

				break
			}
		}

		if fit {
			b.insert(item)

			return fit
		}

		break
	}

	return fit
}

func (b *Box) insert(item *Item) {
	b.Items = append(b.Items, item)

	b.itemsVolume += item.GetVolume()
	b.itemsWeight += item.GetWeight()
}
