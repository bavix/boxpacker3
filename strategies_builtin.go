package boxpacker3

import (
	"context"
	"sort"
)

const (
	// perfectFitThreshold is the threshold for considering a fit as perfect (very close to 0).
	perfectFitThreshold = 0.01
)

// --- MinimizeBoxesStrategy (Default / First Fit Decreasing) ---

type MinimizeBoxesStrategy struct{}

func NewMinimizeBoxesStrategy() *MinimizeBoxesStrategy { return &MinimizeBoxesStrategy{} }
func (s *MinimizeBoxesStrategy) Name() string          { return "MinimizeBoxes" }

func (s *MinimizeBoxesStrategy) Pack(ctx context.Context, boxes []*Box, items []*Item) (*Result, error) {
	sort.Sort(sort.Reverse(itemSlice(items)))

	return runFirstFit(ctx, boxes, items)
}

// --- GreedyStrategy (First Fit Ascending) ---

type GreedyStrategy struct{}

func NewGreedyStrategy() *GreedyStrategy { return &GreedyStrategy{} }
func (s *GreedyStrategy) Name() string   { return "Greedy" }

func (s *GreedyStrategy) Pack(ctx context.Context, boxes []*Box, items []*Item) (*Result, error) {
	sort.Sort(itemSlice(items))

	return runFirstFit(ctx, boxes, items)
}

// --- BestFitStrategy (Ascending) ---

type BestFitStrategy struct{}

func NewBestFitStrategy() *BestFitStrategy { return &BestFitStrategy{} }
func (s *BestFitStrategy) Name() string    { return "BestFit" }

func (s *BestFitStrategy) Pack(ctx context.Context, boxes []*Box, items []*Item) (*Result, error) {
	sort.Sort(itemSlice(items))

	return runBestFit(ctx, boxes, items)
}

// --- BestFitDecreasingStrategy (Descending) ---

type BestFitDecreasingStrategy struct{}

func NewBestFitDecreasingStrategy() *BestFitDecreasingStrategy { return &BestFitDecreasingStrategy{} }
func (s *BestFitDecreasingStrategy) Name() string              { return "BestFitDecreasing" }

func (s *BestFitDecreasingStrategy) Pack(ctx context.Context, boxes []*Box, items []*Item) (*Result, error) {
	sort.Sort(sort.Reverse(itemSlice(items)))

	return runBestFit(ctx, boxes, items)
}

// --- NextFitStrategy ---

type NextFitStrategy struct{}

func NewNextFitStrategy() *NextFitStrategy { return &NextFitStrategy{} }
func (s *NextFitStrategy) Name() string    { return "NextFit" }

func (s *NextFitStrategy) Pack(ctx context.Context, boxes []*Box, items []*Item) (*Result, error) {
	sort.Sort(itemSlice(items))

	return runNextFit(ctx, boxes, items)
}

// --- WorstFitStrategy ---

type WorstFitStrategy struct{}

func NewWorstFitStrategy() *WorstFitStrategy { return &WorstFitStrategy{} }
func (s *WorstFitStrategy) Name() string     { return "WorstFit" }

func (s *WorstFitStrategy) Pack(ctx context.Context, boxes []*Box, items []*Item) (*Result, error) {
	sort.Sort(itemSlice(items))

	return runWorstFit(ctx, boxes, items, false) // false = include empty boxes
}

// --- AlmostWorstFitStrategy ---

type AlmostWorstFitStrategy struct{}

func NewAlmostWorstFitStrategy() *AlmostWorstFitStrategy { return &AlmostWorstFitStrategy{} }
func (s *AlmostWorstFitStrategy) Name() string           { return "AlmostWorstFit" }

func (s *AlmostWorstFitStrategy) Pack(ctx context.Context, boxes []*Box, items []*Item) (*Result, error) {
	sort.Sort(itemSlice(items))

	return runWorstFit(ctx, boxes, items, true) // true = skip almost empty boxes
}

func prepareData(inputBoxes []*Box, items []*Item) (boxSlice, *Result) {
	boxes := boxSlice(CopySlicePtr(inputBoxes))
	sort.Sort(boxes)

	sortedBoxes := preferredSort(boxes, items)

	result := &Result{
		UnfitItems: make(itemSlice, 0, len(items)),
		Boxes:      sortedBoxes,
	}

	return sortedBoxes, result
}

// checkContext is a small helper to reduce boilerplate.
func checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func runFirstFit(ctx context.Context, boxes []*Box, items []*Item) (*Result, error) {
	sortedBoxes, result := prepareData(boxes, items)
	remainingItems := items

	for _, box := range sortedBoxes {
		err := checkContext(ctx)
		if err != nil {
			return nil, err
		}

		if len(remainingItems) == 0 {
			break
		}

		remainingItems = packToBox(ctx, box, remainingItems)
	}

	result.UnfitItems = append(result.UnfitItems, remainingItems...)

	return result, nil
}

func runBestFit(ctx context.Context, boxes []*Box, items []*Item) (*Result, error) {
	sortedBoxes, result := prepareData(boxes, items)
	unpacked := make([]*Item, 0, len(items))

	for _, item := range items {
		err := checkContext(ctx)
		if err != nil {
			return nil, err
		}

		if item == nil {
			continue
		}

		bestBoxIndex, bestPivot := findBestBoxForItem(sortedBoxes, item)

		if bestBoxIndex >= 0 {
			box := sortedBoxes[bestBoxIndex]
			box.PutItem(item, bestPivot)
		} else {
			unpacked = append(unpacked, item)
		}
	}

	result.UnfitItems = unpacked

	return result, nil
}

// findBestBoxForItem iterates all boxes to find the tightest fit for a single item.
func findBestBoxForItem(boxes []*Box, item *Item) (int, Pivot) {
	bestBox := -1
	bestRemainingVolume := -1.0
	bestPivot := Pivot{}

	for i, box := range boxes {
		if box == nil || !box.canQuota(item) {
			continue
		}

		pivot, found, perfect := evaluateBoxForBestFit(box, item)
		if !found {
			continue
		}

		rem := getVolumeAfterPlacement(box, item, pivot)

		if bestBox == -1 || rem < bestRemainingVolume {
			bestBox = i
			bestRemainingVolume = rem

			bestPivot = pivot
			if perfect {
				return bestBox, bestPivot
			}
		}
	}

	return bestBox, bestPivot
}

func getVolumeAfterPlacement(box *Box, item *Item, pivot Pivot) float64 {
	testBox := CopyPtr(box)
	testBox.PutItem(item, pivot)

	return testBox.GetRemainingVolume()
}

// evaluateBoxForBestFit checks if an item fits in a box (empty or relative)
// and returns the pivot, success bool, and if it was a "perfect" fit.
func evaluateBoxForBestFit(box *Box, item *Item) (Pivot, bool, bool) {
	testBox := CopyPtr(box)
	if testBox.PutItem(item, Pivot{}) {
		if testBox.GetRemainingVolume() < perfectFitThreshold {
			return Pivot{}, true, true
		}

		return Pivot{}, true, false
	}

	return tryPlaceItemInBox(box, item)
}

// tryPlaceItemInBox attempts to place an item relative to existing items in the box.
func tryPlaceItemInBox(box *Box, item *Item) (Pivot, bool, bool) {
	for j := range box.items {
		itemPos := box.items[j].position
		dimension := box.items[j].GetDimension()

		for _, axis := range []Axis{WidthAxis, HeightAxis, DepthAxis} {
			pv := Pivot{itemPos[WidthAxis], itemPos[HeightAxis], itemPos[DepthAxis]}
			pv[axis] += dimension[axis]

			testBox := CopyPtr(box)
			if testBox.PutItem(item, pv) {
				if testBox.GetRemainingVolume() < perfectFitThreshold {
					return pv, true, true
				}

				return pv, true, false
			}
		}
	}

	return Pivot{}, false, false
}

func runNextFit(ctx context.Context, boxes []*Box, items []*Item) (*Result, error) {
	sortedBoxes, result := prepareData(boxes, items)
	unpacked := make([]*Item, 0, len(items))
	currentBoxIndex := 0

	for _, item := range items {
		err := checkContext(ctx)
		if err != nil {
			return nil, err
		}

		if item == nil {
			continue
		}

		fitted := false

		if currentBoxIndex < len(sortedBoxes) {
			box := sortedBoxes[currentBoxIndex]
			if fitInSpecificBox(box, item) {
				fitted = true
			} else {
				currentBoxIndex++
			}
		}

		if !fitted {
			for i := currentBoxIndex; i < len(sortedBoxes); i++ {
				box := sortedBoxes[i]
				if fitInSpecificBox(box, item) {
					fitted = true
					currentBoxIndex = i

					break
				}
			}
		}

		if !fitted {
			unpacked = append(unpacked, item)
		}
	}

	result.UnfitItems = unpacked

	return result, nil
}

// fitInSpecificBox tries to put an item into a specific box (empty or relative).
// It modifies the box in place if successful.
func fitInSpecificBox(box *Box, item *Item) bool {
	if box == nil || !box.canQuota(item) {
		return false
	}

	if box.PutItem(item, Pivot{}) {
		return true
	}

	for j := range box.items {
		itemPos := box.items[j].position
		dimension := box.items[j].GetDimension()

		for _, axis := range []Axis{WidthAxis, HeightAxis, DepthAxis} {
			pv := Pivot{itemPos[WidthAxis], itemPos[HeightAxis], itemPos[DepthAxis]}

			pv[axis] += dimension[axis]
			if box.PutItem(item, pv) {
				return true
			}
		}
	}

	return false
}

func runWorstFit(ctx context.Context, boxes []*Box, items []*Item, skipEmpty bool) (*Result, error) {
	sortedBoxes, result := prepareData(boxes, items)
	unpacked := make([]*Item, 0, len(items))

	for _, item := range items {
		err := checkContext(ctx)
		if err != nil {
			return nil, err
		}

		if item == nil {
			continue
		}

		// Find worst box
		worstBox, worstPivot := findWorstBox(item, sortedBoxes, skipEmpty)

		if worstBox == -1 && skipEmpty {
			worstBox, worstPivot = findWorstBox(item, sortedBoxes, false)
		}

		if worstBox >= 0 {
			box := sortedBoxes[worstBox]
			box.PutItem(item, worstPivot)
		} else {
			unpacked = append(unpacked, item)
		}
	}

	result.UnfitItems = unpacked

	return result, nil
}

func findWorstBox(item *Item, boxes boxSlice, skipEmptyBoxes bool) (int, Pivot) {
	worstBox := -1
	worstRemainingVolume := -1.0
	worstPivot := Pivot{}

	for i, box := range boxes {
		if box == nil || !box.canQuota(item) {
			continue
		}

		if skipEmptyBoxes {
			if box.GetRemainingVolume() > box.GetVolume()*0.8 {
				continue
			}
		}

		pivot, found, _ := evaluateBoxForBestFit(box, item)
		if found {
			rem := getVolumeAfterPlacement(box, item, pivot)
			if worstBox == -1 || rem > worstRemainingVolume {
				worstBox = i
				worstRemainingVolume = rem
				worstPivot = pivot
			}
		}
	}

	return worstBox, worstPivot
}

func preferredSort(boxes boxSlice, items itemSlice) boxSlice {
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

		if b.volume >= volume && b.maxWeight >= weight && b.maxLength >= maxLength {
			result := make(boxSlice, 0, len(boxes))
			result = append(result, b)

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
func packToBox(ctx context.Context, b *Box, items []*Item) []*Item {
	unpacked := make([]*Item, 0, len(items))
	index := 0

	if b.items == nil && len(items) > 0 && b.PutItem(items[index], Pivot{}) {
		index++
	}

	for i := index; i < len(items); i++ {
		err := checkContext(ctx)
		if err != nil {
			return appendRest(unpacked, items, i)
		}

		if !packSingleItem(b, items[i]) {
			unpacked = append(unpacked, items[i])
		}
	}

	return unpacked
}

func appendRest(unpacked []*Item, items []*Item, startIndex int) []*Item {
	for j := startIndex; j < len(items); j++ {
		if items[j] != nil {
			unpacked = append(unpacked, items[j])
		}
	}

	return unpacked
}

func packSingleItem(b *Box, item *Item) bool {
	if item == nil {
		return false
	}

	// 1. Try Simple Pack (Append to existing layout)
	if fitInSpecificBox(b, item) {
		return true
	}

	// 2. Try Repacking (Brute force shuffle)
	// Only attempt if the item technically meets quota but failed geometry,
	// and if there are existing items to reshuffle.
	if b.canQuota(item) && len(b.items) > 0 {
		return attemptRepack(b, item)
	}

	return false
}

// attemptRepack tries to reshuffle the box to fit the new item.
func attemptRepack(b *Box, newItem *Item) bool {
	backup := CopyPtr(b)
	copyItems := CopySlicePtr(b.items)

	backup.Reset()

	if !backup.PutItem(newItem, Pivot{}) {
		return false
	}

	itemsFit := 0

	for _, originalItem := range copyItems {
		if fitInSpecificBox(backup, originalItem) {
			itemsFit++
		} else {
			break
		}
	}

	if itemsFit == len(copyItems) {
		*b = *backup

		return true
	}

	return false
}
