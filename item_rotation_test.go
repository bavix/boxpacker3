package boxpacker3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestItem_WidthHeightDepth(t *testing.T) {
	t.Parallel()

	item := boxpacker3.NewItem("test", 10, 20, 30, 1)

	require.InDelta(t, 10.0, item.Width(), 0.0001, "Width should be 10")
	require.InDelta(t, 20.0, item.Height(), 0.0001, "Height should be 20")
	require.InDelta(t, 30.0, item.Depth(), 0.0001, "Depth should be 30")
}

func TestItem_WhdArray_Access(t *testing.T) {
	t.Parallel()

	item := boxpacker3.NewItem("test", 10, 20, 30, 1)

	require.InDelta(t, 10.0, item.Width(), 0.0001, "Width should be accessible via GetWidth")
	require.InDelta(t, 20.0, item.Height(), 0.0001, "Height should be accessible via GetHeight")
	require.InDelta(t, 30.0, item.Depth(), 0.0001, "Depth should be accessible via GetDepth")
}

func TestItem_PlacedDirectlyKeepsItsOrientation(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("box", 100, 100, 100, 1000)
	item := boxpacker3.NewItem("item", 10, 20, 30, 1)

	result, err := boxpacker3.NewPacker().Pack(t.Context(), []*boxpacker3.Box{box}, []*boxpacker3.Item{item})
	require.NoError(t, err)
	require.Len(t, result.Boxes, 1)

	items := result.Boxes[0].Items
	require.Len(t, items, 1, "Should have one item")
	require.Equal(t, item.ID(), items[0].ID(), "Should be the same item")

	dim := items[0].Dimension
	require.Greater(t, dim[0], 0.0, "Width should be positive")
	require.Greater(t, dim[1], 0.0, "Height should be positive")
	require.Greater(t, dim[2], 0.0, "Depth should be positive")
}

func TestItem_WhdArray_EdgeCases(t *testing.T) {
	t.Parallel()

	item := boxpacker3.NewItem("tiny", 0.001, 0.002, 0.003, 0.1)
	require.InDelta(t, 0.001, item.Width(), 0.0001, "Very small width should work")
	require.InDelta(t, 0.002, item.Height(), 0.0001, "Very small height should work")
	require.InDelta(t, 0.003, item.Depth(), 0.0001, "Very small depth should work")

	item = boxpacker3.NewItem("large", 10000, 20000, 30000, 1000)
	require.InDelta(t, 10000.0, item.Width(), 0.0001, "Very large width should work")
	require.InDelta(t, 20000.0, item.Height(), 0.0001, "Very large height should work")
	require.InDelta(t, 30000.0, item.Depth(), 0.0001, "Very large depth should work")

	item = boxpacker3.NewItem("cube", 5, 5, 5, 1)
	require.InDelta(t, 5.0, item.Width(), 0.0001, "Equal dimensions should work")
	require.InDelta(t, 5.0, item.Height(), 0.0001, "Equal dimensions should work")
	require.InDelta(t, 5.0, item.Depth(), 0.0001, "Equal dimensions should work")
}
