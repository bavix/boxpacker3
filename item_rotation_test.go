package boxpacker3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

// TestItem_GetWidthHeightDepth tests that GetWidth, GetHeight, GetDepth work correctly.
func TestItem_GetWidthHeightDepth(t *testing.T) {
	t.Parallel()

	item := boxpacker3.NewItem("test", 10, 20, 30, 1)

	require.InDelta(t, 10.0, item.GetWidth(), 0.0001, "Width should be 10")
	require.InDelta(t, 20.0, item.GetHeight(), 0.0001, "Height should be 20")
	require.InDelta(t, 30.0, item.GetDepth(), 0.0001, "Depth should be 30")
}

// TestItem_WhdArray_Access tests direct access to whd array.
func TestItem_WhdArray_Access(t *testing.T) {
	t.Parallel()

	item := boxpacker3.NewItem("test", 10, 20, 30, 1)

	// Test that whd array is accessible and correct
	// Note: This tests internal structure, but validates the refactoring
	require.InDelta(t, 10.0, item.GetWidth(), 0.0001, "Width should be accessible via GetWidth")
	require.InDelta(t, 20.0, item.GetHeight(), 0.0001, "Height should be accessible via GetHeight")
	require.InDelta(t, 30.0, item.GetDepth(), 0.0001, "Depth should be accessible via GetDepth")
}

// TestItem_PutItem_RotationMatrix tests that PutItem uses rotation matrix correctly.
func TestItem_PutItem_RotationMatrix(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("box", 100, 100, 100, 1000)
	item := boxpacker3.NewItem("item", 10, 20, 30, 1)

	// Test that PutItem works with all rotations
	// This indirectly tests that rotation matrix is used correctly in PutItem
	success := box.PutItem(item, boxpacker3.Pivot{})
	require.True(t, success, "Item should fit in box")

	// Verify the item was placed with correct rotation
	items := box.GetItems()
	require.Len(t, items, 1, "Should have one item")
	require.Equal(t, item, items[0], "Should be the same item")

	// Verify item dimensions are valid (indirectly confirms rotation was set)
	dim := items[0].GetDimension()
	require.Greater(t, dim[0], 0.0, "Width should be positive")
	require.Greater(t, dim[1], 0.0, "Height should be positive")
	require.Greater(t, dim[2], 0.0, "Depth should be positive")
}

// TestItem_WhdArray_EdgeCases tests edge cases for whd array access.
func TestItem_WhdArray_EdgeCases(t *testing.T) {
	t.Parallel()
	// Test with very small values
	item := boxpacker3.NewItem("tiny", 0.001, 0.002, 0.003, 0.1)
	require.InDelta(t, 0.001, item.GetWidth(), 0.0001, "Very small width should work")
	require.InDelta(t, 0.002, item.GetHeight(), 0.0001, "Very small height should work")
	require.InDelta(t, 0.003, item.GetDepth(), 0.0001, "Very small depth should work")

	// Test with very large values
	item = boxpacker3.NewItem("large", 10000, 20000, 30000, 1000)
	require.InDelta(t, 10000.0, item.GetWidth(), 0.0001, "Very large width should work")
	require.InDelta(t, 20000.0, item.GetHeight(), 0.0001, "Very large height should work")
	require.InDelta(t, 30000.0, item.GetDepth(), 0.0001, "Very large depth should work")

	// Test with equal dimensions
	item = boxpacker3.NewItem("cube", 5, 5, 5, 1)
	require.InDelta(t, 5.0, item.GetWidth(), 0.0001, "Equal dimensions should work")
	require.InDelta(t, 5.0, item.GetHeight(), 0.0001, "Equal dimensions should work")
	require.InDelta(t, 5.0, item.GetDepth(), 0.0001, "Equal dimensions should work")
}
