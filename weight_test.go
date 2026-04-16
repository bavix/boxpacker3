package boxpacker3_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func lopsidedFixture() ([]*boxpacker3.Box, []*boxpacker3.Item) {
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("first", 100, 100, 100, 100000),
		boxpacker3.NewBox("second", 100, 100, 100, 100000),
	}

	items := make([]*boxpacker3.Item, 0, 12)
	for i := range 12 {
		items = append(items, boxpacker3.NewItem(fmt.Sprintf("brick-%d", i), 50, 50, 50, 1000))
	}

	return boxes, items
}

func weightSpread(result *boxpacker3.Result) float64 {
	weights := make([]float64, 0, len(result.Boxes))

	for _, box := range result.Boxes {
		if len(box.Items) > 0 {
			weights = append(weights, box.Stats.ItemsWeight)
		}
	}

	if len(weights) < 2 {
		return 0
	}

	spread := 0.0

	for i := range weights {
		for j := i + 1; j < len(weights); j++ {
			spread = math.Max(spread, math.Abs(weights[i]-weights[j]))
		}
	}

	return spread
}

func TestWeight_SpreadIsReduced(t *testing.T) {
	t.Parallel()

	boxes, items := lopsidedFixture()

	unbalanced, err := boxpacker3.NewPacker(boxpacker3.WithFinishers()).
		Pack(t.Context(), boxes, items)
	require.NoError(t, err)

	balanced, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)

	requireInvariants(t, items, balanced)
	require.Less(t, weightSpread(balanced), weightSpread(unbalanced))
	require.Empty(t, balanced.Unpacked)
}

func TestWeight_PassCanBeDisabled(t *testing.T) {
	t.Parallel()

	boxes, items := lopsidedFixture()

	for _, bound := range []int{0, 1} {
		result, err := boxpacker3.NewPacker(boxpacker3.WithFinishers(boxpacker3.BalanceWeight{MaxBoxes: bound})).
			Pack(t.Context(), boxes, items)
		require.NoError(t, err)

		reference, err := boxpacker3.NewPacker(boxpacker3.WithFinishers()).
			Pack(t.Context(), boxes, items)
		require.NoError(t, err)

		require.InDelta(t, weightSpread(reference), weightSpread(result), 1e-9)
	}
}

func TestWeight_BoundIsRespected(t *testing.T) {
	t.Parallel()

	boxes, items := lopsidedFixture()

	unbalanced, err := boxpacker3.NewPacker(boxpacker3.WithFinishers()).
		Pack(t.Context(), boxes, items)
	require.NoError(t, err)

	capped, err := boxpacker3.NewPacker(boxpacker3.WithFinishers(boxpacker3.BalanceWeight{MaxBoxes: 1})).
		Pack(t.Context(), boxes, items)
	require.NoError(t, err)

	require.InDelta(t, weightSpread(unbalanced), weightSpread(capped), 1e-9,
		"a bound below the number of boxes used keeps the pass out")
}

func TestWeight_SingleBoxIsLeftAlone(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("only", 100, 100, 100, 100000)}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("a", 40, 40, 40, 900),
		boxpacker3.NewItem("b", 40, 40, 40, 100),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)
	require.Len(t, result.Boxes[0].Items, 2)
}

func TestWeight_NothingIsLost(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("a", 90, 90, 90, 100000),
		boxpacker3.NewBox("b", 90, 90, 90, 100000),
		boxpacker3.NewBox("c", 90, 90, 90, 100000),
	}

	items := make([]*boxpacker3.Item, 0, 20)

	for i := range 20 {
		size := float64(20 + i%5*8)
		items = append(items, boxpacker3.NewItem(fmt.Sprintf("i%d", i), size, size, size, float64(100+i*250)))
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)
}

func TestWeight_GroupsAreNeverMoved(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("first", 100, 100, 100, 100000),
		boxpacker3.NewBox("second", 100, 100, 100, 100000),
	}

	items := []*boxpacker3.Item{
		mustItem(t, boxpacker3.ItemSpec{ID: "kit-a", Width: 50, Height: 50, Depth: 50, Weight: 4000, Group: groupKit}),
		mustItem(t, boxpacker3.ItemSpec{ID: "kit-b", Width: 50, Height: 50, Depth: 50, Weight: 4000, Group: groupKit}),
		boxpacker3.NewItem("loose-1", 50, 50, 50, 100),
		boxpacker3.NewItem("loose-2", 50, 50, 50, 100),
		boxpacker3.NewItem("loose-3", 50, 50, 50, 100),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)

	first := boxOf(result, "kit-a")
	second := boxOf(result, "kit-b")

	require.NotNil(t, first)
	require.Same(t, first, second, "the pass never splits a group")
}
