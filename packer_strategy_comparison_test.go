package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

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
		{nameFirstFitDecreasing, rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit)},
		{nameFirstFitIncreasing, rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)},
		{nameBestFitIncreasing, rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)},
		{nameBestFitDecreasing, rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectBestFit)},
		{nameNextFitIncreasing, rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectNextFit)},
		{nameWorstFitIncreasing, rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectWorstFit)},
		{nameAlmostWorstFitIncreasing, rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectAlmostWorstFit)},
	}

	for _, strategy := range strategies {
		t.Run(strategy.name, func(t *testing.T) {
			t.Parallel()

			result, err := strategy.packer.Pack(context.Background(), boxes, items)
			require.NoError(t, err)
			require.NotNil(t, result)

			validatePackingInvariants(t, result)

			totalPacked := 0
			for _, box := range result.Boxes {
				totalPacked += len(box.Items)
			}

			require.Equal(t, len(items), totalPacked+len(result.Unpacked),
				"%s: All items must be either packed or in UnfitItems", strategy.name)
		})
	}
}

func TestPacker_AllStrategies_HandleUnfitItems(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small-box", 50, 50, 50, 500),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("fit-1", 30, 30, 30, 200),
		boxpacker3.NewItem("unfit-1", 100, 100, 100, 1000),
		boxpacker3.NewItem("unfit-2", 100, 100, 100, 1000),
	}

	strategies := []struct {
		name   string
		packer *boxpacker3.Packer
	}{
		{nameFirstFitDecreasing, rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit)},
		{nameFirstFitIncreasing, rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)},
		{nameBestFitIncreasing, rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)},
		{nameBestFitDecreasing, rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectBestFit)},
		{nameNextFitIncreasing, rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectNextFit)},
		{nameWorstFitIncreasing, rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectWorstFit)},
		{nameAlmostWorstFitIncreasing, rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectAlmostWorstFit)},
	}

	for _, strategy := range strategies {
		t.Run(strategy.name, func(t *testing.T) {
			t.Parallel()

			result, err := strategy.packer.Pack(context.Background(), boxes, items)
			require.NoError(t, err)
			require.NotNil(t, result)
			validatePackingInvariants(t, result)

			unfitIDs := make(map[string]bool)
			for _, item := range result.Unpacked {
				unfitIDs[item.ID()] = true
			}

			require.True(t, unfitIDs["unfit-1"], "%s: unfit-1 should be unfit", strategy.name)
			require.True(t, unfitIDs["unfit-2"], "%s: unfit-2 should be unfit", strategy.name)

			totalPacked := 0
			for _, box := range result.Boxes {
				totalPacked += len(box.Items)
			}

			require.Equal(t, 1, totalPacked, "%s: One item should be packed", strategy.name)
		})
	}
}

func TestPacker_StrategyComparison_SpaceUtilization(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-3", 100, 100, 100, 2000),
	}

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
		{nameFirstFitIncreasing, rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)},
		{nameFirstFitDecreasing, rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit)},
		{nameBestFitIncreasing, rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)},
		{nameBestFitDecreasing, rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectBestFit)},
	}

	results := make(map[string]*boxpacker3.Result)

	for _, strategy := range strategies {
		result, err := strategy.packer.Pack(context.Background(), boxes, items)
		require.NoError(t, err)
		require.NotNil(t, result)
		validatePackingInvariants(t, result)
		results[strategy.name] = result
	}

	for name, result := range results {
		require.Empty(t, result.Unpacked, "%s should pack all items", name)
	}

	utilization := make(map[string]float64)

	for name, result := range results {
		totalUsed := 0.0
		totalAvailable := 0.0

		for _, box := range result.Boxes {
			if len(box.Items) > 0 {
				totalUsed += box.Volume() - box.RemainingVolume()
				totalAvailable += box.Volume()
			}
		}

		if totalAvailable > 0 {
			utilization[name] = totalUsed / totalAvailable
		}
	}

	require.Greater(t, utilization[nameBestFitDecreasing], 0.0,
		"BestFitDecreasing should have positive utilization")
}
