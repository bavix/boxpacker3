package boxpacker3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBox_SupportRatioOnTheFloor(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	require.InDelta(t, 1.0, box.supportRatio(Pivot{}, Dimension{10, 10, 10}), 1e-9)
}

func TestBox_SupportRatioIsPartial(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	require.True(t, box.putItem(testPiece("base", 10, 20, 10, 1), Pivot{}))

	require.InDelta(t, 0.5, box.supportRatio(Pivot{0, 0, 10}, Dimension{20, 20, 10}), 1e-9)
}

func TestBox_SupportRatioSumsAcrossItems(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	require.True(t, box.putItem(testPiece("left", 10, 20, 10, 1), Pivot{}))
	require.True(t, box.putItem(testPiece("right", 10, 20, 10, 1), Pivot{10, 0, 0}))

	require.InDelta(t, 1.0, box.supportRatio(Pivot{0, 0, 10}, Dimension{20, 20, 10}), 1e-9)
}

func TestBox_SupportRatioIgnoresItemsAtOtherHeights(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	require.True(t, box.putItem(testPiece("low", 20, 20, 5, 1), Pivot{}))

	require.InDelta(t, 0.0, box.supportRatio(Pivot{0, 0, 10}, Dimension{20, 20, 10}), 1e-9)
}

func TestBox_IsStableOrientation(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)

	require.True(t, box.isStableOrientation(Dimension{50, 50, 10}), "a flat slab is stable")
	require.False(t, box.isStableOrientation(Dimension{5, 50, 90}), "a thin upright is not")
	require.True(t, box.isStableOrientation(Dimension{5, 50, 100}), "spanning the full depth is stable")
	require.True(t, box.isStableOrientation(Dimension{50, 50, 0}), "a degenerate depth is stable")
}

func TestBox_HasStableOrientation(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)

	require.True(t, box.hasStableOrientation(testPiece("slab", 50, 50, 10, 1)))
	require.True(t, box.hasStableOrientation(testPiece("pencil", 3, 3, 90, 1)),
		"a pencil can lie down, which is stable")
	require.True(t, box.hasStableOrientation(testPiece("pillar", 3, 3, 100, 1)),
		"a pillar spanning the box is stable")

	narrow := openBox("narrow", 5, 5, 100, 1000)
	require.False(t, narrow.hasStableOrientation(testPiece("pencil", 3, 3, 90, 1)),
		"in a narrow box the pencil can only stand, and standing is not stable")
}

func TestBestPlacement_PrefersAStableOrientation(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)

	_, rotation, found := bestPlacement(box, testPiece("slab", 60, 40, 10, 1))
	require.True(t, found)
	require.InDelta(t, 10.0, rotatedDimension(testPiece("slab", 60, 40, 10, 1).Item, rotation)[DepthAxis], 1e-9,
		"the item lies flat rather than standing on a narrow face")
}

func TestBestPlacement_PacksAnItemWithNoStableOrientation(t *testing.T) {
	t.Parallel()

	box := openBox("narrow", 5, 5, 100, 1000)

	_, _, found := bestPlacement(box, testPiece("pencil", 3, 3, 90, 1))
	require.True(t, found, "an item with no stable orientation is still packed")
}

func TestBestPlacement_RejectsUnsupportedPlacement(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	box.rules = Rules{MinSupportRatio: 0.8, Placement: nil}

	require.True(t, box.putItem(testPiece("base", 20, 20, 10, 1), Pivot{}))

	position, _, found := bestPlacement(box, testPiece("wide", 60, 60, 10, 1))
	require.True(t, found)
	require.InDelta(t, 0.0, position[DepthAxis], 1e-9, "the item goes to the floor, not onto the small base")
}

func TestBestPlacement_AcceptsSufficientSupport(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	box.rules = Rules{MinSupportRatio: 0.5, Placement: nil}

	require.True(t, box.putItem(testPiece("base", 100, 100, 10, 1), Pivot{}))

	position, _, found := bestPlacement(box, testPiece("top", 40, 40, 10, 1))
	require.True(t, found)
	require.InDelta(t, 10.0, position[DepthAxis], 1e-9, "a fully supported placement on top is taken")
}

func TestFillGaps_RetriesUntilNothingMoves(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 30, 100000)
	box.rules = Rules{MinSupportRatio: 1, Placement: nil}

	base := testPiece("base", 60, 100, 20, 1)
	require.True(t, box.putItem(base, Pivot{}))

	lid := testPiece("lid", 100, 100, 10, 1)
	filler := testPiece("filler", 40, 100, 20, 1)

	require.InDelta(t, 0.6, box.supportRatio(Pivot{0, 0, 20}, Dimension{100, 100, 10}), 1e-9)

	remaining, err := fillGaps(t.Context(), box, []*piece{lid, filler}, false)
	require.NoError(t, err)
	require.Empty(t, remaining, "the second pass places the lid once the filler supports it")
	require.Len(t, box.items, 3)
	require.InDelta(t, 20.0, lid.position[DepthAxis], 1e-9)
}

func TestFillGaps_StopsWhenNothingFits(t *testing.T) {
	t.Parallel()

	box := openBox("box", 10, 10, 10, 1000)
	require.True(t, box.putItem(testPiece("full", 10, 10, 10, 1), Pivot{}))

	remaining, err := fillGaps(t.Context(), box, []*piece{
		testPiece("a", 5, 5, 5, 1),
		testPiece("b", 5, 5, 5, 1),
	}, false)
	require.NoError(t, err)
	require.Len(t, remaining, 2)
}
