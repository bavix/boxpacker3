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

	for _, box := range result.Boxes {
		if items = p.packToBox(box, items); len(items) == 0 {
			break
		}
	}

	if len(items) > 0 {
		result.UnfitItems = append(result.UnfitItems, items...)
	}

	return result, nil
}

func (p *Packer) preferredSort(boxes boxSlice, items itemSlice) boxSlice {
	volume := 0.
	weight := 0.
	maxLength := 0.

	for _, item := range items {
		volume += item.GetVolume()
		weight += item.GetWeight()

		// optimize
		if maxLength < item.maxLength {
			maxLength = item.maxLength
		}
	}

	for i, b := range boxes {
		if b.volume >= volume && b.maxWeight >= weight && b.maxLength >= maxLength {
			boxes = append(boxSlice{b}, slices.Delete(boxes, i, len(boxes))...)

			break
		}
	}

	return boxes
}

// packToBox Packs goods in a box b. Returns unpackaged goods.
//
//nolint:cyclop,gocognit,funlen
func (p *Packer) packToBox(b *Box, items []*Item) []*Item {
	var fitted bool

	unpacked := make([]*Item, 0, len(items))
	pv := Pivot{}

	if b.items == nil && b.PutItem(items[0], Pivot{}) {
		items = items[1:]
	}

	// Packing unpackaged goods.
	for _, i := range items {
		fitted = false

		// Trying anchor points for the box that don't intersect with existing items in the box.
		for j := range b.items {
			dimension := b.items[j].GetDimension()
			for pt := WidthAxis; pt <= DepthAxis && !fitted; pt++ {
				pv[WidthAxis] = b.items[j].position[WidthAxis]
				pv[HeightAxis] = b.items[j].position[HeightAxis]
				pv[DepthAxis] = b.items[j].position[DepthAxis]

				switch pt {
				case WidthAxis:
					pv[WidthAxis] += +dimension[WidthAxis]
				case HeightAxis:
					pv[HeightAxis] += dimension[HeightAxis]
				case DepthAxis:
					pv[DepthAxis] += dimension[DepthAxis]
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
				itemsFit := 0

				for k := 0; k < total && itemsFit < total; k++ {
					// Trying anchor points for the box that don't intersect with existing items in the box.
					for j := len(b.items) - 1; j >= 0; j-- {
						dimension := b.items[j].GetDimension()

						for pt := WidthAxis; pt <= DepthAxis && k < total; pt++ {
							pv[WidthAxis] = b.items[j].position[WidthAxis]
							pv[HeightAxis] = b.items[j].position[HeightAxis]
							pv[DepthAxis] = b.items[j].position[DepthAxis]

							switch pt {
							case WidthAxis:
								pv[WidthAxis] += +dimension[WidthAxis]
							case HeightAxis:
								pv[HeightAxis] += dimension[HeightAxis]
							case DepthAxis:
								pv[DepthAxis] += dimension[DepthAxis]
							}

							if b.PutItem(copyItems[k], pv) {
								itemsFit++

								break
							}
						}
					}
				}

				fitted = itemsFit == total
			}

			if !fitted {
				*b = *backup
			}
		}

		if !fitted {
			unpacked = append(unpacked, i)
		}
	}

	return unpacked
}
