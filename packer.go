package boxpacker3

import (
	"context"
	"sort"
)

const (
	// perfectFitThreshold is the threshold for considering a fit as perfect (very close to 0).
	perfectFitThreshold = 0.01
)

// PackerOption is a functional option for configuring a Packer.
type PackerOption func(*Packer)

// WithStrategy sets the packing strategy.
func WithStrategy(strategy PackingStrategy) PackerOption {
	return func(p *Packer) {
		p.strategy = strategy
	}
}

// Packer packs items into boxes using a configurable strategy.
type Packer struct {
	strategy PackingStrategy
}

// Result represents the result of packing items into boxes.
type Result struct {
	UnfitItems itemSlice
	Boxes      boxSlice
}

// NewPacker creates a new Packer with the default minimize boxes strategy.
func NewPacker(opts ...PackerOption) *Packer {
	p := &Packer{
		strategy: StrategyMinimizeBoxes,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// PackCtx packs items into boxes asynchronously with context support.
func (p *Packer) PackCtx(ctx context.Context, inputBoxes []*Box, inputItems []*Item) (*Result, error) {
	result := make(chan *Result, 1)

	go func() {
		res := p.packInternal(ctx, inputBoxes, inputItems)

		select {
		case result <- res:
		case <-ctx.Done():
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-result:
		return res, nil
	}
}

// Pack packs items into boxes.
//
// Deprecated: Use PackCtx instead. This function is kept for backward compatibility
// but PackCtx provides better control with context support for cancellation.
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
	return p.packInternal(context.Background(), inputBoxes, inputItems)
}

// packInternal packs items into boxes with context support for cancellation.
// This is the internal implementation used by both Pack and PackCtx.
//
//nolint:cyclop,funlen
func (p *Packer) packInternal(ctx context.Context, inputBoxes []*Box, inputItems []*Item) *Result {
	if inputBoxes == nil {
		inputBoxes = []*Box{}
	}

	if inputItems == nil {
		inputItems = []*Item{}
	}

	boxes := boxSlice(CopySlicePtr(inputBoxes))
	items := itemSlice(CopySlicePtr(inputItems))

	sort.Sort(boxes)
	p.sortItemsByStrategy(items)

	result := &Result{
		UnfitItems: make(itemSlice, 0, len(items)),
		Boxes:      p.preferredSort(boxes, items),
	}

	select {
	case <-ctx.Done():
		return result
	default:
	}

	switch p.strategy {
	case StrategyBestFit, StrategyBestFitDecreasing:
		items = p.packWithBestFit(ctx, result.Boxes, items)
	case StrategyMinimizeBoxes:
		for _, box := range result.Boxes {
			select {
			case <-ctx.Done():
				return result
			default:
			}

			if len(items) == 0 {
				break
			}

			items = p.packToBox(ctx, box, items)
		}
	case StrategyGreedy:
		for _, box := range result.Boxes {
			select {
			case <-ctx.Done():
				return result
			default:
			}

			if len(items) == 0 {
				break
			}

			items = p.packToBox(ctx, box, items)
		}
	case StrategyNextFit:
		items = p.packWithNextFit(ctx, result.Boxes, items)
	case StrategyWorstFit:
		items = p.packWithWorstFit(ctx, result.Boxes, items)
	case StrategyAlmostWorstFit:
		items = p.packWithAlmostWorstFit(ctx, result.Boxes, items)
	}

	result.UnfitItems = append(result.UnfitItems, items...)

	return result
}

// sortItemsByStrategy sorts items according to the packing strategy.
func (p *Packer) sortItemsByStrategy(items itemSlice) {
	switch p.strategy {
	case StrategyGreedy, StrategyBestFit, StrategyNextFit, StrategyWorstFit, StrategyAlmostWorstFit:
		// Sort items by volume in ascending order (smallest first)
		sort.Sort(items)
	case StrategyBestFitDecreasing, StrategyMinimizeBoxes:
		// StrategyMinimizeBoxes and StrategyBestFitDecreasing sort items in descending order
		sort.Sort(sort.Reverse(items))
	default:
		// Default: sort items by volume in ascending order
		sort.Sort(items)
	}
}

// packWithBestFit packs items using the best-fit strategy.
// For each item, it finds the box with the smallest remaining space that can accommodate the item.
//
// Optimizations:
// - Early exit when perfect fit is found (remainingVolume = 0)
// - Only creates test boxes when necessary.
//
//nolint:gocognit,nestif,cyclop,funlen
func (p *Packer) packWithBestFit(ctx context.Context, boxes boxSlice, items []*Item) []*Item {
	unpacked := make([]*Item, 0, len(items))

	for idx, item := range items {
		select {
		case <-ctx.Done():
			unpacked = append(unpacked, item)

			for j := idx + 1; j < len(items); j++ {
				if items[j] != nil {
					unpacked = append(unpacked, items[j])
				}
			}

			return unpacked
		default:
		}

		if item == nil {
			continue
		}

		fitted := false
		bestBox := -1
		bestRemainingVolume := -1.0
		bestPivot := Pivot{}
		perfectFit := false // Flag for early exit

		for i, box := range boxes {
			if box == nil {
				continue
			}

			if perfectFit {
				break
			}

			if !box.canQuota(item) {
				continue
			}

			testBox := CopyPtr(box)
			if testBox.PutItem(item, Pivot{}) {
				remainingVolume := testBox.GetRemainingVolume()
				if bestBox == -1 || remainingVolume < bestRemainingVolume {
					bestBox = i
					bestRemainingVolume = remainingVolume
					bestPivot = Pivot{}

					if remainingVolume < perfectFitThreshold {
						perfectFit = true
					}
				}
			}

			if !perfectFit {
				for j := range box.items {
					itemPos := box.items[j].position
					dimension := box.items[j].GetDimension()

					for _, axis := range []Axis{WidthAxis, HeightAxis, DepthAxis} {
						pv := Pivot{
							itemPos[WidthAxis],
							itemPos[HeightAxis],
							itemPos[DepthAxis],
						}
						pv[axis] += dimension[axis]

						testBox := CopyPtr(box)
						if testBox.PutItem(item, pv) {
							remainingVolume := testBox.GetRemainingVolume()
							if bestBox == -1 || remainingVolume < bestRemainingVolume {
								bestBox = i
								bestRemainingVolume = remainingVolume
								bestPivot = pv

								if remainingVolume < perfectFitThreshold {
									perfectFit = true

									break
								}
							}
						}
					}

					if perfectFit {
						break
					}
				}
			}
		}

		if bestBox >= 0 {
			box := boxes[bestBox]
			if box.PutItem(item, bestPivot) {
				fitted = true
			}
		}

		if !fitted {
			unpacked = append(unpacked, item)
		}
	}

	return unpacked
}

// packWithNextFit packs items using the next-fit strategy.
// Items are placed in the current box if it fits, otherwise a new box is used.
//
//nolint:gocognit,nestif,cyclop,funlen
func (p *Packer) packWithNextFit(ctx context.Context, boxes boxSlice, items []*Item) []*Item {
	unpacked := make([]*Item, 0, len(items))
	currentBoxIndex := 0

	for idx, item := range items {
		select {
		case <-ctx.Done():
			unpacked = append(unpacked, item)

			for j := idx + 1; j < len(items); j++ {
				if items[j] != nil {
					unpacked = append(unpacked, items[j])
				}
			}

			return unpacked
		default:
		}

		if item == nil {
			continue
		}

		fitted := false

		if currentBoxIndex < len(boxes) {
			box := boxes[currentBoxIndex]
			if box == nil {
				currentBoxIndex++

				continue
			}

			if box.canQuota(item) {
				if box.PutItem(item, Pivot{}) {
					fitted = true
				} else {
					for j := range box.items {
						// Cache dimension and position to avoid repeated method calls
						itemPos := box.items[j].position
						dimension := box.items[j].GetDimension()

						for _, axis := range []Axis{WidthAxis, HeightAxis, DepthAxis} {
							pv := Pivot{
								itemPos[WidthAxis],
								itemPos[HeightAxis],
								itemPos[DepthAxis],
							}
							pv[axis] += dimension[axis]

							if box.PutItem(item, pv) {
								fitted = true

								break
							}
						}

						if fitted {
							break
						}
					}
				}
			}

			if !fitted {
				currentBoxIndex++
			}
		}

		if !fitted {
			for i := currentBoxIndex; i < len(boxes); i++ {
				box := boxes[i]
				if box == nil {
					continue
				}

				if !box.canQuota(item) {
					continue
				}

				if box.PutItem(item, Pivot{}) {
					fitted = true
					currentBoxIndex = i

					break
				}

				for j := range box.items {
					// Cache dimension and position to avoid repeated method calls
					itemPos := box.items[j].position
					dimension := box.items[j].GetDimension()

					for _, axis := range []Axis{WidthAxis, HeightAxis, DepthAxis} {
						pv := Pivot{
							itemPos[WidthAxis],
							itemPos[HeightAxis],
							itemPos[DepthAxis],
						}
						pv[axis] += dimension[axis]

						if box.PutItem(item, pv) {
							fitted = true
							currentBoxIndex = i

							break
						}
					}

					if fitted {
						break
					}
				}

				if fitted {
					break
				}
			}
		}

		if !fitted {
			unpacked = append(unpacked, item)
		}
	}

	return unpacked
}

// findWorstBox finds the box with the largest remaining space that can accommodate the item.
// skipEmptyBoxes: if true, skips boxes that are more than 80% empty (for almost-worst-fit).
// Returns: worstBox index, worstPivot.
//
//nolint:gocognit,cyclop
func (p *Packer) findWorstBox(item *Item, boxes boxSlice, skipEmptyBoxes bool) (int, Pivot) {
	worstBox := -1
	worstRemainingVolume := -1.0
	worstPivot := Pivot{}

	for i, box := range boxes {
		if box == nil {
			continue
		}

		if !box.canQuota(item) {
			continue
		}

		// Skip boxes that are almost empty (more than 80% free space) if requested
		if skipEmptyBoxes {
			boxRemainingVolume := box.GetRemainingVolume()
			if boxRemainingVolume > box.GetVolume()*0.8 {
				continue
			}
		}

		testBox := CopyPtr(box)
		if testBox.PutItem(item, Pivot{}) {
			remainingVolume := testBox.GetRemainingVolume()
			if worstBox == -1 || remainingVolume > worstRemainingVolume {
				worstBox = i
				worstRemainingVolume = remainingVolume
				worstPivot = Pivot{}
			}
		}

		for j := range box.items {
			// Cache dimension and position to avoid repeated method calls
			itemPos := box.items[j].position
			dimension := box.items[j].GetDimension()

			for _, axis := range []Axis{WidthAxis, HeightAxis, DepthAxis} {
				pv := Pivot{
					itemPos[WidthAxis],
					itemPos[HeightAxis],
					itemPos[DepthAxis],
				}
				pv[axis] += dimension[axis]

				testBox := CopyPtr(box)
				if testBox.PutItem(item, pv) {
					remainingVolume := testBox.GetRemainingVolume()
					if worstBox == -1 || remainingVolume > worstRemainingVolume {
						worstBox = i
						worstRemainingVolume = remainingVolume
						worstPivot = pv
					}
				}
			}
		}
	}

	return worstBox, worstPivot
}

// packWithWorstFit packs items using the worst-fit strategy.
// For each item, it finds the box with the largest remaining space that can accommodate the item.
func (p *Packer) packWithWorstFit(ctx context.Context, boxes boxSlice, items []*Item) []*Item {
	unpacked := make([]*Item, 0, len(items))

	for idx, item := range items {
		select {
		case <-ctx.Done():
			unpacked = append(unpacked, item)

			for j := idx + 1; j < len(items); j++ {
				if items[j] != nil {
					unpacked = append(unpacked, items[j])
				}
			}

			return unpacked
		default:
		}

		if item == nil {
			continue
		}

		fitted := false

		// Find the worst box (largest remaining space) for this item
		worstBox, worstPivot := p.findWorstBox(item, boxes, false)

		if worstBox >= 0 {
			box := boxes[worstBox]
			if box.PutItem(item, worstPivot) {
				fitted = true
			}
		}

		if !fitted {
			unpacked = append(unpacked, item)
		}
	}

	return unpacked
}

// packWithAlmostWorstFit packs items using the almost-worst-fit strategy.
// Similar to Worst Fit, but excludes boxes that are too large (almost empty).
// This prevents items from being placed in boxes that are nearly empty.
//
//nolint:cyclop
func (p *Packer) packWithAlmostWorstFit(ctx context.Context, boxes boxSlice, items []*Item) []*Item {
	unpacked := make([]*Item, 0, len(items))

	for idx, item := range items {
		select {
		case <-ctx.Done():
			unpacked = append(unpacked, item)

			for j := idx + 1; j < len(items); j++ {
				if items[j] != nil {
					unpacked = append(unpacked, items[j])
				}
			}

			return unpacked
		default:
		}

		if item == nil {
			continue
		}

		fitted := false

		// Find the worst box (largest remaining space) for this item
		// but exclude boxes that are almost empty (more than 80% empty)
		worstBox, worstPivot := p.findWorstBox(item, boxes, true)

		// If no suitable box found with almost-worst-fit criteria, try worst-fit
		if worstBox == -1 {
			// Fall back to worst-fit without the 80% restriction
			worstBox, worstPivot = p.findWorstBox(item, boxes, false)
		}

		if worstBox >= 0 {
			box := boxes[worstBox]
			if box.PutItem(item, worstPivot) {
				fitted = true
			}
		}

		if !fitted {
			unpacked = append(unpacked, item)
		}
	}

	return unpacked
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
	var volume, weight, maxLength float64

	for _, item := range items {
		if item == nil {
			continue
		}

		volume += item.GetVolume()
		weight += item.GetWeight()
		maxLength = max(maxLength, item.maxLength)
	}

	for i, b := range boxes {
		if b == nil {
			continue
		}

		// A box can accommodate the items if its volume is greater than or equal to the total volume of the items,
		// its maximum weight is greater than or equal to the total weight of the items,
		// and its maximum length is greater than or equal to the maximum length of the items.
		if b.volume >= volume && b.maxWeight >= weight && b.maxLength >= maxLength {
			// Optimize: reuse existing slice capacity instead of creating new one
			// Create result with same capacity as boxes to avoid reallocation
			result := make(boxSlice, 0, len(boxes))
			result = append(result, b)

			// Append remaining boxes (excluding the preferred one)
			for j, box := range boxes {
				if j != i {
					result = append(result, box)
				}
			}

			return result
		}
	}

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
func (p *Packer) packToBox(ctx context.Context, b *Box, items []*Item) []*Item {
	var fitted bool

	unpacked := make([]*Item, 0, len(items))
	pv := Pivot{}
	index := 0

	// If box is empty, first item is put in box
	if b.items == nil && len(items) > 0 && b.PutItem(items[index], pv) {
		index++
	}

	for i := index; i < len(items); i++ {
		select {
		case <-ctx.Done():
			for j := i; j < len(items); j++ {
				if items[j] != nil {
					unpacked = append(unpacked, items[j])
				}
			}

			return unpacked
		default:
		}

		if items[i] == nil {
			unpacked = append(unpacked, items[i])

			continue
		}

		fitted = false

		// Quick check: if item doesn't fit by volume/weight, skip expensive repacking attempt
		// But still try normal placement (canQuota only checks quotas, not geometry)
		skipRepacking := !b.canQuota(items[i])

		if len(b.items) == 0 {
			if b.PutItem(items[i], Pivot{}) {
				fitted = true
			}
		} else {
			for j := range b.items {
				// Cache dimension and position to avoid repeated method calls
				itemPos := b.items[j].position
				dimension := b.items[j].GetDimension()

				for _, axis := range []Axis{WidthAxis, HeightAxis, DepthAxis} {
					// Calculate pivot position (reuse pv array)
					pv[WidthAxis] = itemPos[WidthAxis]
					pv[HeightAxis] = itemPos[HeightAxis]
					pv[DepthAxis] = itemPos[DepthAxis]

					pv[axis] += dimension[axis]

					if b.PutItem(items[i], pv) {
						fitted = true

						break
					}
				}

				if fitted {
					break
				}
			}
		}

		// If item cannot be put in the box, try repacking
		// This is expensive, so skip if item doesn't meet quotas
		// and only do it if we have items already in the box
		if !fitted && !skipRepacking && len(b.items) > 0 {
			backup := CopyPtr(b)
			copyItems := CopySlicePtr(b.items)

			backup.Reset()

			if backup.PutItem(items[i], Pivot{}) {
				itemsFit := 0

				for k := 0; k < len(copyItems) && itemsFit < len(copyItems); k++ {
					itemFitted := false

					// Start from the end of backup.items for better cache locality
					for j := len(backup.items) - 1; j >= 0; j-- {
						dimension := backup.items[j].GetDimension()

						if backup.PutItem(copyItems[k], backup.items[j].position) {
							itemsFit++
							itemFitted = true

							break
						}

						// Pre-calculate base pivot to avoid repeated calculations
						basePivot := backup.items[j].position
						for _, pt := range []Axis{WidthAxis, HeightAxis, DepthAxis} {
							pv[WidthAxis] = basePivot[WidthAxis]
							pv[HeightAxis] = basePivot[HeightAxis]
							pv[DepthAxis] = basePivot[DepthAxis]

							pv[pt] += dimension[pt]

							if backup.PutItem(copyItems[k], pv) {
								itemsFit++
								itemFitted = true

								break
							}
						}

						if itemFitted {
							break
						}
					}

					if !itemFitted {
						break
					}
				}

				fitted = itemsFit == len(copyItems)

				if fitted {
					*b = *backup
				}
			}
		}

		if !fitted {
			unpacked = append(unpacked, items[i])
		}
	}

	return unpacked
}
