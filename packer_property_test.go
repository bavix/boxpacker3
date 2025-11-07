package boxpacker3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

// strategyName returns a string representation of the packing strategy.
func strategyName(strategy boxpacker3.PackingStrategy) string {
	switch strategy {
	case boxpacker3.StrategyMinimizeBoxes:
		return "MinimizeBoxes"
	case boxpacker3.StrategyGreedy:
		return "Greedy"
	case boxpacker3.StrategyBestFit:
		return "BestFit"
	case boxpacker3.StrategyBestFitDecreasing:
		return "BestFitDecreasing"
	case boxpacker3.StrategyNextFit:
		return "NextFit"
	case boxpacker3.StrategyWorstFit:
		return "WorstFit"
	case boxpacker3.StrategyAlmostWorstFit:
		return "AlmostWorstFit"
	default:
		return fmt.Sprintf("Unknown(%d)", int(strategy))
	}
}

// TestPacker_Property_AllItemsAccountedFor is a property-based test that verifies
// all items are either packed or marked as unfit.
//
//nolint:funlen
func TestPacker_Property_AllItemsAccountedFor(t *testing.T) {
	t.Parallel()

	// Test with different strategies
	strategies := []boxpacker3.PackingStrategy{
		boxpacker3.StrategyMinimizeBoxes,
		boxpacker3.StrategyGreedy,
		boxpacker3.StrategyBestFit,
		boxpacker3.StrategyBestFitDecreasing,
		boxpacker3.StrategyNextFit,
		boxpacker3.StrategyWorstFit,
		boxpacker3.StrategyAlmostWorstFit,
	}

	// Test with different scenarios
	testCases := []struct {
		name  string
		boxes []*boxpacker3.Box
		items []*boxpacker3.Item
	}{
		{
			name: "SingleBox",
			boxes: []*boxpacker3.Box{
				boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
			},
			items: []*boxpacker3.Item{
				boxpacker3.NewItem("item-1", 30, 30, 30, 100),
				boxpacker3.NewItem("item-2", 30, 30, 30, 100),
				boxpacker3.NewItem("item-3", 30, 30, 30, 100),
			},
		},
		{
			name: "MultipleBoxes",
			boxes: []*boxpacker3.Box{
				boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
				boxpacker3.NewBox("box-2", 100, 100, 100, 2000),
				boxpacker3.NewBox("box-3", 100, 100, 100, 2000),
			},
			items: []*boxpacker3.Item{
				boxpacker3.NewItem("item-1", 40, 40, 40, 200),
				boxpacker3.NewItem("item-2", 40, 40, 40, 200),
				boxpacker3.NewItem("item-3", 40, 40, 40, 200),
				boxpacker3.NewItem("item-4", 40, 40, 40, 200),
				boxpacker3.NewItem("item-5", 40, 40, 40, 200),
			},
		},
		{
			name: "MixedSizes",
			boxes: []*boxpacker3.Box{
				boxpacker3.NewBox("small", 50, 50, 50, 1000),
				boxpacker3.NewBox("medium", 100, 100, 100, 2000),
				boxpacker3.NewBox("large", 200, 200, 200, 5000),
			},
			items: []*boxpacker3.Item{
				boxpacker3.NewItem("tiny", 10, 10, 10, 50),
				boxpacker3.NewItem("small", 30, 30, 30, 200),
				boxpacker3.NewItem("medium", 60, 60, 60, 500),
				boxpacker3.NewItem("large", 80, 80, 80, 800),
			},
		},
		{
			name: "UnfitItems",
			boxes: []*boxpacker3.Box{
				boxpacker3.NewBox("small", 50, 50, 50, 1000),
			},
			items: []*boxpacker3.Item{
				boxpacker3.NewItem("fit", 30, 30, 30, 200),
				boxpacker3.NewItem("unfit", 200, 200, 200, 5000), // Too large
			},
		},
	}

	for _, strategy := range strategies {
		for _, tc := range testCases {
			strategyName := strategyName(strategy)
			t.Run(strategyName+"_"+tc.name, func(t *testing.T) {
				t.Parallel()

				packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(strategy))
				result, err := packer.PackCtx(context.Background(), tc.boxes, tc.items)
				require.NoError(t, err)
				require.NotNil(t, result)

				// Property: All items must be accounted for
				totalPacked := 0
				for _, box := range result.Boxes {
					totalPacked += len(box.GetItems())
				}

				require.Equal(t, len(tc.items), totalPacked+len(result.UnfitItems),
					"All items must be either packed or in UnfitItems")
			})
		}
	}
}

// TestPacker_Property_NoDuplicates is a property-based test that verifies
// no item is packed multiple times.
func TestPacker_Property_NoDuplicates(t *testing.T) {
	t.Parallel()

	strategies := []boxpacker3.PackingStrategy{
		boxpacker3.StrategyMinimizeBoxes,
		boxpacker3.StrategyGreedy,
		boxpacker3.StrategyBestFit,
		boxpacker3.StrategyBestFitDecreasing,
		boxpacker3.StrategyNextFit,
		boxpacker3.StrategyWorstFit,
		boxpacker3.StrategyAlmostWorstFit,
	}

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 2000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 30, 30, 30, 100),
		boxpacker3.NewItem("item-2", 30, 30, 30, 100),
		boxpacker3.NewItem("item-3", 30, 30, 30, 100),
		boxpacker3.NewItem("item-4", 30, 30, 30, 100),
		boxpacker3.NewItem("item-5", 30, 30, 30, 100),
	}

	for _, strategy := range strategies {
		strategyName := strategyName(strategy)
		t.Run(strategyName, func(t *testing.T) {
			t.Parallel()

			packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(strategy))
			result, err := packer.PackCtx(context.Background(), boxes, items)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Property: No item should appear multiple times
			itemCounts := make(map[string]int)

			// Count items in boxes
			for _, box := range result.Boxes {
				for _, item := range box.GetItems() {
					itemCounts[item.GetID()]++
				}
			}

			// Count items in UnfitItems
			for _, item := range result.UnfitItems {
				itemCounts[item.GetID()]++
			}

			// Each item should appear exactly once
			for _, item := range items {
				require.Equal(t, 1, itemCounts[item.GetID()],
					"Item %s should appear exactly once", item.GetID())
			}
		})
	}
}

// TestPacker_Property_NoIntersections is a property-based test that verifies
// no items intersect within the same box.
func TestPacker_Property_NoIntersections(t *testing.T) {
	t.Parallel()

	strategies := []boxpacker3.PackingStrategy{
		boxpacker3.StrategyMinimizeBoxes,
		boxpacker3.StrategyGreedy,
		boxpacker3.StrategyBestFit,
		boxpacker3.StrategyBestFitDecreasing,
		boxpacker3.StrategyNextFit,
		boxpacker3.StrategyWorstFit,
		boxpacker3.StrategyAlmostWorstFit,
	}

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 2000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 30, 30, 30, 100),
		boxpacker3.NewItem("item-2", 30, 30, 30, 100),
		boxpacker3.NewItem("item-3", 30, 30, 30, 100),
		boxpacker3.NewItem("item-4", 30, 30, 30, 100),
	}

	for _, strategy := range strategies {
		strategyName := strategyName(strategy)
		t.Run(strategyName, func(t *testing.T) {
			t.Parallel()

			packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(strategy))
			result, err := packer.PackCtx(context.Background(), boxes, items)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Property: No items should intersect within the same box
			for _, box := range result.Boxes {
				boxItems := box.GetItems()
				for i := range boxItems {
					for j := i + 1; j < len(boxItems); j++ {
						require.False(t, boxItems[i].Intersect(boxItems[j]),
							"Items %s and %s in box %s should not intersect",
							boxItems[i].GetID(), boxItems[j].GetID(), box.GetID())
					}
				}
			}
		})
	}
}

// TestPacker_Property_WeightConstraints is a property-based test that verifies
// weight constraints are respected.
func TestPacker_Property_WeightConstraints(t *testing.T) {
	t.Parallel()

	strategies := []boxpacker3.PackingStrategy{
		boxpacker3.StrategyMinimizeBoxes,
		boxpacker3.StrategyGreedy,
		boxpacker3.StrategyBestFit,
		boxpacker3.StrategyBestFitDecreasing,
		boxpacker3.StrategyNextFit,
		boxpacker3.StrategyWorstFit,
		boxpacker3.StrategyAlmostWorstFit,
	}

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("light", 100, 100, 100, 500),
		boxpacker3.NewBox("heavy", 100, 100, 100, 2000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 30, 30, 30, 200),
		boxpacker3.NewItem("item-2", 30, 30, 30, 200),
		boxpacker3.NewItem("item-3", 30, 30, 30, 200),
	}

	for _, strategy := range strategies {
		strategyName := strategyName(strategy)
		t.Run(strategyName, func(t *testing.T) {
			t.Parallel()

			packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(strategy))
			result, err := packer.PackCtx(context.Background(), boxes, items)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Property: Weight constraints must be respected
			for _, box := range result.Boxes {
				totalWeight := 0.0
				for _, item := range box.GetItems() {
					totalWeight += item.GetWeight()
				}

				require.LessOrEqual(t, totalWeight, box.GetMaxWeight(),
					"Box %s should respect weight constraint", box.GetID())
			}
		})
	}
}

// validatePackingInvariants verifies that packing invariants are maintained:
// - No items intersect within the same box
// - Weight constraints are respected
// - Volume constraints are respected
// - Items are within box boundaries.
func validatePackingInvariants(t *testing.T, result *boxpacker3.Result) {
	t.Helper()

	for _, box := range result.Boxes {
		boxItems := box.GetItems()

		// Check for intersections
		for i := range boxItems {
			for j := i + 1; j < len(boxItems); j++ {
				require.False(t, boxItems[i].Intersect(boxItems[j]),
					"Items %s and %s in box %s should not intersect",
					boxItems[i].GetID(), boxItems[j].GetID(), box.GetID())
			}
		}

		// Check weight constraint
		totalWeight := 0.0
		for _, item := range boxItems {
			totalWeight += item.GetWeight()
		}

		require.LessOrEqual(t, totalWeight, box.GetMaxWeight(),
			"Box %s should respect weight constraint", box.GetID())

		// Check volume constraint
		totalVolume := 0.0
		for _, item := range boxItems {
			totalVolume += item.GetVolume()
		}

		require.LessOrEqual(t, totalVolume, box.GetVolume(),
			"Box %s should respect volume constraint", box.GetID())

		// Check geometric constraints (items within box boundaries)
		for _, item := range boxItems {
			dim := item.GetDimension()
			pos := item.GetPosition()

			require.LessOrEqual(t, pos[0]+dim[0], box.GetWidth(),
				"Item %s width should be <= box width", item.GetID())
			require.LessOrEqual(t, pos[1]+dim[1], box.GetHeight(),
				"Item %s height should be <= box height", item.GetID())
			require.LessOrEqual(t, pos[2]+dim[2], box.GetDepth(),
				"Item %s depth should be <= box depth", item.GetID())
		}
	}
}
