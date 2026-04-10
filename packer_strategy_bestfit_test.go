package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestPacker_StrategyBestFit_ItemsSortedAscending(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 200, 200, 200, 5000),
		boxpacker3.NewBox("box-2", 200, 200, 200, 5000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("large", 40, 40, 40, 200),
		boxpacker3.NewItem("small", 20, 20, 20, 100),
		boxpacker3.NewItem("medium", 30, 30, 30, 150),
	}

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	require.Empty(t, result.Unpacked, "All items should fit")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.Items)
	}

	require.Equal(t, len(items), totalPacked, "All items should be packed")
}

func TestPacker_StrategyBestFit_SelectsBestBox(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("huge-box", 300, 300, 300, 10000),
		boxpacker3.NewBox("large-box", 200, 200, 200, 5000),
		boxpacker3.NewBox("medium-box", 150, 150, 150, 3000),
		boxpacker3.NewBox("small-box", 100, 100, 100, 2000),
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

	require.Equal(t, "small-box", boxWithItem.ID(),
		"BestFit should select box with smallest remaining space")
}

func TestPacker_StrategyBestFit_MultipleItemsBestFit(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-3", 100, 100, 100, 2000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("large-1", 50, 50, 50, 500),
		boxpacker3.NewItem("large-2", 50, 50, 50, 500),
		boxpacker3.NewItem("small-1", 30, 30, 30, 200),
		boxpacker3.NewItem("small-2", 30, 30, 30, 200),
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

	boxesWithItems := usedBoxCountOf(result)

	require.GreaterOrEqual(t, boxesWithItems, 1, "At least one box should have items")
}

func TestPacker_StrategyBestFit_EmptyBoxes(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)

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

func TestPacker_StrategyBestFit_EmptyItems(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
	}
	items := []*boxpacker3.Item{}

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Empty(t, result.Unpacked, "Should have no unfit items")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.Items)
	}

	require.Equal(t, 0, totalPacked, "Should have no packed items")
}

func TestPacker_StrategyBestFit_UnfitItems(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)

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

	unfitIDs := make(map[string]bool)
	for _, item := range result.Unpacked {
		unfitIDs[item.ID()] = true
	}

	require.True(t, unfitIDs["unfit-item"], "unfit-item should be in UnfitItems")

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.Items)
	}

	require.Equal(t, 1, totalPacked, "One item should be packed")
}

func TestPacker_StrategyBestFit_WeightConstraints(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("light-box", 100, 100, 100, 500),
		boxpacker3.NewBox("heavy-box", 100, 100, 100, 2000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("heavy-1", 30, 30, 30, 400),
		boxpacker3.NewItem("heavy-2", 30, 30, 30, 400),
		boxpacker3.NewItem("light-1", 20, 20, 20, 50),
	}

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	for _, box := range result.Boxes {
		totalWeight := 0.0
		for _, item := range box.Items {
			totalWeight += item.Weight()
		}

		require.LessOrEqual(t, totalWeight, box.Box.MaxWeight(),
			"Box %s should respect weight constraint", box.ID())
	}
}

func TestPacker_StrategyBestFit_GeometricPlacement(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 50, 50, 500),
		boxpacker3.NewItem("item-2", 50, 50, 50, 500),
	}

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	for _, box := range result.Boxes {
		boxItems := box.Items
		for idx := range boxItems {
			for j := idx + 1; j < len(boxItems); j++ {
				require.False(t, itemsOverlap(boxItems[idx], boxItems[j]),
					"Items should not intersect")
			}
		}
	}
}

//nolint:funlen
func TestPacker_StrategyBestFit_RemainingVolumeOptimization(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 2000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("large", 70, 70, 70, 700),
		boxpacker3.NewItem("small", 25, 25, 25, 200),
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

	var box1, box2 *boxpacker3.PackedBox

	for _, box := range result.Boxes {
		if box.ID() == "box-1" {
			box1 = &box
		} else if box.ID() == "box-2" {
			box2 = &box
		}
	}

	require.NotNil(t, box1, "box-1 should be used")

	box1Remaining := box1.RemainingVolume()

	var box2Remaining float64
	if box2 != nil {
		box2Remaining = box2.RemainingVolume()
	}

	box1Items := box1.Items
	smallInBox1 := false

	for _, item := range box1Items {
		if item.ID() == "small" {
			smallInBox1 = true

			break
		}
	}

	if smallInBox1 {
		if box2 != nil {
			require.Less(t, box1Remaining, box2Remaining,
				"If small item is in box-1, box-1 should have less remaining volume than box-2")
		}
	} else {
		require.NotNil(t, box2, "box-2 should exist if small is there")
		require.Less(t, box2Remaining, box1Remaining,
			"If small item is in box-2, box-2 should have less remaining volume than box-1")
	}
}
