package boxpacker3

import (
	"context"
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

// PackCtx packs items into boxes asynchronously and handles the context.
//
// This function sorts input boxes and items by volume and weight.
// It selects the box with the largest volume and weight that
// can accommodate the items. If there are still items left
// after packing the boxes, they will be set as unfit items.
//
// Parameters:
// - ctx: the context.Context to use.
// - inputBoxes: a list of boxes.
// - inputItems: a list of items.
//
// Returns:
//   - *Result: the result of packing items into boxes.
//   - error: If the context is done before the packing process is complete,
//     an error will be returned. nil will be returned otherwise.
//
// This function is useful when you want to pack items into boxes
// asynchronously and handle the context at the same time.
func (p *Packer) PackCtx(ctx context.Context, inputBoxes []*Box, inputItems []*Item) (*Result, error) {
	// Create a channel to receive the result of the packing process.
	result := make(chan *Result, 1)

	// Start a goroutine to perform the packing process.
	go func() { result <- p.Pack(inputBoxes, inputItems) }()

	// Wait for the context to be done or the packing process to complete.
	select {
	case <-ctx.Done():
		// If the context is done, return nil.
		return nil, ctx.Err()
	case res := <-result:
		// If the packing process is complete, return the result.
		return res, nil
	}
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
		UnfitItems: make(itemSlice, 0, len(items)),
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
	result.UnfitItems = append(result.UnfitItems, items...)

	return result
}

// preferredSort selects the box with the largest volume and weight
// that can accommodate the items.
//
// This function calculates the maximum volume, weight, and maximum length of the items.
// It then iterates through the boxes and checks if a box can accommodate all the items.
// If a box is found that can accommodate the items, it is returned as the preferred box.
// The remaining boxes are sorted after the preferred box.
// If no box can accommodate the items, the original slice of boxes is returned.
//
// Parameters:
// - boxes: a slice of boxes.
// - items: a slice of items.
//
// Returns:
//   - a slice of boxes sorted by volume, weight, and maximum length.
//     The first box in the slice is the preferred box.
//     The remaining boxes are sorted after the preferred box.
//     If there is no box that can accommodate the items, the original
//     slice of boxes is returned.
func (p *Packer) preferredSort(boxes boxSlice, items itemSlice) boxSlice {
	// Calculate the max volume, weight, and maximum length of the items.
	var volume, weight, maxLength float64

	// Iterate through the items and calculate the total volume, weight, and maximum length.
	for _, item := range items {
		volume += item.GetVolume()
		weight += item.GetWeight()
		maxLength = max(maxLength, item.maxLength)
	}

	// Find the preferred box
	for i, b := range boxes {
		// Check if the box can accommodate the items.
		// A box can accommodate the items if its volume is greater than or equal to the total volume of the items,
		// its maximum weight is greater than or equal to the total weight of the items,
		// and its maximum length is greater than or equal to the maximum length of the items.
		if b.volume >= volume && b.maxWeight >= weight && b.maxLength >= maxLength {
			result := make(boxSlice, 0, len(boxes))

			// If the box can accommodate the items, return the box as the preferred box
			// and the remaining boxes sorted after the preferred box.
			return append(append(result, b), slices.Delete(boxes, i, i+1)...)
		}
	}

	// If there is no box that can accommodate the items, return the original slice of boxes.
	return boxes
}

// packToBox Packs goods in a box b. Returns unpackaged goods.
//
// PackToBox attempts to pack items into the box. If the box is not big enough
// to accommodate all the items, it tries to pack the items again into the box
// by rotating the box and items. If the box is still not big enough, it adds
// the items to the unpacked slice.
//
// Parameters:
// - b: the box to pack items into.
// - items: the items to pack into the box.
//
// Returns:
//   - a slice of items that did not fit into the box.
//
//nolint:funlen,gocognit,cyclop,nestif
func (p *Packer) packToBox(b *Box, items []*Item) []*Item {
	var fitted bool

	unpacked := make([]*Item, 0, len(items))
	pv := Pivot{}
	index := 0

	// If box is empty, first item is put in box
	if b.items == nil && len(items) > 0 && b.PutItem(items[index], pv) {
		index++
	}

	// Try to pack all items into the box
	for i := index; i < len(items); i++ {
		fitted = false

		// Try to put item into the box
		for j := range b.items {
			dimension := b.items[j].GetDimension()

			// Try to put item in each axis
			for _, axis := range []Axis{WidthAxis, HeightAxis, DepthAxis} {
				// Calculate pivot position
				pv[WidthAxis] = b.items[j].position[WidthAxis]
				pv[HeightAxis] = b.items[j].position[HeightAxis]
				pv[DepthAxis] = b.items[j].position[DepthAxis]

				// Add item dimension to pivot position
				pv[axis] += dimension[axis]

				// If item can be put in the box
				if b.PutItem(items[i], pv) {
					fitted = true

					break
				}
			}
		}

		// If item cannot be put in the box
		if !fitted {
			// Make a backup of box
			backup := CopyPtr(b)
			copyItems := CopySlicePtr(b.items)

			backup.Reset()

			// Try to put item into the backup box
			if backup.PutItem(items[i], Pivot{}) {
				// Count of items fit in the box
				itemsFit := 0

				// Try to put each item in the box
				for k := 0; k < len(copyItems) && itemsFit < len(copyItems); k++ {
					for j := len(backup.items) - 1; j >= 0; j-- {
						dimension := backup.items[j].GetDimension()

						// Check if item can be put in the box
						if backup.PutItem(copyItems[k], backup.items[j].position) {
							itemsFit++

							break
						}

						// Try to put item in each axis
						for _, pt := range []Axis{WidthAxis, HeightAxis, DepthAxis} {
							// Calculate pivot position
							pv[WidthAxis] = backup.items[j].position[WidthAxis]
							pv[HeightAxis] = backup.items[j].position[HeightAxis]
							pv[DepthAxis] = backup.items[j].position[DepthAxis]

							// Add item dimension to pivot position
							switch pt {
							case WidthAxis:
								pv[WidthAxis] += dimension[WidthAxis]
							case HeightAxis:
								pv[HeightAxis] += dimension[HeightAxis]
							case DepthAxis:
								pv[DepthAxis] += dimension[DepthAxis]
							}

							// If item can be put in the box
							if backup.PutItem(copyItems[k], pv) {
								itemsFit++

								break
							}
						}
					}
				}

				// If all items that were in the box now fit in the box
				fitted = itemsFit == len(copyItems)

				// If successfully filled, restore backup in b
				if fitted {
					*b = *backup
				}
			}
		}

		// If item cannot be put in the box
		if !fitted {
			// Add item to unpacked slice
			unpacked = append(unpacked, items[i])
		}
	}

	// return unpacked slice
	return unpacked
}
