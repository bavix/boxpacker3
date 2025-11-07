package boxpacker3_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

// TestPacker_Pack_NilInputs tests that Pack handles nil inputs gracefully.
func TestPacker_Pack_NilInputs(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker()

	// Test with nil boxes
	result := packer.Pack(nil, []*boxpacker3.Item{
		boxpacker3.NewItem("item1", 10, 10, 10, 1),
	})
	require.NotNil(t, result)
	require.NotNil(t, result.Boxes)
	require.NotNil(t, result.UnfitItems)
	require.Len(t, result.UnfitItems, 1, "Item should be in UnfitItems when no boxes provided")

	// Test with nil items
	box := boxpacker3.NewBox("box1", 100, 100, 100, 100)
	result = packer.Pack([]*boxpacker3.Box{box}, nil)
	require.NotNil(t, result)
	require.NotNil(t, result.Boxes)
	require.NotNil(t, result.UnfitItems)
	require.Empty(t, result.UnfitItems, "No items should be in UnfitItems")

	// Test with both nil
	result = packer.Pack(nil, nil)
	require.NotNil(t, result)
	require.NotNil(t, result.Boxes)
	require.NotNil(t, result.UnfitItems)
	require.Empty(t, result.Boxes)
	require.Empty(t, result.UnfitItems)
}

// TestPacker_Pack_NilElementsInSlices tests that Pack handles nil elements in slices.
func TestPacker_Pack_NilElementsInSlices(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker()

	box := boxpacker3.NewBox("box1", 100, 100, 100, 100)
	item := boxpacker3.NewItem("item1", 10, 10, 10, 1)

	// Test with nil box in slice
	result := packer.Pack([]*boxpacker3.Box{box, nil, box}, []*boxpacker3.Item{item})
	require.NotNil(t, result)
	require.GreaterOrEqual(t, len(result.Boxes), 1, "Should have at least one box")

	// Test with nil item in slice
	result = packer.Pack([]*boxpacker3.Box{box}, []*boxpacker3.Item{item, nil, item})
	require.NotNil(t, result)
	// Count all items
	// Note: With StrategyMinimizeBoxes (default), nil items are added to unpacked in packToBox
	// With other strategies (packWith*), nil items are skipped (continue)
	// So the behavior depends on the strategy, but we just verify it doesn't panic
	totalItems := len(result.UnfitItems)
	for _, b := range result.Boxes {
		totalItems += len(b.GetItems())
	}
	// Should have at least 2 items (the non-nil ones), possibly 3 if nil is added to unpacked
	require.GreaterOrEqual(t, totalItems, 2, "Should have at least 2 items (non-nil items)")
	require.LessOrEqual(t, totalItems, 3, "Should have at most 3 items (2 non-nil + possibly 1 nil in unpacked)")
}

// TestPacker_PackCtx_ContextCancellation tests that PackCtx handles context cancellation correctly.
func TestPacker_PackCtx_ContextCancellation(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker()

	// Create a large number of items to make packing take some time
	items := make([]*boxpacker3.Item, 1000)
	for i := range items {
		items[i] = boxpacker3.NewItem(
			"item"+string(rune(i)),
			float64(10+i%10),
			float64(10+i%10),
			float64(10+i%10),
			1,
		)
	}

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box1", 100, 100, 100, 1000),
	}

	// Create a context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// PackCtx should return error immediately
	result, err := packer.PackCtx(ctx, boxes, items)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, context.Canceled, err)
}

// TestPacker_PackCtx_Timeout tests that PackCtx handles timeout correctly.
func TestPacker_PackCtx_Timeout(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker()

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item1", 10, 10, 10, 1),
	}

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box1", 100, 100, 100, 100),
	}

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait a bit to ensure timeout
	time.Sleep(10 * time.Millisecond)

	// PackCtx should return timeout error
	result, err := packer.PackCtx(ctx, boxes, items)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, context.DeadlineExceeded, err)
}

// TestPacker_PackCtx_Success tests that PackCtx works correctly when context is not cancelled.
func TestPacker_PackCtx_Success(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker()

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item1", 10, 10, 10, 1),
	}

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box1", 100, 100, 100, 100),
	}

	ctx := context.Background()

	// PackCtx should return result successfully
	result, err := packer.PackCtx(ctx, boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.GreaterOrEqual(t, len(result.Boxes), 1, "Should have at least one box")
}

// TestNewItem_NoValidation tests that NewItem accepts values as-is without validation.
func TestNewItem_NoValidation(t *testing.T) {
	t.Parallel()
	// Test with zero dimensions - should be accepted as-is
	item := boxpacker3.NewItem("item1", 0, 0, 0, 0)
	require.NotNil(t, item)
	require.InDelta(t, 0.0, item.GetWidth(), 0.0001, "Width should be exactly 0")
	require.InDelta(t, 0.0, item.GetHeight(), 0.0001, "Height should be exactly 0")
	require.InDelta(t, 0.0, item.GetDepth(), 0.0001, "Depth should be exactly 0")
	require.InDelta(t, 0.0, item.GetWeight(), 0.0001, "Weight should be exactly 0")

	// Test with negative dimensions - should be accepted as-is
	item = boxpacker3.NewItem("item2", -10, -20, -30, -5)
	require.NotNil(t, item)
	require.InDelta(t, -10.0, item.GetWidth(), 0.0001, "Width should be exactly -10")
	require.InDelta(t, -20.0, item.GetHeight(), 0.0001, "Height should be exactly -20")
	require.InDelta(t, -30.0, item.GetDepth(), 0.0001, "Depth should be exactly -30")
	require.InDelta(t, -5.0, item.GetWeight(), 0.0001, "Weight should be exactly -5")

	// Test with empty ID - should be accepted as-is
	item = boxpacker3.NewItem("", 10, 10, 10, 1)
	require.NotNil(t, item)
	require.Empty(t, item.GetID(), "ID should be empty string")
}

// TestNewBox_NoValidation tests that NewBox accepts values as-is without validation.
func TestNewBox_NoValidation(t *testing.T) {
	t.Parallel()
	// Test with zero dimensions - should be accepted as-is
	box := boxpacker3.NewBox("box1", 0, 0, 0, 0)
	require.NotNil(t, box)
	require.InDelta(t, 0.0, box.GetWidth(), 0.0001, "Width should be exactly 0")
	require.InDelta(t, 0.0, box.GetHeight(), 0.0001, "Height should be exactly 0")
	require.InDelta(t, 0.0, box.GetDepth(), 0.0001, "Depth should be exactly 0")
	require.InDelta(t, 0.0, box.GetMaxWeight(), 0.0001, "MaxWeight should be exactly 0")

	// Test with negative dimensions - should be accepted as-is
	box = boxpacker3.NewBox("box2", -10, -20, -30, -5)
	require.NotNil(t, box)
	require.InDelta(t, -10.0, box.GetWidth(), 0.0001, "Width should be exactly -10")
	require.InDelta(t, -20.0, box.GetHeight(), 0.0001, "Height should be exactly -20")
	require.InDelta(t, -30.0, box.GetDepth(), 0.0001, "Depth should be exactly -30")
	require.InDelta(t, -5.0, box.GetMaxWeight(), 0.0001, "MaxWeight should be exactly -5")

	// Test with empty ID - should be accepted as-is
	box = boxpacker3.NewBox("", 100, 100, 100, 100)
	require.NotNil(t, box)
	require.Empty(t, box.GetID(), "ID should be empty string")
}

// TestItem_Intersect_Nil tests that Intersect handles nil items correctly.
func TestItem_Intersect_Nil(t *testing.T) {
	t.Parallel()

	item1 := boxpacker3.NewItem("item1", 10, 10, 10, 1)

	// Test with nil item
	result := item1.Intersect(nil)
	require.False(t, result, "Intersect with nil should return false")

	// Test with nil receiver (this would panic in Go, but we test the method handles it)
	// Note: In Go, calling a method on nil pointer causes panic, so we can't test this directly
	// But we can test that the method doesn't panic when called with nil parameter
	item2 := boxpacker3.NewItem("item2", 10, 10, 10, 1)
	result = item2.Intersect(nil)
	require.False(t, result, "Intersect with nil should return false")
}

// TestBox_PutItem_Nil tests that PutItem handles nil items correctly.
func TestBox_PutItem_Nil(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("box1", 100, 100, 100, 100)

	// Test with nil item
	result := box.PutItem(nil, boxpacker3.Pivot{})
	require.False(t, result, "PutItem with nil should return false")
}

// TestBox_CanQuota_Nil tests that canQuota methods handle nil items correctly.
func TestBox_CanQuota_Nil(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("box1", 100, 100, 100, 100)

	// Test canQuota with nil (indirectly through PutItem which calls canQuota)
	result := box.PutItem(nil, boxpacker3.Pivot{})
	require.False(t, result, "PutItem with nil should return false")
}

// TestPacker_AllStrategies_NilHandling tests that all strategies handle nil elements correctly.
func TestPacker_AllStrategies_NilHandling(t *testing.T) {
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

	box := boxpacker3.NewBox("box1", 100, 100, 100, 100)
	item := boxpacker3.NewItem("item1", 10, 10, 10, 1)

	strategyNames := []string{
		"StrategyMinimizeBoxes",
		"StrategyGreedy",
		"StrategyBestFit",
		"StrategyBestFitDecreasing",
		"StrategyNextFit",
		"StrategyWorstFit",
		"StrategyAlmostWorstFit",
	}

	for i, strategy := range strategies {
		t.Run(strategyNames[i], func(t *testing.T) {
			t.Parallel()

			packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(strategy))

			// Test with nil item in slice
			result := packer.Pack([]*boxpacker3.Box{box}, []*boxpacker3.Item{item, nil, item})
			require.NotNil(t, result)

			// Test with nil box in slice
			result = packer.Pack([]*boxpacker3.Box{box, nil, box}, []*boxpacker3.Item{item})
			require.NotNil(t, result)
		})
	}
}
