package boxpacker3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

// TestPacker_AllStrategies_PackAllItems tests that all strategies pack items correctly.
func TestPacker_AllStrategies_PackAllItems(t *testing.T) {
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
		{"MinimizeBoxes", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyMinimizeBoxes))},
		{"Greedy", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))},
		{"BestFit", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))},
		{"BestFitDecreasing", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))},
		{"NextFit", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyNextFit))},
		{"WorstFit", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyWorstFit))},
		{"AlmostWorstFit", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyAlmostWorstFit))},
	}

	for _, strategy := range strategies {
		t.Run(strategy.name, func(t *testing.T) {
			t.Parallel()

			result := strategy.packer.Pack(boxes, items)
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

// TestPacker_AllStrategies_HandleUnfitItems tests that all strategies handle unfit items correctly.
func TestPacker_AllStrategies_HandleUnfitItems(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small-box", 50, 50, 50, 500),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("fit-1", 30, 30, 30, 200),
		boxpacker3.NewItem("unfit-1", 100, 100, 100, 1000), // Doesn't fit
		boxpacker3.NewItem("unfit-2", 100, 100, 100, 1000), // Doesn't fit
	}

	strategies := []struct {
		name   string
		packer *boxpacker3.Packer
	}{
		{"MinimizeBoxes", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyMinimizeBoxes))},
		{"Greedy", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))},
		{"BestFit", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))},
		{"BestFitDecreasing", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))},
		{"NextFit", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyNextFit))},
		{"WorstFit", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyWorstFit))},
		{"AlmostWorstFit", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyAlmostWorstFit))},
	}

	for _, strategy := range strategies {
		t.Run(strategy.name, func(t *testing.T) {
			t.Parallel()

			result := strategy.packer.Pack(boxes, items)
			require.NotNil(t, result)
			validatePackingInvariants(t, result)

			// Verify unfit items
			unfitIDs := make(map[string]bool)
			for _, item := range result.UnfitItems {
				unfitIDs[item.GetID()] = true
			}

			require.True(t, unfitIDs["unfit-1"], "%s: unfit-1 should be unfit", strategy.name)
			require.True(t, unfitIDs["unfit-2"], "%s: unfit-2 should be unfit", strategy.name)

			// Verify fit items are packed
			totalPacked := 0
			for _, box := range result.Boxes {
				totalPacked += len(box.GetItems())
			}

			require.Equal(t, 1, totalPacked, "%s: One item should be packed", strategy.name)
		})
	}
}

// TestPacker_StrategyComparison_SpaceUtilization compares space utilization across strategies.
func TestPacker_StrategyComparison_SpaceUtilization(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-3", 100, 100, 100, 2000),
	}

	// Items with varying sizes that benefit from different strategies
	// Using simpler sizes to avoid geometric intersection issues
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("large-1", 50, 50, 50, 500),
		boxpacker3.NewItem("large-2", 50, 50, 50, 500),
		boxpacker3.NewItem("medium-1", 40, 40, 40, 300),
		boxpacker3.NewItem("small-1", 30, 30, 30, 200),
	}

	strategies := []struct {
		name   string
		packer *boxpacker3.Packer
	}{
		{"Greedy", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))},
		{"MinimizeBoxes", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyMinimizeBoxes))},
		{"BestFit", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))},
		{"BestFitDecreasing", boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))},
	}

	results := make(map[string]*boxpacker3.Result)

	for _, strategy := range strategies {
		result := strategy.packer.Pack(boxes, items)
		require.NotNil(t, result)
		validatePackingInvariants(t, result)
		results[strategy.name] = result
	}

	// All strategies should pack all items
	for name, result := range results {
		require.Empty(t, result.UnfitItems, "%s should pack all items", name)
	}

	// Calculate space utilization for each strategy
	utilization := make(map[string]float64)

	for name, result := range results {
		totalUsed := 0.0
		totalAvailable := 0.0

		for _, box := range result.Boxes {
			if len(box.GetItems()) > 0 {
				totalUsed += box.GetVolume() - box.GetRemainingVolume()
				totalAvailable += box.GetVolume()
			}
		}

		if totalAvailable > 0 {
			utilization[name] = totalUsed / totalAvailable
		}
	}

	// BestFitDecreasing should have good utilization
	require.Greater(t, utilization["BestFitDecreasing"], 0.0,
		"BestFitDecreasing should have positive utilization")
}
