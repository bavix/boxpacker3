package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestPacker_StrategyNextFit_NextBoxBehavior(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectNextFit)

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

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.Items)
	}

	require.Equal(t, len(items), totalPacked, "All items should be packed")
	require.Empty(t, result.Unpacked, "No items should be unfit")
}

func TestPacker_StrategyNextFit_EmptyBoxes(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectNextFit)

	boxes := []*boxpacker3.Box{}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
	}

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Empty(t, result.Boxes, "Should have no boxes")
	require.Len(t, result.Unpacked, len(items), "All items should be unfit")
}

func TestPacker_StrategyNextFit_UnfitItems(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectNextFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small-box", 50, 50, 50, 500),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("fit-item", 30, 30, 30, 200),
		boxpacker3.NewItem("unfit-item", 100, 100, 100, 1000),
	}

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	require.Len(t, result.Unpacked, 1, "One item should be unfit")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.Items)
	}

	require.Equal(t, 1, totalPacked, "One item should be packed")
}

func TestPacker_StrategyWorstFit_SelectsWorstBox(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectWorstFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small-box", 100, 100, 100, 2000),
		boxpacker3.NewBox("medium-box", 150, 150, 150, 3000),
		boxpacker3.NewBox("large-box", 200, 200, 200, 5000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
	}

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	require.Empty(t, result.Unpacked, "Item should fit")

	var boxWithItem *boxpacker3.PackedBox

	for _, box := range result.Boxes {
		if len(box.Items) > 0 {
			boxWithItem = &box

			break
		}
	}

	require.NotNil(t, boxWithItem, "Item should be in a box")

	require.Equal(t, "large-box", boxWithItem.ID(),
		"WorstFit should select box with largest remaining space")
}

func TestPacker_StrategyWorstFit_EmptyBoxes(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectWorstFit)

	boxes := []*boxpacker3.Box{}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
	}

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Empty(t, result.Boxes, "Should have no boxes")
	require.Len(t, result.Unpacked, len(items), "All items should be unfit")
}

func TestPacker_StrategyWorstFit_UnfitItems(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectWorstFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small-box", 50, 50, 50, 500),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("fit-item", 30, 30, 30, 200),
		boxpacker3.NewItem("unfit-item", 100, 100, 100, 1000),
	}

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	require.Len(t, result.Unpacked, 1, "One item should be unfit")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.Items)
	}

	require.Equal(t, 1, totalPacked, "One item should be packed")
}

func TestPacker_StrategyAlmostWorstFit_EmptyBoxes(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectAlmostWorstFit)

	boxes := []*boxpacker3.Box{}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
	}

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Empty(t, result.Boxes, "Should have no boxes")
	require.Len(t, result.Unpacked, len(items), "All items should be unfit")
}

func TestPacker_StrategyAlmostWorstFit_UnfitItems(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectAlmostWorstFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small-box", 50, 50, 50, 500),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("fit-item", 30, 30, 30, 200),
		boxpacker3.NewItem("unfit-item", 100, 100, 100, 1000),
	}

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	require.Len(t, result.Unpacked, 1, "One item should be unfit")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.Items)
	}

	require.Equal(t, 1, totalPacked, "One item should be packed")
}

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
