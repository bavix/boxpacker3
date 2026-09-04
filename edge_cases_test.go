package boxpacker3_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func packAllStrategies(t *testing.T, boxes []*boxpacker3.Box, items []*boxpacker3.Item) {
	t.Helper()

	for _, strategy := range allStrategies() {
		t.Run(strategyName(strategy), func(t *testing.T) {
			t.Parallel()

			result, err := rulePacker(strategy.Order, strategy.Selection).
				Pack(t.Context(), boxes, items)
			require.NoError(t, err)

			requireInvariants(t, items, result)
		})
	}
}

func TestEdge_ItemExactlyFillsTheBox(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("exact", 100, 80, 60, 1000)}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("filler", 100, 80, 60, 100),
		boxpacker3.NewItem("crumb", 1, 1, 1, 1),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)

	require.Len(t, result.Boxes[0].Items, 1)
	require.Len(t, result.Unpacked, 1)
	require.Equal(t, "crumb", result.Unpacked[0].ID())
}

func TestEdge_EqualDimensions(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("cube", 60, 60, 60, 10000)}

	items := make([]*boxpacker3.Item, 0, 27)
	for i := range 27 {
		items = append(items, boxpacker3.NewItem(fmt.Sprintf("cube-%d", i), 20, 20, 20, 1))
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)

	require.Empty(t, result.Unpacked, "twenty seven cubes fill a cube three on a side")
}

func TestEdge_ZeroSizedItems(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 50, 50, 50, 1000)}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("flat", 10, 10, 0, 1),
		boxpacker3.NewItem("line", 10, 0, 0, 1),
		boxpacker3.NewItem("point", 0, 0, 0, 1),
		boxpacker3.NewItem("solid", 20, 20, 20, 1),
	}

	packAllStrategies(t, boxes, items)
}

func TestEdge_ManyCopiesOfOneItem(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 100000)}

	items := make([]*boxpacker3.Item, 0, 120)
	for i := range 120 {
		items = append(items, boxpacker3.NewItem(fmt.Sprintf("brick-%d", i), 25, 20, 15, 10))
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)
}

func TestEdge_ExtremeAspectRatio(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("corridor", 2000, 20, 20, 100000)}

	items := make([]*boxpacker3.Item, 0, 30)
	for i := range 30 {
		items = append(items, boxpacker3.NewItem(fmt.Sprintf("rod-%d", i), 60, 18, 18, 10))
	}

	packAllStrategies(t, boxes, items)
}

func TestEdge_FitsInOneOrientationOnly(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("slot", 100, 12, 90, 1000)}
	items := []*boxpacker3.Item{boxpacker3.NewItem("blade", 90, 10, 80, 100)}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)
	require.Empty(t, result.Unpacked)
}

func TestEdge_DimensionsThatSumOnlyApproximately(t *testing.T) {
	t.Parallel()

	sizes := []float64{0.1, 0.3, 0.7, 1.1, 2.3}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("size-%g", size), func(t *testing.T) {
			t.Parallel()

			for axis := range 3 {
				dimension := [3]float64{1, 1, 1}
				dimension[axis] = size

				boxDimension := [3]float64{1, 1, 1}
				boxDimension[axis] = size * 7

				boxes := []*boxpacker3.Box{boxpacker3.NewBox("row",
					boxDimension[0], boxDimension[1], boxDimension[2], 10000)}

				items := make([]*boxpacker3.Item, 0, 7)
				for i := range 7 {
					items = append(items, boxpacker3.NewItem(fmt.Sprintf("slice-%d-%d", axis, i),
						dimension[0], dimension[1], dimension[2], 1))
				}

				result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
				require.NoError(t, err)
				requireInvariants(t, items, result)
				require.Empty(t, result.Unpacked,
					"seven slices of %g fill a row of %g on axis %d", size, size*7, axis)
			}
		})
	}
}

func TestEdge_SingleItemLargerThanEveryBox(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small", 10, 10, 10, 100),
		boxpacker3.NewBox("medium", 20, 20, 20, 100),
	}
	items := []*boxpacker3.Item{boxpacker3.NewItem("oversized", 50, 50, 50, 1)}

	packAllStrategies(t, boxes, items)
}

func TestEdge_ItemHeavierThanEveryBox(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 10)}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("anvil", 10, 10, 10, 1000),
		boxpacker3.NewItem("feather", 10, 10, 10, 1),
	}

	packAllStrategies(t, boxes, items)
}
