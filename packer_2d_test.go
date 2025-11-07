package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

// Test2DPacking tests 2D packing functionality using NewBox2D and NewItem2D.
func Test2DPacking(t *testing.T) {
	t.Parallel()

	// Create a 2D box (100x50, depth=1)
	box := boxpacker3.NewBox2D("box-2d-1", 100, 50, 1000)

	// Verify box dimensions
	require.InDelta(t, 100.0, box.GetWidth(), 0.0001)
	require.InDelta(t, 50.0, box.GetHeight(), 0.0001)
	require.InDelta(t, 1.0, box.GetDepth(), 0.0001)
	require.InDelta(t, 5000.0, box.GetVolume(), 0.0001) // 100 * 50 * 1

	// Create 2D items (flat items with depth=1)
	items := []*boxpacker3.Item{
		boxpacker3.NewItem2D(uuid.New().String(), 30, 20, 100),
		boxpacker3.NewItem2D(uuid.New().String(), 30, 20, 100),
		boxpacker3.NewItem2D(uuid.New().String(), 40, 25, 100),
	}

	// Verify item dimensions
	for _, item := range items {
		require.InDelta(t, 1.0, item.GetDepth(), 0.0001, "Item depth should be 1 for 2D items")
	}

	// Create packer and pack items
	packer := boxpacker3.NewPacker()
	result, err := packer.PackCtx(context.Background(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)

	// Verify packing result
	require.NotNil(t, result)
	require.Len(t, result.Boxes, 1)

	// Check that items were packed
	packedItems := result.Boxes[0].GetItems()
	require.NotEmpty(t, packedItems, "At least some items should be packed")

	// Verify all packed items are 2D (depth = 1)
	for _, item := range packedItems {
		require.InDelta(t, 1.0, item.GetDepth(), 0.0001, "Packed item should have depth=1")
	}
}

// Test2DPackingMultipleBoxes tests 2D packing with multiple boxes.
func Test2DPackingMultipleBoxes(t *testing.T) {
	t.Parallel()

	// Create multiple 2D boxes
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox2D("box-2d-small", 50, 30, 500),
		boxpacker3.NewBox2D("box-2d-medium", 100, 50, 1000),
		boxpacker3.NewBox2D("box-2d-large", 200, 100, 2000),
	}

	// Create various 2D items
	items := []*boxpacker3.Item{
		boxpacker3.NewItem2D(uuid.New().String(), 20, 15, 50),
		boxpacker3.NewItem2D(uuid.New().String(), 25, 20, 50),
		boxpacker3.NewItem2D(uuid.New().String(), 30, 25, 50),
		boxpacker3.NewItem2D(uuid.New().String(), 40, 30, 50),
		boxpacker3.NewItem2D(uuid.New().String(), 50, 40, 50),
	}

	// Create packer and pack items
	packer := boxpacker3.NewPacker()
	result, err := packer.PackCtx(context.Background(), boxes, items)
	require.NoError(t, err)

	// Verify packing result
	require.NotNil(t, result)
	require.Len(t, result.Boxes, len(boxes))

	// Count total packed items
	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	// At least some items should be packed
	require.Positive(t, totalPacked, "At least some items should be packed")
}

// Test2DItemDimensions verifies that 2D items have correct dimensions.
func Test2DItemDimensions(t *testing.T) {
	t.Parallel()

	item := boxpacker3.NewItem2D("item-1", 100, 50, 200)

	require.InDelta(t, 100.0, item.GetWidth(), 0.0001)
	require.InDelta(t, 50.0, item.GetHeight(), 0.0001)
	require.InDelta(t, 1.0, item.GetDepth(), 0.0001)
	require.InDelta(t, 5000.0, item.GetVolume(), 0.0001) // 100 * 50 * 1
	require.InDelta(t, 200.0, item.GetWeight(), 0.0001)
}

// Test2DBoxDimensions verifies that 2D boxes have correct dimensions.
func Test2DBoxDimensions(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox2D("box-1", 200, 100, 500)

	require.InDelta(t, 200.0, box.GetWidth(), 0.0001)
	require.InDelta(t, 100.0, box.GetHeight(), 0.0001)
	require.InDelta(t, 1.0, box.GetDepth(), 0.0001)
	require.InDelta(t, 20000.0, box.GetVolume(), 0.0001) // 200 * 100 * 1
	require.InDelta(t, 500.0, box.GetMaxWeight(), 0.0001)
}
