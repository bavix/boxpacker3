package boxpacker3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func filledVolumeOf(result *boxpacker3.Result) float64 {
	volume := 0.0

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			volume += item.Volume()
		}
	}

	return volume
}

func TestSingleBox_EverythingFits(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("container", 100, 100, 100, 100000)
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("a", 40, 40, 40, 10),
		boxpacker3.NewItem("b", 40, 40, 40, 10),
	}

	result, err := singlePacker().Pack(t.Context(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)
	require.Empty(t, result.Unpacked)
}

func TestSingleBox_MoreCargoThanCapacity(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("container", 100, 100, 100, 100000)

	items := make([]*boxpacker3.Item, 0, 20)
	for i := range 20 {
		items = append(items, boxpacker3.NewItem(fmt.Sprintf("crate-%d", i), 50, 50, 50, 10))
	}

	result, err := singlePacker().Pack(t.Context(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)

	require.Len(t, result.Boxes[0].Items, 8, "eight 50-cubes fill a 100-cube")
	require.Len(t, result.Unpacked, 12)
}

func TestSingleBox_LeavesABulkyItemBehind(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("container", 100, 100, 100, 100000)

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("bulky", 100, 100, 60, 10),
		boxpacker3.NewItem("neat-1", 100, 100, 50, 10),
		boxpacker3.NewItem("neat-2", 100, 100, 50, 10),
	}

	ordinary, err := boxpacker3.NewPacker().
		Pack(t.Context(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)

	dense, err := singlePacker().Pack(t.Context(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)
	requireInvariants(t, items, dense)

	require.Greater(t, filledVolumeOf(dense), filledVolumeOf(ordinary),
		"dropping the bulky item lets the two neat ones fill the container")
	require.Len(t, dense.Unpacked, 1)
	require.Equal(t, "bulky", dense.Unpacked[0].ID())
}

func TestSingleBox_NeverWorseThanOrdinaryPacking(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("container", 120, 100, 80, 100000)

	items := make([]*boxpacker3.Item, 0, 25)

	for i := range 25 {
		size := float64(20 + i%6*9)
		items = append(items, boxpacker3.NewItem(fmt.Sprintf("i%d", i), size, size+5, size-3, 10))
	}

	ordinary, err := boxpacker3.NewPacker().
		Pack(t.Context(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)

	dense, err := singlePacker().Pack(t.Context(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)
	requireInvariants(t, items, dense)

	require.GreaterOrEqual(t, filledVolumeOf(dense), filledVolumeOf(ordinary))
}

func TestSingleBox_NilBox(t *testing.T) {
	t.Parallel()

	items := []*boxpacker3.Item{boxpacker3.NewItem("a", 10, 10, 10, 1)}

	result, err := singlePacker().Pack(t.Context(), []*boxpacker3.Box{nil}, items)
	require.NoError(t, err)
	require.Len(t, result.Unpacked, 1)
}

func TestSingleBox_CancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	box := boxpacker3.NewBox("container", 100, 100, 100, 100000)
	items := []*boxpacker3.Item{boxpacker3.NewItem("a", 10, 10, 10, 1)}

	_, err := singlePacker().Pack(ctx, []*boxpacker3.Box{box}, items)
	require.ErrorIs(t, err, context.Canceled)
}

func TestSingleBox_StockOfAKindIsStillOneBox(t *testing.T) {
	t.Parallel()

	box, err := boxpacker3.NewBoxFromSpec(boxpacker3.BoxSpec{
		ID: "container", OuterWidth: 10, OuterHeight: 10, OuterDepth: 10,
		MaxWeight: 1e9, Quantity: 3,
	})
	require.NoError(t, err)

	items := make([]*boxpacker3.Item, 0, 20)
	for i := range 20 {
		items = append(items, boxpacker3.NewItem(fmt.Sprintf("cube-%d", i), 5, 5, 5, 1))
	}

	result, err := singlePacker().Pack(context.Background(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)

	require.Len(t, result.Boxes, 1)
	require.Len(t, result.Boxes[0].Items, 8)
	require.Len(t, result.Unpacked, 12)
}

func TestSingleBox_LeftoverIsTheRightPieceWhenIDsRepeat(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("container", 10, 10, 10, 1e9)
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("sku", 3, 3, 3, 1),
		boxpacker3.NewItem("sku", 10, 10, 10, 1),
	}

	result, err := singlePacker().Pack(context.Background(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)
	require.Len(t, result.Boxes, 1)
	require.Len(t, result.Boxes[0].Items, 1)
	require.Len(t, result.Unpacked, 1)

	packed, left := result.Boxes[0].Items[0], result.Unpacked[0]
	require.NotEqual(t, packed.Volume(), left.Volume())
	require.InDelta(t, 1027, packed.Volume()+left.Volume(), 1e-9)
}
