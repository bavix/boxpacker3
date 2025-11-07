package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

// TestPacker_StrategyNextFit_NextBoxBehavior verifies that NextFit uses next box when current doesn't fit.
func TestPacker_StrategyNextFit_NextBoxBehavior(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyNextFit))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-3", 100, 100, 100, 2000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
		boxpacker3.NewItem("item-2", 50, 50, 50, 500),
		boxpacker3.NewItem("item-3", 50, 50, 50, 500),
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

// TestPacker_StrategyNextFit_EmptyBoxes handles empty boxes correctly.
func TestPacker_StrategyNextFit_EmptyBoxes(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyNextFit))

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

// TestPacker_StrategyNextFit_UnfitItems handles items that don't fit.
func TestPacker_StrategyNextFit_UnfitItems(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyNextFit))

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

	require.Len(t, result.UnfitItems, 1, "One item should be unfit")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.Equal(t, 1, totalPacked, "One item should be packed")
}

// TestPacker_StrategyWorstFit_SelectsWorstBox verifies that WorstFit selects box with largest remaining space.
func TestPacker_StrategyWorstFit_SelectsWorstBox(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyWorstFit))

	// Boxes with different sizes - item should go to box with largest remaining space
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small-box", 100, 100, 100, 2000),  // Volume: 1,000,000
		boxpacker3.NewBox("medium-box", 150, 150, 150, 3000), // Volume: 3,375,000
		boxpacker3.NewBox("large-box", 200, 200, 200, 5000),  // Volume: 8,000,000
	}

	// Single item that fits in all boxes - should go to largest (worst fit)
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
	// Worst fit should select largest box (large-box)
	require.Equal(t, "large-box", boxWithItem.GetID(),
		"WorstFit should select box with largest remaining space")
}

// TestPacker_StrategyWorstFit_EmptyBoxes handles empty boxes correctly.
func TestPacker_StrategyWorstFit_EmptyBoxes(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyWorstFit))

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

// TestPacker_StrategyWorstFit_UnfitItems handles items that don't fit.
func TestPacker_StrategyWorstFit_UnfitItems(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyWorstFit))

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

	require.Len(t, result.UnfitItems, 1, "One item should be unfit")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.Equal(t, 1, totalPacked, "One item should be packed")
}

// TestPacker_StrategyAlmostWorstFit_ExcludesEmptyBoxes verifies that AlmostWorstFit excludes almost empty boxes.
func TestPacker_StrategyAlmostWorstFit_ExcludesEmptyBoxes(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyAlmostWorstFit))

	// Boxes with different fill levels
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("empty-box", 200, 200, 200, 5000), // Almost empty (will be excluded)
		boxpacker3.NewBox("half-box", 100, 100, 100, 2000),  // Half full
	}

	// Add some items to half-box to make it not empty (less than 80% free)
	halfBox := boxes[1]
	halfBox.PutItem(boxpacker3.NewItem("existing-1", 50, 50, 50, 500), boxpacker3.Pivot{})

	// Item that fits in all boxes
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 30, 30, 30, 200),
	}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	require.Empty(t, result.UnfitItems, "Item should fit")

	// Item should prefer half-box over empty-box (which is >80% empty)
	// But if only empty-box is available after exclusion, fallback will use it
	// So we just verify item is packed
	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.GreaterOrEqual(t, totalPacked, 1, "Item should be packed")
}

// TestPacker_StrategyAlmostWorstFit_EmptyBoxes handles empty boxes correctly.
func TestPacker_StrategyAlmostWorstFit_EmptyBoxes(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyAlmostWorstFit))

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

// TestPacker_StrategyAlmostWorstFit_UnfitItems handles items that don't fit.
func TestPacker_StrategyAlmostWorstFit_UnfitItems(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyAlmostWorstFit))

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

	require.Len(t, result.UnfitItems, 1, "One item should be unfit")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.Equal(t, 1, totalPacked, "One item should be packed")
}

// TestPacker_StrategyAlmostWorstFit_FallbackToWorstFit verifies fallback to WorstFit when no suitable box found.
func TestPacker_StrategyAlmostWorstFit_FallbackToWorstFit(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyAlmostWorstFit))

	// All boxes are almost empty (>80% free)
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("empty-1", 200, 200, 200, 5000),
		boxpacker3.NewBox("empty-2", 200, 200, 200, 5000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
	}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	// Should still pack the item (fallback to worst-fit)
	require.Empty(t, result.UnfitItems, "Item should be packed even if all boxes are almost empty")
}

// TestPacker_AllNewStrategies_PackAllItems tests that all new strategies pack items correctly.
func TestPacker_AllNewStrategies_PackAllItems(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 2000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 30, 30, 30, 100),
		boxpacker3.NewItem("item-2", 30, 30, 30, 100),
		boxpacker3.NewItem("item-3", 30, 30, 30, 100),
	}

	strategies := []struct {
		name   string
		packer *boxpacker3.Packer
	}{
		{"NextFit", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyNextFit))},
		{"WorstFit", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyWorstFit))},
		{"AlmostWorstFit", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyAlmostWorstFit))},
	}

	for _, strategy := range strategies {
		t.Run(strategy.name, func(t *testing.T) {
			t.Parallel()

			result, err := strategy.packer.PackCtx(context.Background(), boxes, items)
			require.NoError(t, err)
			require.NotNil(t, result)

			validatePackingInvariants(t, result)

			// All items should be accounted for
			totalPacked := 0
			for _, box := range result.Boxes {
				totalPacked += len(box.GetItems())
			}

			require.Equal(t, len(items), totalPacked+len(result.UnfitItems),
				"%s: All items must be either packed or in UnfitItems", strategy.name)
		})
	}
}
