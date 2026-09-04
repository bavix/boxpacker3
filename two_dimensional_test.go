package boxpacker3_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestTwoD_FlatItemStaysFlatInADeepBox(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("deep", 100, 100, 100, 10000)}
	items := []*boxpacker3.Item{boxpacker3.NewItem2D("sheet", 80, 60, 10)}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)
	require.Empty(t, result.Unpacked)

	packed := result.Boxes[0].Items[0]
	require.InDelta(t, 1.0, packed.Dimension[2], 1e-9, "the sheet keeps its declared depth")
}

func TestTwoD_QuarterTurnIsAllowed(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox2D("narrow", 40, 90, 10000)}
	items := []*boxpacker3.Item{boxpacker3.NewItem2D("sheet", 80, 30, 10)}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Empty(t, result.Unpacked, "the sheet fits after a quarter turn")

	sheet := result.Boxes[0].Items[0]
	width, height := sheet.Dimension[0], sheet.Dimension[1]
	require.InDelta(t, 30.0, width, 1e-9)
	require.InDelta(t, 80.0, height, 1e-9)
}

func TestTwoD_ThreeDimensionalItemsKeepEveryOrientation(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("slot", 100, 12, 90, 1000)}
	items := []*boxpacker3.Item{boxpacker3.NewItem("blade", 90, 80, 10, 100)}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Empty(t, result.Unpacked, "a general item may stand on edge to fit a slot")
}

func TestTwoD_FlatnessIsVisible(t *testing.T) {
	t.Parallel()

	require.True(t, boxpacker3.NewBox2D("flat", 10, 10, 100).Flat())
	require.False(t, boxpacker3.NewBox("deep", 10, 10, 10, 100).Flat())
	require.True(t, boxpacker3.NewItem2D("sheet", 5, 5, 1).Flat())
	require.False(t, boxpacker3.NewItem("brick", 5, 5, 5, 1).Flat())
	require.False(t, boxpacker3.NewItem("thin", 5, 5, 1, 1).Flat(),
		"a one-deep general item is not marked flat")
}

func TestTwoD_Areas(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox2D("sheet", 100, 50, 10000)
	require.InDelta(t, 5000.0, box.Area(), 1e-9)

	items := []*boxpacker3.Item{
		boxpacker3.NewItem2D("a", 30, 20, 1),
		boxpacker3.NewItem2D("b", 40, 25, 1),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)
	require.Empty(t, result.Unpacked)

	packedBox := result.Boxes[0]
	require.InDelta(t, 5000.0, packedBox.Box.Area(), 1e-9)
	require.InDelta(t, 5000.0-30*20-40*25, packedBox.RemainingArea(), 1e-9)

	for _, item := range packedBox.Items {
		width, height := item.Dimension[0], item.Dimension[1]
		require.InDelta(t, width*height, item.Area(), 1e-9)

		x, y := item.Position[0], item.Position[1]
		require.GreaterOrEqual(t, x, 0.0)
		require.GreaterOrEqual(t, y, 0.0)
	}
}

func TestTwoD_MixedWithThreeDimensionalItems(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("deep", 120, 120, 60, 100000)}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("brick-1", 40, 40, 40, 100),
		boxpacker3.NewItem2D("sheet-1", 100, 100, 10),
		boxpacker3.NewItem("brick-2", 30, 30, 30, 100),
		boxpacker3.NewItem2D("sheet-2", 60, 60, 10),
	}

	for _, strategy := range allStrategies() {
		t.Run(strategyName(strategy), func(t *testing.T) {
			t.Parallel()

			result, err := rulePacker(strategy.Order, strategy.Selection).
				Pack(t.Context(), boxes, items)
			require.NoError(t, err)
			requireInvariants(t, items, result)

			for _, box := range result.Boxes {
				for _, item := range box.Items {
					if !item.Item.Flat() {
						continue
					}

					require.InDelta(t, 1.0, item.Dimension[2], 1e-9,
						"flat item %s was stood on its edge", item.ID())
				}
			}
		})
	}
}

func TestTwoD_PacksASheetOfPanels(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox2D("sheet", 120, 80, 100000)}

	items := make([]*boxpacker3.Item, 0, 24)
	for i := range 24 {
		items = append(items, boxpacker3.NewItem2D(fmt.Sprintf("panel-%d", i), 20, 20, 1))
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)
	require.Empty(t, result.Unpacked, "twenty four 20x20 panels tile a 120x80 sheet")
	require.InDelta(t, 0.0, result.Boxes[0].RemainingArea(), 1e-9)
}
