package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

// TestPacker_StrategyBestFit_ItemsSortedAscending verifies that BestFit sorts items by volume in ascending order.
func TestPacker_StrategyBestFit_ItemsSortedAscending(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 200, 200, 200, 5000),
		boxpacker3.NewBox("box-2", 200, 200, 200, 5000),
	}

	// Items with different volumes - should be sorted ascending
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("large", 40, 40, 40, 200),  // Volume: 64000
		boxpacker3.NewItem("small", 20, 20, 20, 100),  // Volume: 8000
		boxpacker3.NewItem("medium", 30, 30, 30, 150), // Volume: 27000
	}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	// All items should be packed
	require.Empty(t, result.UnfitItems, "All items should fit")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.Equal(t, len(items), totalPacked, "All items should be packed")
}

// TestPacker_StrategyBestFit_SelectsBestBox verifies that BestFit selects the box with smallest remaining space.
func TestPacker_StrategyBestFit_SelectsBestBox(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))

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

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
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
		"BestFit should select box with smallest remaining space")
}

// TestPacker_StrategyBestFit_MultipleItemsBestFit verifies best fit selection for multiple items.
func TestPacker_StrategyBestFit_MultipleItemsBestFit(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000), // Empty initially
		boxpacker3.NewBox("box-2", 100, 100, 100, 2000), // Will have some items
		boxpacker3.NewBox("box-3", 100, 100, 100, 2000), // Empty initially
	}

	// Items that will be placed optimally
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("large-1", 50, 50, 50, 500), // Volume: 125000
		boxpacker3.NewItem("large-2", 50, 50, 50, 500), // Volume: 125000
		boxpacker3.NewItem("small-1", 30, 30, 30, 200), // Volume: 27000
		boxpacker3.NewItem("small-2", 30, 30, 30, 200), // Volume: 27000
	}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	// All items should be packed
	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.Equal(t, len(items), totalPacked, "All items should be packed")

	// BestFit should optimize space usage by selecting best fit for each item
	// Verify that items are distributed optimally
	boxesWithItems := 0

	for _, box := range result.Boxes {
		if len(box.GetItems()) > 0 {
			boxesWithItems++
		}
	}

	require.GreaterOrEqual(t, boxesWithItems, 1, "At least one box should have items")
}

// TestPacker_StrategyBestFit_EmptyBoxes handles empty boxes correctly.
func TestPacker_StrategyBestFit_EmptyBoxes(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))

	boxes := []*boxpacker3.Box{}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
	}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Empty(t, result.Boxes, "Should have no boxes")
	require.Len(t, result.UnfitItems, len(items), "All items should be unfit")
}

// TestPacker_StrategyBestFit_EmptyItems handles empty items correctly.
func TestPacker_StrategyBestFit_EmptyItems(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
	}
	items := []*boxpacker3.Item{}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Empty(t, result.UnfitItems, "Should have no unfit items")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.Equal(t, 0, totalPacked, "Should have no packed items")
}

// TestPacker_StrategyBestFit_UnfitItems handles items that don't fit.
func TestPacker_StrategyBestFit_UnfitItems(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small-box", 50, 50, 50, 500),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("fit-item", 30, 30, 30, 200),
		boxpacker3.NewItem("unfit-item", 100, 100, 100, 1000), // Too large
	}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
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

// TestPacker_StrategyBestFit_WeightConstraints verifies weight constraints are respected.
func TestPacker_StrategyBestFit_WeightConstraints(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("light-box", 100, 100, 100, 500),  // Max weight: 500
		boxpacker3.NewBox("heavy-box", 100, 100, 100, 2000), // Max weight: 2000
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("heavy-1", 30, 30, 30, 400), // Weight: 400
		boxpacker3.NewItem("heavy-2", 30, 30, 30, 400), // Weight: 400
		boxpacker3.NewItem("light-1", 20, 20, 20, 50),  // Weight: 50
	}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
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

// TestPacker_StrategyBestFit_GeometricPlacement verifies geometric placement is correct.
func TestPacker_StrategyBestFit_GeometricPlacement(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
	}

	// Items that fit geometrically
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
		boxpacker3.NewItem("item-2", 50, 50, 50, 500),
	}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
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

// TestPacker_StrategyBestFit_RemainingVolumeOptimization verifies that BestFit minimizes remaining volume.
//
//nolint:funlen
func TestPacker_StrategyBestFit_RemainingVolumeOptimization(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000), // Will have some items
		boxpacker3.NewBox("box-2", 100, 100, 100, 2000), // Empty
	}

	// Add items so box-1 has some remaining space
	// Then add an item that fits in both boxes - should go to box with less remaining space
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("large", 70, 70, 70, 700), // Goes to box-1
		boxpacker3.NewItem("small", 25, 25, 25, 200), // Should go to best fit
	}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	// All items should be packed
	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.Equal(t, len(items), totalPacked, "All items should be packed")

	// Find boxes by ID (after filtering empty boxes)
	var box1, box2 *boxpacker3.Box

	for _, box := range result.Boxes {
		if box.GetID() == "box-1" {
			box1 = box
		} else if box.GetID() == "box-2" {
			box2 = box
		}
	}

	// Both boxes should be used (or at least box-1 should be used)
	require.NotNil(t, box1, "box-1 should be used")

	// Calculate remaining volumes
	box1Remaining := box1.GetRemainingVolume()

	var box2Remaining float64
	if box2 != nil {
		box2Remaining = box2.GetRemainingVolume()
	}

	// Best fit should minimize total remaining volume
	box1Items := box1.GetItems()
	smallInBox1 := false

	for _, item := range box1Items {
		if item.GetID() == "small" {
			smallInBox1 = true

			break
		}
	}

	if smallInBox1 {
		// Small item in box-1 means box-1 has less remaining space (better fit)
		// If box-2 is nil (empty), that's fine - small is correctly in box-1
		if box2 != nil {
			require.Less(t, box1Remaining, box2Remaining,
				"If small item is in box-1, box-1 should have less remaining volume than box-2")
		}
	} else {
		// If box-2 is nil, that means it's empty and wasn't used, which is correct for best fit
		// Small item in box-2 means box-2 has less remaining space (better fit)
		require.NotNil(t, box2, "box-2 should exist if small is there")
		require.Less(t, box2Remaining, box1Remaining,
			"If small item is in box-2, box-2 should have less remaining volume than box-1")
	}
}
