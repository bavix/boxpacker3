package boxpacker3

import (
	"slices"
	"sort"
)

// Packer is a struct that packs items into boxes.
//
// It sorts input boxes and items by volume and weight.
// It then selects the box with the largest volume and weight
// that can accommodate the items. If there are still items left
// after packing the boxes, it sets them as unfit items.
type Packer struct{}

// Result represents the result of packing items into boxes.
//
// It is a struct that contains two slices:
// - UnfitItems: a list of items that didn't fit into boxes.
// - Boxes: a list of boxes with items.
type Result struct {
	// UnfitItems is a list of items that didn't fit into boxes.
	UnfitItems itemSlice

	// Boxes is a list of boxes with items.
	Boxes boxSlice
}

// NewPacker creates a new instance of Packer.
//
// Returns:
// - a pointer to a Packer struct.
func NewPacker() *Packer {
	return &Packer{}
}

// Pack packs items into boxes.
//
// This function sorts input boxes and items by volume and weight.
// It selects the box with the largest volume and weight that
// can accommodate the items. If there are still items left
// after packing the boxes, they will be set as unfit items.
//
// Parameters:
// - inputBoxes: a list of boxes.
// - inputItems: a list of items.
//
// Returns:
// - a Result struct that contains two slices:
//   - Boxes: a list of boxes with items.
//   - UnfitItems: a list of items that didn't fit into boxes.
func (p *Packer) Pack(inputBoxes []*Box, inputItems []*Item) *Result {
	// Copy input slices to avoid modifying them.
	boxes := boxSlice(CopySlicePtr(inputBoxes))
	items := itemSlice(CopySlicePtr(inputItems))

	// Sort boxes and items in ascending order.
	sort.Sort(boxes)
	sort.Sort(items)

	// Create a new Result struct with empty slices.
	result := &Result{
		UnfitItems: nil,
		Boxes:      p.preferredSort(boxes, items),
	}

	// Pack items into boxes.
	for _, box := range result.Boxes {
		// If there are no items left, exit the loop.
		if len(items) == 0 {
			break
		}

		// Pack items into the box.
		items = p.packToBox(box, items)
	}

	// If there are still items left, set them as unfit items.
	if len(items) > 0 {
		result.UnfitItems = append(result.UnfitItems, items...)
	}

	return result
}

// preferredSort selects the box with the largest volume and weight
// that can accommodate the items.
//
// Parameters:
// - boxes: a slice of boxes.
// - items: a slice of items.
//
// Returns:
//   - a slice of boxes sorted by volume, weight and maxLength.
//     The first box in the slice is the preferred box.
//     The remaining boxes are sorted after the preferred box.
//     If there is no box that can accommodate the items, the original
//     slice of boxes is returned.
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
// It attempts to pack items into the box. If the box is not big enough
// to accommodate all the items, it tries to pack the items again
// into the box by rotating the box and items.
// If the box is still not big enough, it adds the items to the unpacked slice.
//
// Parameters:
// - b: the box to pack items into.
// - items: the items to pack into the box.
//
// Returns:
//   - a slice of items that did not fit into the box.
//
//nolint:funlen,gocognit,cyclop
func (p *Packer) packToBox(b *Box, items []*Item) []*Item {
	var fitted bool

	unpacked := make([]*Item, 0, len(items))
	pv := Pivot{}

	// if box is empty, first item is put in box
	if b.items == nil && len(items) > 0 && b.PutItem(items[0], pv) {
		items = items[1:]
	}

	// try to pack all items into the box
	for _, i := range items {
		fitted = false

		// for each item already in the box
		for j := range b.items {
			dimension := b.items[j].GetDimension()

			// for each axis
			for _, pt := range []Axis{WidthAxis, HeightAxis, DepthAxis} {
				// calculate pivot position
				pv[WidthAxis] = b.items[j].position[WidthAxis]
				pv[HeightAxis] = b.items[j].position[HeightAxis]
				pv[DepthAxis] = b.items[j].position[DepthAxis]

				// add item dimension to pivot position
				switch pt {
				case WidthAxis:
					pv[WidthAxis] += dimension[WidthAxis]
				case HeightAxis:
					pv[HeightAxis] += dimension[HeightAxis]
				case DepthAxis:
					pv[DepthAxis] += dimension[DepthAxis]
				}

				// if item can be put in the box
				if b.PutItem(i, pv) {
					fitted = true

					break
				}
			}
		}

		// if item can not be put in the box
		if !fitted {
			// make a backup of box
			backup := CopyPtr(b)
			copyItems := CopySlicePtr(b.items)

			// clean box
			b.purge()

			// try to put item into the box
			if b.PutItem(i, Pivot{}) {
				// count of items fit in the box
				itemsFit := 0

				// for each item that was in the box
				for k := 0; k < len(copyItems) && itemsFit < len(copyItems); k++ {
					// try to put item in the box
					for j := len(b.items) - 1; j >= 0; j-- {
						dimension := b.items[j].GetDimension()

						// for each axis
						for _, pt := range []Axis{WidthAxis, HeightAxis, DepthAxis} {
							// calculate pivot position
							pv[WidthAxis] = b.items[j].position[WidthAxis]
							pv[HeightAxis] = b.items[j].position[HeightAxis]
							pv[DepthAxis] = b.items[j].position[DepthAxis]

							// add item dimension to pivot position
							switch pt {
							case WidthAxis:
								pv[WidthAxis] += dimension[WidthAxis]
							case HeightAxis:
								pv[HeightAxis] += dimension[HeightAxis]
							case DepthAxis:
								pv[DepthAxis] += dimension[DepthAxis]
							}

							// if item can be put in the box
							if b.PutItem(copyItems[k], pv) {
								itemsFit++

								break
							}
						}
					}
				}

				// if all items that were in the box now fit in the box
				fitted = itemsFit == len(copyItems)
			}

			// if not all items that were in the box now fit in the box
			if !fitted {
				// restore box from backup
				*b = *backup
			}
		}

		// if item can not be put in the box
		if !fitted {
			// add item to unpacked slice
			unpacked = append(unpacked, i)
		}
	}

	// return unpacked slice
	return unpacked
}
