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

func (p *Packer) Pack(inputBoxes []*Box, inputItems []*Item) *Result {
	boxes := boxSlice(copySlicePtr(inputBoxes))
	items := itemSlice(copySlicePtr(inputItems))

	sort.Sort(boxes)
	sort.Sort(items)

	result := &Result{
		UnfitItems: nil,
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

	return result
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
			return append(boxSlice{b}, slices.Delete(boxes, i, i+1)...)
		}
	}

	return boxes
}

// packToBox Packs goods in a box b. Returns unpackaged goods.
//
//nolint:cyclop,gocognit,funlen
func (p *Packer) packToBox(b *Box, items []*Item) []*Item {
	var fitted bool

	cntItems := len(items)
	unpacked := make([]*Item, 0, cntItems)
	pv := Pivot{}

	if b.items == nil && cntItems > 0 && b.PutItem(items[0], pv) {
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
					pv[WidthAxis] += dimension[WidthAxis]
				case HeightAxis:
					pv[HeightAxis] += dimension[HeightAxis]
				case DepthAxis:
					pv[DepthAxis] += dimension[DepthAxis]
				}

				fitted = b.PutItem(i, pv)
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
								pv[WidthAxis] += dimension[WidthAxis]
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
