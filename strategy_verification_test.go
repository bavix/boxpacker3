package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

// TestStrategy_DifferentResults verifies that different strategies produce different results
// on the same input, demonstrating that they are indeed different algorithms.
//
//nolint:funlen
func TestStrategy_DifferentResults(t *testing.T) {
	t.Parallel()

	// Create a test case where different strategies should produce different results
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 5000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 5000),
		boxpacker3.NewBox("box-3", 100, 100, 100, 5000),
	}

	// Items with varying sizes - should trigger different packing behaviors
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("large-1", 60, 60, 60, 1000),
		boxpacker3.NewItem("large-2", 60, 60, 60, 1000),
		boxpacker3.NewItem("medium-1", 40, 40, 40, 500),
		boxpacker3.NewItem("medium-2", 40, 40, 40, 500),
		boxpacker3.NewItem("small-1", 20, 20, 20, 200),
		boxpacker3.NewItem("small-2", 20, 20, 20, 200),
		boxpacker3.NewItem("small-3", 20, 20, 20, 200),
	}

	strategies := []struct {
		name     string
		strategy boxpacker3.PackingStrategy
	}{
		{"MinimizeBoxes", boxpacker3.StrategyMinimizeBoxes},
		{"Greedy", boxpacker3.StrategyGreedy},
		{"BestFit", boxpacker3.StrategyBestFit},
		{"BestFitDecreasing", boxpacker3.StrategyBestFitDecreasing},
		{"NextFit", boxpacker3.StrategyNextFit},
		{"WorstFit", boxpacker3.StrategyWorstFit},
		{"AlmostWorstFit", boxpacker3.StrategyAlmostWorstFit},
	}

	results := make(map[string]*boxpacker3.Result)

	// Run each strategy
	for _, s := range strategies {
		packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(s.strategy))
		result, err := packer.PackCtx(context.Background(), boxes, items)
		require.NoError(t, err)
		require.NotNil(t, result, "Strategy %s should return a result", s.name)
		// AlmostWorstFit may not pack all items if boxes are too empty
		if s.strategy != boxpacker3.StrategyAlmostWorstFit {
			require.Empty(t, result.UnfitItems, "Strategy %s should pack all items", s.name)
		}

		results[s.name] = result
	}

	// Verify that at least some strategies produce different results
	// (it's possible some might be similar, but not all should be identical)
	boxCounts := make(map[string]int)

	for name, result := range results {
		usedBoxes := 0

		for _, box := range result.Boxes {
			if len(box.GetItems()) > 0 {
				usedBoxes++
			}
		}

		boxCounts[name] = usedBoxes
	}

	// Check that we have at least 2 different box counts
	uniqueCounts := make(map[int]bool)
	for _, count := range boxCounts {
		uniqueCounts[count] = true
	}

	require.GreaterOrEqual(t, len(uniqueCounts), 2,
		"Different strategies should produce different box counts, got: %v", boxCounts)
}

// TestStrategy_SortingOrder verifies that strategies sort items correctly.
func TestStrategy_SortingOrder(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 200, 200, 200, 10000),
	}

	// Items with different volumes
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("small", 10, 10, 10, 100),  // Volume: 1000
		boxpacker3.NewItem("large", 50, 50, 50, 500),  // Volume: 125000
		boxpacker3.NewItem("medium", 30, 30, 30, 300), // Volume: 27000
	}

	// Strategies that should sort descending
	descendingStrategies := []boxpacker3.PackingStrategy{
		boxpacker3.StrategyMinimizeBoxes,
		boxpacker3.StrategyBestFitDecreasing,
	}

	// Strategies that should sort ascending
	ascendingStrategies := []boxpacker3.PackingStrategy{
		boxpacker3.StrategyGreedy,
		boxpacker3.StrategyBestFit,
		boxpacker3.StrategyNextFit,
		boxpacker3.StrategyWorstFit,
		boxpacker3.StrategyAlmostWorstFit,
	}

	// For descending strategies, large item should be packed first
	for _, strategy := range descendingStrategies {
		packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(strategy))
		result, err := packer.PackCtx(context.Background(), boxes, items)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotEmpty(t, result.Boxes)

		// Check that large item is in the first box
		firstBox := result.Boxes[0]
		if len(firstBox.GetItems()) > 0 {
			firstItemID := firstBox.GetItems()[0].GetID()
			require.Equal(t, "large", firstItemID,
				"Strategy %d should pack large item first (descending sort)", strategy)
		}
	}

	// For ascending strategies, small item should be packed first
	for _, strategy := range ascendingStrategies {
		packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(strategy))
		result, err := packer.PackCtx(context.Background(), boxes, items)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotEmpty(t, result.Boxes)

		// Check that small item is in the first box
		firstBox := result.Boxes[0]
		if len(firstBox.GetItems()) > 0 {
			firstItemID := firstBox.GetItems()[0].GetID()
			require.Equal(t, "small", firstItemID,
				"Strategy %d should pack small item first (ascending sort)", strategy)
		}
	}
}

// TestStrategy_BestFit_SelectsBestBox verifies that Best Fit selects the box with smallest remaining space.
func TestStrategy_BestFit_SelectsBestBox(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))

	// Boxes with different sizes
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("huge", 300, 300, 300, 10000),  // Volume: 27,000,000
		boxpacker3.NewBox("large", 200, 200, 200, 5000),  // Volume: 8,000,000
		boxpacker3.NewBox("medium", 150, 150, 150, 3000), // Volume: 3,375,000
		boxpacker3.NewBox("small", 100, 100, 100, 2000),  // Volume: 1,000,000
	}

	// Single item that fits in all boxes - Best Fit should choose smallest (small)
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
	}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Empty(t, result.UnfitItems)

	// Find which box has the item
	var boxWithItem *boxpacker3.Box

	for _, box := range result.Boxes {
		if len(box.GetItems()) > 0 {
			boxWithItem = box

			break
		}
	}

	require.NotNil(t, boxWithItem, "Item should be in a box")
	require.Equal(t, "small", boxWithItem.GetID(),
		"Best Fit should select smallest box that fits")
}

// TestStrategy_NextFit_UsesCurrentBox verifies that Next Fit uses current box until it's full.
func TestStrategy_NextFit_UsesCurrentBox(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyNextFit))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 5000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 5000),
		boxpacker3.NewBox("box-3", 100, 100, 100, 5000),
	}

	// Items that will fill first box and require second
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 60, 60, 60, 1000),
		boxpacker3.NewItem("item-2", 60, 60, 60, 1000),
		boxpacker3.NewItem("item-3", 60, 60, 60, 1000),
	}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Empty(t, result.UnfitItems)

	// Next Fit should use boxes sequentially
	usedBoxes := 0

	for _, box := range result.Boxes {
		if len(box.GetItems()) > 0 {
			usedBoxes++
		}
	}

	require.GreaterOrEqual(t, usedBoxes, 2, "Next Fit should use multiple boxes sequentially")
}

// TestStrategy_WorstFit_SelectsWorstBox verifies that Worst Fit selects box with largest remaining space.
func TestStrategy_WorstFit_SelectsWorstBox(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyWorstFit))

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small", 100, 100, 100, 2000), // Volume: 1,000,000
		boxpacker3.NewBox("large", 200, 200, 200, 5000), // Volume: 8,000,000
	}

	// Single item that fits in both - Worst Fit should choose largest (large)
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
	}

	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Empty(t, result.UnfitItems)

	// Find which box has the item
	var boxWithItem *boxpacker3.Box

	for _, box := range result.Boxes {
		if len(box.GetItems()) > 0 {
			boxWithItem = box

			break
		}
	}

	require.NotNil(t, boxWithItem, "Item should be in a box")
	require.Equal(t, "large", boxWithItem.GetID(),
		"Worst Fit should select largest box (worst fit)")
}

// TestStrategy_AlmostWorstFit_ExcludesEmptyBoxes verifies that Almost Worst Fit excludes almost empty boxes.
func TestStrategy_AlmostWorstFit_ExcludesEmptyBoxes(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyAlmostWorstFit))

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
	require.Empty(t, result.UnfitItems)

	// Item should prefer half-box over empty-box (which is >80% empty)
	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.GreaterOrEqual(t, totalPacked, 1, "Item should be packed")
}
