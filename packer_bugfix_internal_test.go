package boxpacker3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPackToBox_ReportsCancellation(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 10000)

	items := make([]*piece, 0, 50)
	for i := range 50 {
		items = append(items, testPiece(string(rune('a'+i%26))+"-item", 10, 10, 10, 1))
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	unpacked, err := packToBox(ctx, box, items)
	require.ErrorIs(t, err, context.Canceled)
	require.Len(t, unpacked, len(items), "nothing was packed, so everything comes back")
}

func TestRunFirstFit_PropagatesCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	boxes := []*Box{NewBox("box", 100, 100, 100, 10000)}
	items := []*piece{testPiece("item", 10, 10, 10, 1)}

	result, err := runFirstFit(ctx, boxes, items, problemOf(boxes, items))

	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, result)
}

func TestIsPerfectFit_IsRelativeToBoxVolume(t *testing.T) {
	t.Parallel()

	metreScale := openBox("metres", 1, 1, 1, 1000)
	require.True(t, metreScale.putItem(testPiece("almost", 0.998, 0.998, 0.998, 1), Pivot{}))
	require.Greater(t, metreScale.remainingVolume(), 0.001)
	require.False(t, isPerfectFit(metreScale), "6 litres of air is not a perfect fit")

	exact := openBox("exact", 10, 10, 10, 1000)
	require.True(t, exact.putItem(testPiece("full", 10, 10, 10, 1), Pivot{}))
	require.True(t, isPerfectFit(exact))

	require.False(t, isPerfectFit(nil))
	require.False(t, isPerfectFit(openBox("degenerate", 0, 0, 0, 1)))
}

func TestWithinLimit_AbsorbsAccumulatedRoundingError(t *testing.T) {
	t.Parallel()

	require.True(t, withinLimit(0.8+0.8+0.8, 2.4))
	require.True(t, withinLimit(2.4, 2.4))
	require.False(t, withinLimit(2.5, 2.4))
	require.True(t, withinLimit(0, 0))
}

func TestAttemptRepack_RestoresStateOnFailure(t *testing.T) {
	t.Parallel()

	t.Run("new item exceeds quota", func(t *testing.T) {
		t.Parallel()

		box := openBox("box", 10, 10, 10, 1000)
		lower := testPiece("lower", 10, 10, 4, 1)
		upper := testPiece("upper", 10, 10, 4, 1)

		require.True(t, box.putItem(lower, Pivot{}))
		require.True(t, box.putItem(upper, Pivot{0, 0, 4}))

		newcomer := testPiece("newcomer", 10, 10, 4, 1)

		require.False(t, attemptRepack(box, newcomer))
		require.Equal(t, Pivot{}, lower.position)
		require.Equal(t, Pivot{0, 0, 4}, upper.position)
		require.Equal(t, Pivot{}, newcomer.position)
		require.Len(t, box.items, 2)
	})

	t.Run("existing item no longer fits", func(t *testing.T) {
		t.Parallel()

		box := openBox("box", 10, 10, 10, 1000)
		slab := testPiece("slab", 10, 10, 5, 1)
		require.True(t, box.putItem(slab, Pivot{}))

		cube := testPiece("cube", 6, 6, 6, 1)

		require.False(t, attemptRepack(box, cube))
		require.Equal(t, Pivot{}, slab.position)
		require.Equal(t, OrientationWHD, slab.orientation)
		require.Equal(t, Pivot{}, cube.position)
		require.Len(t, box.items, 1)
		require.Same(t, slab, box.items[0], "the caller's item stays in the box")
	})
}

func TestCompact(t *testing.T) {
	t.Parallel()

	items := compact([]*piece{nil, testPiece("a", 1, 1, 1, 1), nil, testPiece("b", 1, 1, 1, 1)})
	require.Len(t, items, 2)
	require.Equal(t, "a", items[0].ID())
	require.Equal(t, "b", items[1].ID())

	boxes := compact([]*Box{nil, NewBox("a", 1, 1, 1, 1), nil})
	require.Len(t, boxes, 1)
	require.Equal(t, "a", boxes[0].ID())

	require.Empty(t, compact([]*piece(nil)))
	require.Empty(t, compact([]*Container(nil)))
}
