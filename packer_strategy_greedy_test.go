package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestPacker_StrategyGreedy_ItemsSortedAscending(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)

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

	firstBoxItems := result.Boxes[0].Items
	if len(firstBoxItems) > 0 {
		firstItemVolume := firstBoxItems[0].Volume()

		require.LessOrEqual(t, firstItemVolume, 27000.0,
			"First item should be small (Greedy processes smallest first)")
	}
}

func TestPacker_StrategyGreedy_FirstFitBehavior(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 2000),
		boxpacker3.NewBox("box-3", 100, 100, 100, 2000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 30, 30, 30, 100),
		boxpacker3.NewItem("item-2", 30, 30, 30, 100),
		boxpacker3.NewItem("item-3", 30, 30, 30, 100),
	}

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	require.Empty(t, result.Unpacked, "All items should fit")
}

func TestPacker_StrategyGreedy_EmptyBoxes(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)

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

func TestPacker_StrategyGreedy_EmptyItems(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)

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

func TestPacker_StrategyGreedy_UnfitItems(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)

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

func TestPacker_StrategyGreedy_WeightConstraints(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)

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

func TestPacker_StrategyGreedy_GeometricPlacement(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)

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

func TestPacker_StrategyGreedy_ComplexScenario(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("large-1", 200, 200, 200, 5000),
		boxpacker3.NewBox("large-2", 200, 200, 200, 5000),
		boxpacker3.NewBox("medium-1", 150, 150, 150, 3000),
		boxpacker3.NewBox("medium-2", 150, 150, 150, 3000),
		boxpacker3.NewBox("small-1", 100, 100, 100, 2000),
		boxpacker3.NewBox("small-2", 100, 100, 100, 2000),
	}

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
