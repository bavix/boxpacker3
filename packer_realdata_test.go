package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestPacker_RealWorld_ECommerceOrder(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small", 30, 20, 15, 5000),
		boxpacker3.NewBox("medium", 40, 30, 25, 10000),
		boxpacker3.NewBox("large", 60, 40, 35, 20000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("book-1", 20, 15, 3, 500),
		boxpacker3.NewItem("book-2", 20, 15, 3, 500),
		boxpacker3.NewItem("book-3", 20, 15, 3, 500),
		boxpacker3.NewItem("t-shirt", 30, 25, 2, 200),
		boxpacker3.NewItem("mug", 12, 12, 15, 300),
		boxpacker3.NewItem("notebook", 25, 20, 2, 400),
		boxpacker3.NewItem("phone-case", 15, 8, 1, 50),
		boxpacker3.NewItem("charger", 10, 5, 2, 100),
	}

	strategies := []boxpacker3.RuleSettings{
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectBestFit},
	}

	for _, strategy := range strategies {
		t.Run(strategyName(strategy), func(t *testing.T) {
			t.Parallel()

			packer := rulePacker(strategy.Order, strategy.Selection)
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

			boxesUsed := usedBoxCountOf(result)

			require.LessOrEqual(t, boxesUsed, 2, "Should use at most 2 boxes for this order")
		})
	}
}

func TestPacker_RealWorld_PackriftEcommerceCartonFixture(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("packrift-40x20x20-1", 40, 20, 20, 75),
		boxpacker3.NewBox("packrift-40x20x20-2", 40, 20, 20, 75),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("demo-small-item", 7.5, 4.5, 3.5, 1),
		boxpacker3.NewItem("demo-flat-item-1", 15, 7, 2.5, 1.5),
		boxpacker3.NewItem("demo-flat-item-2", 15, 7, 2.5, 1.5),
		boxpacker3.NewItem("demo-long-item", 21, 8.5, 5, 3),
		boxpacker3.NewItem("demo-bulk-item-1", 18, 12, 5.5, 4),
		boxpacker3.NewItem("demo-bulk-item-2", 18, 12, 5.5, 4),
		boxpacker3.NewItem("demo-bulk-item-3", 18, 12, 5.5, 4),
		boxpacker3.NewItem("demo-bulk-item-4", 18, 12, 5.5, 4),
	}

	packer := rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectBestFit)
	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)

	require.NotNil(t, result)
	validatePackingInvariants(t, result)
	require.Empty(t, result.Unpacked, "No Packrift fixture items should be unfit")

	totalPacked := 0
	boxesUsed := 0

	for _, box := range result.Boxes {
		if len(box.Items) > 0 {
			boxesUsed++
			totalPacked += len(box.Items)
		}
	}

	require.Equal(t, len(items), totalPacked, "All Packrift fixture items should be packed")
	require.LessOrEqual(t, boxesUsed, 2, "Fixture should fit in at most two cartons")
}

func TestPacker_RealWorld_WarehousePacking(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("pallet-small", 100, 100, 100, 50000),
		boxpacker3.NewBox("pallet-medium", 120, 120, 120, 100000),
		boxpacker3.NewBox("pallet-large", 150, 150, 150, 200000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("box-small-1", 30, 30, 30, 5000),
		boxpacker3.NewItem("box-small-2", 30, 30, 30, 5000),
		boxpacker3.NewItem("box-small-3", 30, 30, 30, 5000),
		boxpacker3.NewItem("box-medium-1", 50, 50, 50, 10000),
		boxpacker3.NewItem("box-medium-2", 50, 50, 50, 10000),
		boxpacker3.NewItem("box-large-1", 70, 70, 70, 20000),
		boxpacker3.NewItem("box-large-2", 70, 70, 70, 20000),
	}

	packer := rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectBestFit)
	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)

	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.Items)
	}

	require.Equal(t, len(items), totalPacked, "All items should be packed")

	totalVolume := 0.0
	usedVolume := 0.0

	for _, box := range result.Boxes {
		if len(box.Items) > 0 {
			totalVolume += box.Volume()
			usedVolume += box.Volume() - box.RemainingVolume()
		}
	}

	if totalVolume > 0 {
		utilization := usedVolume / totalVolume
		require.Greater(t, utilization, 0.3, "Space utilization should be at least 30%%")
	}
}

func TestPacker_RealWorld_MovingBoxes(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small-moving", 40, 30, 30, 15000),
		boxpacker3.NewBox("medium-moving", 50, 40, 40, 25000),
		boxpacker3.NewBox("large-moving", 60, 50, 50, 40000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("vase", 20, 20, 30, 2000),
		boxpacker3.NewItem("books", 30, 25, 20, 5000),
		boxpacker3.NewItem("plates", 35, 35, 10, 3000),
		boxpacker3.NewItem("cups", 25, 25, 15, 1500),
		boxpacker3.NewItem("small-decor", 15, 15, 15, 500),
		boxpacker3.NewItem("photo-frame", 25, 20, 3, 300),
	}

	packer := rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit)
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

func TestPacker_RealWorld_RetailStore(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("display-small", 50, 40, 30, 10000),
		boxpacker3.NewBox("display-medium", 70, 50, 40, 20000),
		boxpacker3.NewBox("display-large", 100, 70, 50, 30000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("product-1", 10, 10, 10, 500),
		boxpacker3.NewItem("product-2", 10, 10, 10, 500),
		boxpacker3.NewItem("product-3", 15, 15, 15, 800),
		boxpacker3.NewItem("product-4", 15, 15, 15, 800),
		boxpacker3.NewItem("product-5", 20, 20, 20, 1200),
		boxpacker3.NewItem("product-6", 20, 20, 20, 1200),
		boxpacker3.NewItem("product-7", 25, 25, 25, 2000),
		boxpacker3.NewItem("product-8", 25, 25, 25, 2000),
		boxpacker3.NewItem("product-9", 30, 30, 30, 3000),
	}

	strategies := []boxpacker3.RuleSettings{
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectFirstFit},
	}

	for _, strategy := range strategies {
		t.Run(strategyName(strategy), func(t *testing.T) {
			t.Parallel()

			packer := rulePacker(strategy.Order, strategy.Selection)
			result, err := packer.Pack(context.Background(), boxes, items)
			require.NoError(t, err)

			require.NotNil(t, result)
			validatePackingInvariants(t, result)

			totalPacked := 0
			for _, box := range result.Boxes {
				totalPacked += len(box.Items)
			}

			require.Equal(t, len(items), totalPacked, "All items should be packed")
		})
	}
}

func TestPacker_RealWorld_MixedConstraints(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("light-box", 100, 100, 100, 5000),
		boxpacker3.NewBox("medium-box", 100, 100, 100, 15000),
		boxpacker3.NewBox("heavy-box", 100, 100, 100, 30000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("light-1", 20, 20, 20, 1000),
		boxpacker3.NewItem("light-2", 20, 20, 20, 1000),
		boxpacker3.NewItem("light-3", 20, 20, 20, 1000),
		boxpacker3.NewItem("medium-1", 30, 30, 30, 5000),
		boxpacker3.NewItem("medium-2", 30, 30, 30, 5000),
		boxpacker3.NewItem("heavy-1", 40, 40, 40, 10000),
	}

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)
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

	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.Items)
	}

	require.Equal(t, len(items), totalPacked, "All items should be packed")
}
