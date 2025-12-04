package boxpacker3

// MinimizeBoxesGoal prioritizes using the fewest number of boxes possible.
// This is the classic bin packing goal, ideal for reducing shipping label costs.
//
// 1. Maximize items packed (minimize unfit items).
// 2. Minimize number of boxes used.
// 3. Minimize total volume of boxes used (prefer smaller boxes).
func MinimizeBoxesGoal(cand, best *Result) bool {
	if best == nil {
		return true
	}

	// 1. Unfit count (lower is better)
	if len(cand.UnfitItems) != len(best.UnfitItems) {
		return len(cand.UnfitItems) < len(best.UnfitItems)
	}

	// 2. Box count (lower is better)
	cBox, bBox := countUsedBoxes(cand.Boxes), countUsedBoxes(best.Boxes)
	if cBox != bBox {
		return cBox < bBox
	}

	// 3. Total Box Volume (lower is better - means smaller boxes were used)
	return getUsedVolume(cand.Boxes) < getUsedVolume(best.Boxes)
}

// TightestPackingGoal prioritizes high density / volume utilization.
// Ideal when shipping costs are calculated based on dimensional weight or total volume.
//
// 1. Maximize items packed (minimize unfit items).
// 2. Minimize total volume of boxes used.
// 3. Minimize number of boxes used.
func TightestPackingGoal(cand, best *Result) bool {
	if best == nil {
		return true
	}

	// 1. Unfit count (lower is better)
	if len(cand.UnfitItems) != len(best.UnfitItems) {
		return len(cand.UnfitItems) < len(best.UnfitItems)
	}

	// 2. Total Box Volume (lower is better)
	// Using less container volume for the same items = higher density
	cVol, bVol := getUsedVolume(cand.Boxes), getUsedVolume(best.Boxes)
	// Float comparison tolerance could be added here, but simple < is usually sufficient for volume
	if cVol != bVol {
		return cVol < bVol
	}

	// 3. Box count (lower is better)
	return countUsedBoxes(cand.Boxes) < countUsedBoxes(best.Boxes)
}

// MaximizeItemsGoal prioritizes fitting the maximum number of items, regardless of box efficiency.
// Ideal for fixed-container scenarios (e.g., loading a truck) where leaving items behind is the worst outcome.
//
// 1. Maximize items packed (minimize unfit items).
func MaximizeItemsGoal(cand, best *Result) bool {
	if best == nil {
		return true
	}

	// 1. Unfit count (lower is better)
	return len(cand.UnfitItems) < len(best.UnfitItems)
}

// countUsedBoxes helps calculate how many boxes actually contain items.
func countUsedBoxes(boxes []*Box) int {
	n := 0

	for _, b := range boxes {
		if len(b.items) > 0 {
			n++
		}
	}

	return n
}

// getUsedVolume calculates the total capacity of all boxes that contain items.
func getUsedVolume(boxes []*Box) float64 {
	var v float64

	for _, b := range boxes {
		// Only count boxes that actually have items in them
		if len(b.items) > 0 {
			v += b.volume // Use the box's volume (capacity), not itemsVolume
		}
	}

	return v
}
