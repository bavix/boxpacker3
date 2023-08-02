package boxpacker3

import (
	"sort"
)

type Packer struct{}

type Result struct {
	UnfitItems ItemSlice
	Boxes      BoxSlice
}

func NewPacker() *Packer {
	return &Packer{}
}

func (p *Packer) Pack(inputBoxes []*Box, inputItems []*Item) (*Result, error) {
	boxes := BoxSlice(copySlicePtr(inputBoxes))
	items := ItemSlice(copySlicePtr(inputItems))

	sort.Sort(boxes)
	sort.Sort(items)

	result := &Result{
		UnfitItems: make(ItemSlice, 0, len(items)),
		Boxes:      p.preferredSort(boxes, items),
	}

	maxIter := len(items)
	for i := 0; len(items) > 0 && i < maxIter; i++ {
		box := p.FindFittedBox(result.Boxes, items[0])
		if box == nil {
			result.UnfitItems = append(result.UnfitItems, items[0])
			items = items[1:]

			continue
		}

		items = p.packToBox(result.Boxes, box, items)
	}

	return result, nil
}

func (p *Packer) preferredSort(boxes BoxSlice, items ItemSlice) BoxSlice {
	volume := 0.
	weight := 0.

	for _, item := range items {
		volume += item.GetVolume()
		weight += item.GetWeight()
	}

	preferredBoxes := make(BoxSlice, 0, len(boxes))
	otherBoxes := make(BoxSlice, 0, len(boxes))

	for _, b := range boxes {
		if b.GetVolume() >= volume && b.GetMaxWeight() >= weight {
			preferredBoxes = append(preferredBoxes, b)
		} else {
			otherBoxes = append(otherBoxes, b)
		}
	}

	return append(preferredBoxes, otherBoxes...)
}

// packToBox Упаковывает товары в коробку b. Возвращает не упакованные товары.
//
//nolint:cyclop,nonamedreturns,gocognit,funlen
func (p *Packer) packToBox(boxes []*Box, b *Box, items []*Item) (unpacked []*Item) {
	if !b.PutItem(items[0], Pivot{0, 0, 0}) {
		if b2 := p.getBiggerBoxThan(boxes, b); b2 != nil {
			return p.packToBox(boxes, b2, items)
		}

		return items
	}

	// Упаковываем неупакованные товары.
	for _, i := range items[1:] {
		var fitted bool

		// Пробуем опорные точки для коробки, которые не пересекаются с существующими товарами в коробке.
		for pt := WidthAxis; pt <= DepthAxis && !fitted; pt++ {
			for _, ib := range b.Items {
				var pv Pivot

				switch pt {
				case WidthAxis:
					pv = Pivot{ib.Position[WidthAxis] + ib.GetWidth(), ib.Position[HeightAxis], ib.Position[DepthAxis]}
				case HeightAxis:
					pv = Pivot{ib.Position[WidthAxis], ib.Position[HeightAxis] + ib.GetHeight(), ib.Position[DepthAxis]}
				case DepthAxis:
					pv = Pivot{ib.Position[WidthAxis], ib.Position[HeightAxis], ib.Position[DepthAxis] + ib.GetDepth()}
				}

				if b.PutItem(i, pv) {
					fitted = true

					break
				}
			}
		}

		//nolint:nestif
		if !fitted {
			forRevert := copySlicePtr(b.Items)
			itemSlice := copySlicePtr(b.Items)
			bkitemsWeight := b.itemsWeight
			bkitemsVolume := b.itemsVolume
			b.Items = []*Item{}
			b.itemsWeight = 0
			b.itemsVolume = 0

			if b.PutItem(i, Pivot{0, 0, 0}) {
				cnt := len(itemSlice)
				for k := 0; k < 100 && cnt > 0; k++ {
					j := itemSlice[0]

					// Пробуем опорные точки для коробки, которые не пересекаются с существующими товарами в коробке.
					for pt := WidthAxis; pt <= DepthAxis && cnt > 0; pt++ {
						for _, ib := range b.Items {
							if j == ib {
								continue
							}

							var pv Pivot

							switch pt {
							case WidthAxis:
								pv = Pivot{ib.Position[WidthAxis] + ib.GetWidth(), ib.Position[HeightAxis], ib.Position[DepthAxis]}
							case HeightAxis:
								pv = Pivot{ib.Position[WidthAxis], ib.Position[HeightAxis] + ib.GetHeight(), ib.Position[DepthAxis]}
							case DepthAxis:
								pv = Pivot{ib.Position[WidthAxis], ib.Position[HeightAxis], ib.Position[DepthAxis] + ib.GetDepth()}
							}

							if b.PutItem(j, pv) {
								itemSlice = itemSlice[1:]
								cnt--

								break
							}
						}
					}
				}

				fitted = len(itemSlice) == 0

				if !fitted {
					b.Items = forRevert
					b.itemsVolume = bkitemsVolume
					b.itemsWeight = bkitemsWeight
				}
			} else {
				b.Items = forRevert
				b.itemsVolume = bkitemsVolume
				b.itemsWeight = bkitemsWeight
			}

			for b2 := p.getBiggerBoxThan(boxes, b); b2 != nil && !fitted; b2 = p.getBiggerBoxThan(boxes, b) {
				if left := p.packToBox(boxes, b2, append(b2.Items, i)); len(left) == 0 {
					b = b2
					fitted = true
				}
			}

			if !fitted {
				unpacked = append(unpacked, i)
			}
		}
	}

	return //nolint:nakedret
}

func (p *Packer) getBiggerBoxThan(boxes []*Box, b *Box) *Box {
	v := b.GetVolume()
	for _, b2 := range boxes {
		if b2.GetVolume() > v {
			return b2
		}
	}

	return nil
}

// FindFittedBox находит коробку для товара.
func (p *Packer) FindFittedBox(boxes []*Box, i *Item) *Box {
	for _, b := range boxes {
		if !b.PutItem(i, Pivot{0, 0, 0}) {
			continue
		}

		if len(b.Items) == 1 && b.Items[0] == i {
			b.itemsVolume -= i.GetVolume()
			b.itemsWeight -= i.GetWeight()

			b.Items = []*Item{}
		}

		return b
	}

	return nil
}
