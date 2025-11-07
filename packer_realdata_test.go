package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

// TestPacker_RealWorld_ECommerceOrder tests a typical e-commerce order scenario.
func TestPacker_RealWorld_ECommerceOrder(t *testing.T) {
	t.Parallel()

	// Typical shipping boxes
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small", 30, 20, 15, 5000),   // Small box: 30x20x15 cm, 5kg
		boxpacker3.NewBox("medium", 40, 30, 25, 10000), // Medium box: 40x30x25 cm, 10kg
		boxpacker3.NewBox("large", 60, 40, 35, 20000),  // Large box: 60x40x35 cm, 20kg
	}

	// Typical e-commerce items
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("book-1", 20, 15, 3, 500),   // Book: 20x15x3 cm, 0.5kg
		boxpacker3.NewItem("book-2", 20, 15, 3, 500),   // Book: 20x15x3 cm, 0.5kg
		boxpacker3.NewItem("book-3", 20, 15, 3, 500),   // Book: 20x15x3 cm, 0.5kg
		boxpacker3.NewItem("t-shirt", 30, 25, 2, 200),  // T-shirt: 30x25x2 cm, 0.2kg
		boxpacker3.NewItem("mug", 12, 12, 15, 300),     // Mug: 12x12x15 cm, 0.3kg
		boxpacker3.NewItem("notebook", 25, 20, 2, 400), // Notebook: 25x20x2 cm, 0.4kg
		boxpacker3.NewItem("phone-case", 15, 8, 1, 50), // Phone case: 15x8x1 cm, 0.05kg
		boxpacker3.NewItem("charger", 10, 5, 2, 100),   // Charger: 10x5x2 cm, 0.1kg
	}

	strategies := []boxpacker3.PackingStrategy{
		boxpacker3.StrategyBestFit,
		boxpacker3.StrategyBestFitDecreasing,
	}

	for _, strategy := range strategies {
		t.Run(strategyName(strategy), func(t *testing.T) {
			t.Parallel()

			packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(strategy))
			result, err := packer.PackCtx(context.Background(), boxes, items)
			require.NoError(t, err)

			require.NotNil(t, result)
			validatePackingInvariants(t, result)

			// All items should be packed (they all fit)
			totalPacked := 0
			for _, box := range result.Boxes {
				totalPacked += len(box.GetItems())
			}

			require.Equal(t, len(items), totalPacked, "All items should be packed")
			require.Empty(t, result.UnfitItems, "No items should be unfit")

			// Verify reasonable box usage (should use 1-2 boxes)
			boxesUsed := 0

			for _, box := range result.Boxes {
				if len(box.GetItems()) > 0 {
					boxesUsed++
				}
			}

			require.LessOrEqual(t, boxesUsed, 2, "Should use at most 2 boxes for this order")
		})
	}
}

// TestPacker_RealWorld_WarehousePacking tests a warehouse packing scenario.
func TestPacker_RealWorld_WarehousePacking(t *testing.T) {
	t.Parallel()

	// Standard warehouse boxes
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("pallet-small", 100, 100, 100, 50000),   // Small pallet: 1x1x1m, 50kg
		boxpacker3.NewBox("pallet-medium", 120, 120, 120, 100000), // Medium pallet: 1.2x1.2x1.2m, 100kg
		boxpacker3.NewBox("pallet-large", 150, 150, 150, 200000),  // Large pallet: 1.5x1.5x1.5m, 200kg
	}

	// Warehouse items of various sizes
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("box-small-1", 30, 30, 30, 5000),   // Small box: 30x30x30 cm, 5kg
		boxpacker3.NewItem("box-small-2", 30, 30, 30, 5000),   // Small box: 30x30x30 cm, 5kg
		boxpacker3.NewItem("box-small-3", 30, 30, 30, 5000),   // Small box: 30x30x30 cm, 5kg
		boxpacker3.NewItem("box-medium-1", 50, 50, 50, 10000), // Medium box: 50x50x50 cm, 10kg
		boxpacker3.NewItem("box-medium-2", 50, 50, 50, 10000), // Medium box: 50x50x50 cm, 10kg
		boxpacker3.NewItem("box-large-1", 70, 70, 70, 20000),  // Large box: 70x70x70 cm, 20kg
		boxpacker3.NewItem("box-large-2", 70, 70, 70, 20000),  // Large box: 70x70x70 cm, 20kg
	}

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))
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

	// Calculate space utilization
	totalVolume := 0.0
	usedVolume := 0.0

	for _, box := range result.Boxes {
		if len(box.GetItems()) > 0 {
			totalVolume += box.GetVolume()
			usedVolume += box.GetVolume() - box.GetRemainingVolume()
		}
	}

	if totalVolume > 0 {
		utilization := usedVolume / totalVolume
		require.Greater(t, utilization, 0.3, "Space utilization should be at least 30%%")
	}
}

// TestPacker_RealWorld_MovingBoxes tests a moving/relocation scenario.
func TestPacker_RealWorld_MovingBoxes(t *testing.T) {
	t.Parallel()

	// Moving boxes
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small-moving", 40, 30, 30, 15000),  // Small: 40x30x30 cm, 15kg
		boxpacker3.NewBox("medium-moving", 50, 40, 40, 25000), // Medium: 50x40x40 cm, 25kg
		boxpacker3.NewBox("large-moving", 60, 50, 50, 40000),  // Large: 60x50x50 cm, 40kg
	}

	// Household items
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("vase", 20, 20, 30, 2000),       // Vase: 20x20x30 cm, 2kg
		boxpacker3.NewItem("books", 30, 25, 20, 5000),      // Books: 30x25x20 cm, 5kg
		boxpacker3.NewItem("plates", 35, 35, 10, 3000),     // Plates: 35x35x10 cm, 3kg
		boxpacker3.NewItem("cups", 25, 25, 15, 1500),       // Cups: 25x25x15 cm, 1.5kg
		boxpacker3.NewItem("small-decor", 15, 15, 15, 500), // Small decor: 15x15x15 cm, 0.5kg
		boxpacker3.NewItem("photo-frame", 25, 20, 3, 300),  // Photo frame: 25x20x3 cm, 0.3kg
	}

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyMinimizeBoxes))
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

// TestPacker_RealWorld_RetailStore tests a retail store inventory scenario.
func TestPacker_RealWorld_RetailStore(t *testing.T) {
	t.Parallel()

	// Store display boxes
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("display-small", 50, 40, 30, 10000),  // Small display: 50x40x30 cm, 10kg
		boxpacker3.NewBox("display-medium", 70, 50, 40, 20000), // Medium display: 70x50x40 cm, 20kg
		boxpacker3.NewBox("display-large", 100, 70, 50, 30000), // Large display: 100x70x50 cm, 30kg
	}

	// Retail products
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("product-1", 10, 10, 10, 500),  // Product: 10x10x10 cm, 0.5kg
		boxpacker3.NewItem("product-2", 10, 10, 10, 500),  // Product: 10x10x10 cm, 0.5kg
		boxpacker3.NewItem("product-3", 15, 15, 15, 800),  // Product: 15x15x15 cm, 0.8kg
		boxpacker3.NewItem("product-4", 15, 15, 15, 800),  // Product: 15x15x15 cm, 0.8kg
		boxpacker3.NewItem("product-5", 20, 20, 20, 1200), // Product: 20x20x20 cm, 1.2kg
		boxpacker3.NewItem("product-6", 20, 20, 20, 1200), // Product: 20x20x20 cm, 1.2kg
		boxpacker3.NewItem("product-7", 25, 25, 25, 2000), // Product: 25x25x25 cm, 2kg
		boxpacker3.NewItem("product-8", 25, 25, 25, 2000), // Product: 25x25x25 cm, 2kg
		boxpacker3.NewItem("product-9", 30, 30, 30, 3000), // Product: 30x30x30 cm, 3kg
	}

	strategies := []boxpacker3.PackingStrategy{
		boxpacker3.StrategyBestFitDecreasing,
		boxpacker3.StrategyMinimizeBoxes,
	}

	for _, strategy := range strategies {
		t.Run(strategyName(strategy), func(t *testing.T) {
			t.Parallel()

			packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(strategy))
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
		})
	}
}

// TestPacker_RealWorld_MixedConstraints tests a scenario with mixed constraints.
func TestPacker_RealWorld_MixedConstraints(t *testing.T) {
	t.Parallel()

	// Boxes with different weight limits
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("light-box", 100, 100, 100, 5000),   // Light box: 100x100x100 cm, 5kg max
		boxpacker3.NewBox("medium-box", 100, 100, 100, 15000), // Medium box: 100x100x100 cm, 15kg max
		boxpacker3.NewBox("heavy-box", 100, 100, 100, 30000),  // Heavy box: 100x100x100 cm, 30kg max
	}

	// Items with varying weights
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("light-1", 20, 20, 20, 1000),  // Light: 20x20x20 cm, 1kg
		boxpacker3.NewItem("light-2", 20, 20, 20, 1000),  // Light: 20x20x20 cm, 1kg
		boxpacker3.NewItem("light-3", 20, 20, 20, 1000),  // Light: 20x20x20 cm, 1kg
		boxpacker3.NewItem("medium-1", 30, 30, 30, 5000), // Medium: 30x30x30 cm, 5kg
		boxpacker3.NewItem("medium-2", 30, 30, 30, 5000), // Medium: 30x30x30 cm, 5kg
		boxpacker3.NewItem("heavy-1", 40, 40, 40, 10000), // Heavy: 40x40x40 cm, 10kg
	}

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))
	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)

	require.NotNil(t, result)
	validatePackingInvariants(t, result)

	// Verify weight constraints are respected
	for _, box := range result.Boxes {
		totalWeight := 0.0
		for _, item := range box.GetItems() {
			totalWeight += item.GetWeight()
		}

		require.LessOrEqual(t, totalWeight, box.GetMaxWeight(),
			"Box %s should respect weight constraint", box.GetID())
	}

	// All items should be packed (they all fit)
	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.Equal(t, len(items), totalPacked, "All items should be packed")
}
