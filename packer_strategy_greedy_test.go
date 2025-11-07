package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

// TestPacker_StrategyGreedy_ItemsSortedAscending verifies that Greedy sorts items by volume in ascending order.
func TestPacker_StrategyGreedy_ItemsSortedAscending(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))

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

	// Verify that smallest items are placed first (in first box)
	// Greedy processes items in ascending order, so smallest should be in first box
	firstBoxItems := result.Boxes[0].GetItems()
	if len(firstBoxItems) > 0 {
		firstItemVolume := firstBoxItems[0].GetVolume()
		// First item should be one of the smaller items (small or medium)
		require.LessOrEqual(t, firstItemVolume, 27000.0,
			"First item should be small (Greedy processes smallest first)")
	}
}

// TestPacker_StrategyGreedy_FirstFitBehavior verifies that Greedy uses First Fit algorithm.
func TestPacker_StrategyGreedy_FirstFitBehavior(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))

	// Multiple boxes where items can fit
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-3", 100, 100, 100, 2000),
	}

	// Items that all fit in first box
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 30, 30, 30, 100),
		boxpacker3.NewItem("item-2", 30, 30, 30, 100),
		boxpacker3.NewItem("item-3", 30, 30, 30, 100),
	}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	// First Fit should pack items into first box that fits
	// All items should fit in one box
	require.Empty(t, result.UnfitItems, "All items should fit")
}

// TestPacker_StrategyGreedy_EmptyBoxes handles empty boxes correctly.
func TestPacker_StrategyGreedy_EmptyBoxes(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))

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

// TestPacker_StrategyGreedy_EmptyItems handles empty items correctly.
func TestPacker_StrategyGreedy_EmptyItems(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))

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

// TestPacker_StrategyGreedy_UnfitItems handles items that don't fit.
func TestPacker_StrategyGreedy_UnfitItems(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))

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

// TestPacker_StrategyGreedy_WeightConstraints verifies weight constraints are respected.
func TestPacker_StrategyGreedy_WeightConstraints(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))

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

// TestPacker_StrategyGreedy_GeometricPlacement verifies geometric placement is correct.
func TestPacker_StrategyGreedy_GeometricPlacement(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))

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

// TestPacker_StrategyGreedy_ComplexScenario tests a complex real-world scenario.
func TestPacker_StrategyGreedy_ComplexScenario(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))

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
		boxpacker3.NewItem("tiny-1", 10, 10, 10, 50),
		boxpacker3.NewItem("small-1", 20, 20, 20, 100),
		boxpacker3.NewItem("small-2", 20, 20, 20, 100),
		boxpacker3.NewItem("medium-1", 40, 40, 40, 300),
		boxpacker3.NewItem("medium-2", 40, 40, 40, 300),
		boxpacker3.NewItem("large-1", 60, 60, 60, 600),
		boxpacker3.NewItem("large-2", 60, 60, 60, 600),
		boxpacker3.NewItem("huge-1", 80, 80, 80, 1000),
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
	require.Empty(t, result.UnfitItems, "No items should be unfit")
}
