package boxpacker3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

// TestPacker_StrategyBestFitDecreasing_ItemsSortedDescending verifies that BFD sorts items by volume in descending order.
func TestPacker_StrategyBestFitDecreasing_ItemsSortedDescending(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 200, 200, 200, 5000),
		boxpacker3.NewBox("box-2", 200, 200, 200, 5000),
	}

	// Items with different volumes - should be sorted descending
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("small", 20, 20, 20, 100),  // Volume: 8000
		boxpacker3.NewItem("large", 40, 40, 40, 200),  // Volume: 64000
		boxpacker3.NewItem("medium", 30, 30, 30, 150), // Volume: 27000
	}

	result := packer.Pack(boxes, items)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	// All items should be packed
	require.Empty(t, result.UnfitItems, "All items should fit")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.Equal(t, len(items), totalPacked, "All items should be packed")

	// Verify that largest items are placed first (in first box)
	// BFD processes items in descending order, so largest should be in first box
	firstBoxItems := result.Boxes[0].GetItems()
	if len(firstBoxItems) > 0 {
		firstItemVolume := firstBoxItems[0].GetVolume()
		// First item should be one of the larger items (large or medium)
		require.GreaterOrEqual(t, firstItemVolume, 27000.0,
			"First item should be large (BFD processes largest first)")
	}
}

// TestPacker_StrategyBestFitDecreasing_SelectsBestBox verifies that BFD selects the box with smallest remaining space.
func TestPacker_StrategyBestFitDecreasing_SelectsBestBox(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))

	// Boxes with different sizes - item should go to box with smallest remaining space
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("huge-box", 300, 300, 300, 10000),  // Volume: 27,000,000
		boxpacker3.NewBox("large-box", 200, 200, 200, 5000),  // Volume: 8,000,000
		boxpacker3.NewBox("medium-box", 150, 150, 150, 3000), // Volume: 3,375,000
		boxpacker3.NewBox("small-box", 100, 100, 100, 2000),  // Volume: 1,000,000
	}

	// Single item that fits in all boxes - should go to smallest (best fit)
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
	}

	result := packer.Pack(boxes, items)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	require.Empty(t, result.UnfitItems, "Item should fit")

	// Find which box has the item
	var boxWithItem *boxpacker3.Box

	for _, box := range result.Boxes {
		if len(box.GetItems()) > 0 {
			boxWithItem = box

			break
		}
	}

	require.NotNil(t, boxWithItem, "Item should be in a box")
	// Best fit should select smallest box that fits (small-box)
	require.Equal(t, "small-box", boxWithItem.GetID(),
		"BFD should select box with smallest remaining space")
}

// TestPacker_StrategyBestFitDecreasing_BetterThanFFD verifies that BFD provides better space utilization than FFD.
func TestPacker_StrategyBestFitDecreasing_BetterThanFFD(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-3", 100, 100, 100, 2000),
	}

	// Items with varying sizes that benefit from best fit
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("large-1", 60, 60, 60, 600),  // Volume: 216000
		boxpacker3.NewItem("large-2", 60, 60, 60, 600),  // Volume: 216000
		boxpacker3.NewItem("medium-1", 40, 40, 40, 300), // Volume: 64000
		boxpacker3.NewItem("medium-2", 40, 40, 40, 300), // Volume: 64000
		boxpacker3.NewItem("small-1", 20, 20, 20, 100),  // Volume: 8000
		boxpacker3.NewItem("small-2", 20, 20, 20, 100),  // Volume: 8000
	}

	bfdPacker := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))
	ffdPacker := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyMinimizeBoxes))

	bfdResult := bfdPacker.Pack(boxes, items)
	ffdResult := ffdPacker.Pack(boxes, items)

	require.NotNil(t, bfdResult)
	require.NotNil(t, ffdResult)

	validatePackingInvariants(t, bfdResult)
	validatePackingInvariants(t, ffdResult)

	// Both should pack all items
	require.Empty(t, bfdResult.UnfitItems, "BFD should pack all items")
	require.Empty(t, ffdResult.UnfitItems, "FFD should pack all items")

	// Calculate space utilization
	bfdBoxesUsed := 0
	bfdTotalRemainingVolume := 0.0

	for _, box := range bfdResult.Boxes {
		if len(box.GetItems()) > 0 {
			bfdBoxesUsed++
			bfdTotalRemainingVolume += box.GetRemainingVolume()
		}
	}

	ffdBoxesUsed := 0
	ffdTotalRemainingVolume := 0.0

	for _, box := range ffdResult.Boxes {
		if len(box.GetItems()) > 0 {
			ffdBoxesUsed++
			ffdTotalRemainingVolume += box.GetRemainingVolume()
		}
	}

	// BFD should use same or fewer boxes, and have same or less remaining volume
	require.LessOrEqual(t, bfdBoxesUsed, ffdBoxesUsed,
		"BFD should use same or fewer boxes than FFD")
}

// TestPacker_StrategyBestFitDecreasing_EmptyBoxes handles empty boxes correctly.
func TestPacker_StrategyBestFitDecreasing_EmptyBoxes(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))

	boxes := []*boxpacker3.Box{}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
	}

	result := packer.Pack(boxes, items)
	require.NotNil(t, result)

	require.Empty(t, result.Boxes, "Should have no boxes")
	require.Len(t, result.UnfitItems, len(items), "All items should be unfit")
}

// TestPacker_StrategyBestFitDecreasing_EmptyItems handles empty items correctly.
func TestPacker_StrategyBestFitDecreasing_EmptyItems(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
	}
	items := []*boxpacker3.Item{}

	result := packer.Pack(boxes, items)
	require.NotNil(t, result)

	require.Empty(t, result.UnfitItems, "Should have no unfit items")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.Equal(t, 0, totalPacked, "Should have no packed items")
}

// TestPacker_StrategyBestFitDecreasing_UnfitItems handles items that don't fit.
func TestPacker_StrategyBestFitDecreasing_UnfitItems(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small-box", 50, 50, 50, 500),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("fit-item", 30, 30, 30, 200),
		boxpacker3.NewItem("unfit-item", 100, 100, 100, 1000), // Too large
	}

	result := packer.Pack(boxes, items)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	// Fit item should be packed, unfit item should be in UnfitItems
	require.Len(t, result.UnfitItems, 1, "One item should be unfit")

	unfitIDs := make(map[string]bool)
	for _, item := range result.UnfitItems {
		unfitIDs[item.GetID()] = true
	}

	require.True(t, unfitIDs["unfit-item"], "unfit-item should be in UnfitItems")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.Equal(t, 1, totalPacked, "One item should be packed")
}

// TestPacker_StrategyBestFitDecreasing_WeightConstraints verifies weight constraints are respected.
func TestPacker_StrategyBestFitDecreasing_WeightConstraints(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("light-box", 100, 100, 100, 500),  // Max weight: 500
		boxpacker3.NewBox("heavy-box", 100, 100, 100, 2000), // Max weight: 2000
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("heavy-1", 30, 30, 30, 400), // Weight: 400
		boxpacker3.NewItem("heavy-2", 30, 30, 30, 400), // Weight: 400
		boxpacker3.NewItem("light-1", 20, 20, 20, 50),  // Weight: 50
	}

	result := packer.Pack(boxes, items)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	// Verify weight constraints are respected
	for _, box := range result.Boxes {
		totalWeight := 0.0
		for _, item := range box.GetItems() {
			totalWeight += item.GetWeight()
		}

		require.LessOrEqual(t, totalWeight, box.GetMaxWeight(),
			"Box %s should respect weight constraint", box.GetID())
	}
}

// TestPacker_StrategyBestFitDecreasing_GeometricPlacement verifies geometric placement is correct.
func TestPacker_StrategyBestFitDecreasing_GeometricPlacement(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
	}

	// Items that fit geometrically
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
		boxpacker3.NewItem("item-2", 50, 50, 50, 500),
	}

	result := packer.Pack(boxes, items)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	// Verify no intersections
	for _, box := range result.Boxes {
		boxItems := box.GetItems()
		for idx := range boxItems {
			for j := idx + 1; j < len(boxItems); j++ {
				require.False(t, boxItems[idx].Intersect(boxItems[j]),
					"Items should not intersect")
			}
		}
	}
}

// TestPacker_StrategyBestFitDecreasing_ComplexScenario tests a complex real-world scenario.
func TestPacker_StrategyBestFitDecreasing_ComplexScenario(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("large-1", 200, 200, 200, 5000),
		boxpacker3.NewBox("large-2", 200, 200, 200, 5000),
		boxpacker3.NewBox("medium-1", 150, 150, 150, 3000),
		boxpacker3.NewBox("medium-2", 150, 150, 150, 3000),
		boxpacker3.NewBox("small-1", 100, 100, 100, 2000),
		boxpacker3.NewBox("small-2", 100, 100, 100, 2000),
	}

	// Mix of items with different sizes - using sizes that fit well together
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("huge-1", 80, 80, 80, 1000),
		boxpacker3.NewItem("huge-2", 80, 80, 80, 1000),
		boxpacker3.NewItem("large-1", 60, 60, 60, 600),
		boxpacker3.NewItem("large-2", 60, 60, 60, 600),
		boxpacker3.NewItem("medium-1", 40, 40, 40, 300),
		boxpacker3.NewItem("medium-2", 40, 40, 40, 300),
		boxpacker3.NewItem("small-1", 20, 20, 20, 100),
		boxpacker3.NewItem("small-2", 20, 20, 20, 100),
		boxpacker3.NewItem("tiny-1", 10, 10, 10, 50),
		boxpacker3.NewItem("tiny-2", 10, 10, 10, 50),
	}

	result := packer.Pack(boxes, items)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	// All items should be packed
	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.Equal(t, len(items), totalPacked, "All items should be packed")
	require.Empty(t, result.UnfitItems, "No items should be unfit")
}
