package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func strategyName(strategy boxpacker3.RuleSettings) string {
	return strategy.Name()
}

//nolint:funlen
func TestPacker_Property_AllItemsAccountedFor(t *testing.T) {
	t.Parallel()

	strategies := []boxpacker3.RuleSettings{
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectNextFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectWorstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectAlmostWorstFit},
	}

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
				boxpacker3.NewItem("unfit", 200, 200, 200, 5000),
			},
		},
	}

	for _, strategy := range strategies {
		for _, tc := range testCases {
			strategyName := strategyName(strategy)
			t.Run(strategyName+"_"+tc.name, func(t *testing.T) {
				t.Parallel()

				packer := rulePacker(strategy.Order, strategy.Selection)
				result, err := packer.Pack(context.Background(), tc.boxes, tc.items)
				require.NoError(t, err)
				require.NotNil(t, result)

				totalPacked := 0
				for _, box := range result.Boxes {
					totalPacked += len(box.Items)
				}

				require.Equal(t, len(tc.items), totalPacked+len(result.Unpacked),
					"All items must be either packed or in UnfitItems")
			})
		}
	}
}

func TestPacker_Property_NoDuplicates(t *testing.T) {
	t.Parallel()

	strategies := []boxpacker3.RuleSettings{
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectNextFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectWorstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectAlmostWorstFit},
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

			packer := rulePacker(strategy.Order, strategy.Selection)
			result, err := packer.Pack(context.Background(), boxes, items)
			require.NoError(t, err)
			require.NotNil(t, result)

			itemCounts := make(map[string]int)

			for _, box := range result.Boxes {
				for _, item := range box.Items {
					itemCounts[item.ID()]++
				}
			}

			for _, item := range result.Unpacked {
				itemCounts[item.ID()]++
			}

			for _, item := range items {
				require.Equal(t, 1, itemCounts[item.ID()],
					"Item %s should appear exactly once", item.ID())
			}
		})
	}
}

func TestPacker_Property_NoIntersections(t *testing.T) {
	t.Parallel()

	strategies := []boxpacker3.RuleSettings{
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectNextFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectWorstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectAlmostWorstFit},
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

			packer := rulePacker(strategy.Order, strategy.Selection)
			result, err := packer.Pack(context.Background(), boxes, items)
			require.NoError(t, err)
			require.NotNil(t, result)

			validatePackingInvariants(t, result)
		})
	}
}

func TestPacker_Property_WeightConstraints(t *testing.T) {
	t.Parallel()

	strategies := []boxpacker3.RuleSettings{
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectNextFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectWorstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectAlmostWorstFit},
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

			packer := rulePacker(strategy.Order, strategy.Selection)
			result, err := packer.Pack(context.Background(), boxes, items)
			require.NoError(t, err)
			require.NotNil(t, result)

			for _, box := range result.Boxes {
				totalWeight := 0.0
				for _, item := range box.Items {
					totalWeight += item.Weight()
				}

				require.LessOrEqual(t, totalWeight, box.Box.MaxWeight(),
					"Box %s should respect weight constraint", box.ID())
			}
		})
	}
}

func validatePackingInvariants(t *testing.T, result *boxpacker3.Result) {
	t.Helper()

	for _, box := range result.Boxes {
		requireNoOverlap(t, box)
		requireWithinCapacity(t, box)

		for _, item := range box.Items {
			requireWithinBox(t, box, item)
		}
	}
}
