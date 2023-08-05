package boxpacker3

import (
	"sort"

	"golang.org/x/exp/slices"
)

type Packer struct{}

type Result struct {
	UnfitItems itemSlice
	Boxes      boxSlice
}

func NewPacker() *Packer {
	return &Packer{}
}

func (p *Packer) Pack(inputBoxes []*Box, inputItems []*Item) (*Result, error) {
	boxes := boxSlice(copySlicePtr(inputBoxes))
	items := itemSlice(copySlicePtr(inputItems))

	sort.Sort(boxes)
	sort.Sort(items)

	result := &Result{
		UnfitItems: make(itemSlice, 0, len(items)),
		Boxes:      p.preferredSort(boxes, items),
	}

	for i := 0; len(items) > 0; i++ {
		box := p.findRightBox(result.Boxes, items[0])
		if box == nil {
			result.UnfitItems = append(result.UnfitItems, items[0])
			items = items[1:]

			continue
		}

		items = p.packToBox(result.Boxes, box, items)
	}

	return result, nil
}

func (p *Packer) preferredSort(boxes boxSlice, items itemSlice) boxSlice {
	volume := 0.
	weight := 0.

	for _, item := range items {
		volume += item.GetVolume()
		weight += item.GetWeight()
	}

	for i, b := range boxes {
		if b.GetVolume() >= volume && b.GetMaxWeight() >= weight {
			boxes = append(boxSlice{b}, slices.Delete(boxes, i, i+1)...)

			break
		}
	}

	return boxes
}

// packToBox Packs goods in a box b. Returns unpackaged goods.
//
//nolint:cyclop,gocognit,funlen
func (p *Packer) packToBox(boxes []*Box, b *Box, items []*Item) []*Item {
	var fitted bool

	unpacked := make([]*Item, 0, len(items))
	pv := Pivot{}

	// Packing unpackaged goods.
	for _, i := range items {
		fitted = false

		// Trying anchor points for the box that don't intersect with existing items in the box.
		for pt := WidthAxis; pt <= DepthAxis && !fitted; pt++ {
			for _, ib := range b.items {
				switch pt {
				case WidthAxis:
					pv = Pivot{ib.position[WidthAxis] + ib.GetWidth(), ib.position[HeightAxis], ib.position[DepthAxis]}
				case HeightAxis:
					pv = Pivot{ib.position[WidthAxis], ib.position[HeightAxis] + ib.GetHeight(), ib.position[DepthAxis]}
				case DepthAxis:
					pv = Pivot{ib.position[WidthAxis], ib.position[HeightAxis], ib.position[DepthAxis] + ib.GetDepth()}
				}

				if b.PutItem(i, pv) {
					fitted = true

					break
				}
			}
		}

		if !fitted {
			backup := copyPtr(b)
			copyItems := copySlicePtr(b.items)

			b.purge()

			if b.PutItem(i, Pivot{}) {
				total := len(copyItems)
				iter := 0

				for k := 0; k < total && iter < total; k++ {
					// Trying anchor points for the box that don't intersect with existing items in the box.
					for pt := WidthAxis; pt <= DepthAxis && iter < total; pt++ {
						for _, ib := range b.items {
							switch pt {
							case WidthAxis:
								pv = Pivot{ib.position[WidthAxis] + ib.GetWidth(), ib.position[HeightAxis], ib.position[DepthAxis]}
							case HeightAxis:
								pv = Pivot{ib.position[WidthAxis], ib.position[HeightAxis] + ib.GetHeight(), ib.position[DepthAxis]}
							case DepthAxis:
								pv = Pivot{ib.position[WidthAxis], ib.position[HeightAxis], ib.position[DepthAxis] + ib.GetDepth()}
							}

							if b.PutItem(copyItems[k], pv) {
								iter++

								break
							}
						}
					}
				}

				fitted = iter == total
			}

			if !fitted {
				*b = *backup
			}
		}

		for lb := p.findBiggerBox(boxes, b); lb != nil && !fitted; lb = p.findBiggerBox(boxes, lb) {
			fitted = len(p.packToBox(boxes, lb, itemSlice{i})) == 0
		}

		if !fitted {
			unpacked = append(unpacked, i)
		}
	}

	return unpacked
}

func (p *Packer) findBiggerBox(boxes []*Box, box *Box) *Box {
	for _, b := range boxes {
		if b.volume > box.volume {
			return b
		}
	}

	return nil
}

func (p *Packer) findRightBox(boxes []*Box, item *Item) *Box {
	for _, b := range boxes {
		if b.volume >= item.volume {
			return b
		}
	}

	return nil
}
