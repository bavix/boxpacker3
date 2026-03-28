package boxpacker3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func reading(position Pivot, contact float64, exactAxes int, minGap, footprint float64) MeritScore {
	return MeritScore{
		Position: position, Contact: contact, Weighted: contact, Residual: Dimension{},
		ExactAxes: exactAxes, MinGap: minGap, Footprint: footprint,
	}
}

func TestMerit_ContactFirstPrefersTheLowerPosition(t *testing.T) {
	t.Parallel()

	low := reading(Pivot{0, 0, 0}, 0, 0, 100, 100)
	high := reading(Pivot{0, 0, 10}, 0, 3, 0, 1)

	require.True(t, ContactFirst.Better(low, high))
	require.False(t, ContactFirst.Better(high, low))
}

func TestMerit_ContactFirstTieBreaks(t *testing.T) {
	t.Parallel()

	base := reading(Pivot{}, 0, 1, 10, 100)

	exact := base
	exact.ExactAxes = 2
	require.True(t, ContactFirst.Better(exact, base), "more exact axes wins")

	tight := base
	tight.MinGap = 5
	require.True(t, ContactFirst.Better(tight, base), "smaller gap wins")

	small := base
	small.Footprint = 50
	require.True(t, ContactFirst.Better(small, base), "smaller footprint wins")

	require.False(t, ContactFirst.Better(base, base), "a score does not beat itself")
}

func TestMerit_ContactBeatsPosition(t *testing.T) {
	t.Parallel()

	touching := reading(Pivot{0, 0, 10}, 100, 0, 0, 1)
	loose := reading(Pivot{0, 0, 0}, 10, 0, 0, 1)

	require.True(t, ContactFirst.Better(touching, loose), "contact comes first")
	require.True(t, CornerFirst.Better(loose, touching), "the corner comes first instead")
}

func TestMerit_ResidualFitPrefersTheTighterHole(t *testing.T) {
	t.Parallel()

	tight := reading(Pivot{}, 0, 0, 0, 1)
	tight.Residual = Dimension{1, 1, 1}

	loose := reading(Pivot{}, 100, 0, 0, 1)
	loose.Residual = Dimension{10, 10, 10}

	require.True(t, ResidualFit.Better(tight, loose))
	require.False(t, ResidualFit.Better(loose, tight))
}

func TestMerit_EveryMeritIsNamedAndFound(t *testing.T) {
	t.Parallel()

	seen := map[string]bool{}

	for _, merit := range Merits() {
		require.NotEmpty(t, merit.Name())
		require.False(t, seen[merit.Name()], "two merits answer to %q", merit.Name())

		seen[merit.Name()] = true

		found, ok := MeritByName(merit.Name())
		require.True(t, ok)
		require.Equal(t, merit, found)
	}

	_, ok := MeritByName("no-such-merit")
	require.False(t, ok)
}

func TestRotatedDimension(t *testing.T) {
	t.Parallel()

	item := testPiece("a", 1, 2, 3, 0)

	require.Equal(t, Dimension{1, 2, 3}, rotatedDimension(item.Item, OrientationWHD))
	require.Equal(t, Dimension{2, 1, 3}, rotatedDimension(item.Item, OrientationHWD))
	require.Equal(t, Dimension{2, 3, 1}, rotatedDimension(item.Item, OrientationHDW))
	require.Equal(t, Dimension{3, 2, 1}, rotatedDimension(item.Item, OrientationDHW))
	require.Equal(t, Dimension{3, 1, 2}, rotatedDimension(item.Item, OrientationDWH))
	require.Equal(t, Dimension{1, 3, 2}, rotatedDimension(item.Item, OrientationWDH))
}

func TestBestPlacement_UsesTheLowestPoint(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	require.True(t, box.putItem(testPiece("a", 40, 40, 40, 1), Pivot{}))

	position, _, found := bestPlacement(box, testPiece("b", 10, 10, 10, 1))
	require.True(t, found)
	require.InDelta(t, 0.0, position[DepthAxis], 1e-9, "the new item stays on the floor")
}

func TestBestPlacement_ReportsNoRoom(t *testing.T) {
	t.Parallel()

	box := openBox("box", 10, 10, 10, 1000)
	require.True(t, box.putItem(testPiece("a", 10, 10, 10, 1), Pivot{}))

	_, _, found := bestPlacement(box, testPiece("b", 1, 1, 1, 1))
	require.False(t, found)
}

func TestBestPlacement_ChoosesAnExactOrientation(t *testing.T) {
	t.Parallel()

	box := openBox("box", 530, 380, 265, 20000)

	_, rotation, found := bestPlacement(box, testPiece("slab", 100, 380, 250, 1))
	require.True(t, found)
	require.Equal(t, Dimension{100, 380, 250}, rotatedDimension(testPiece("slab", 100, 380, 250, 1).Item, rotation),
		"the orientation that fills the height exactly is preferred")
}

func TestBestPlacement_IsDeterministic(t *testing.T) {
	t.Parallel()

	build := func() *Container {
		box := openBox("box", 100, 100, 100, 1000)
		require.True(t, box.putItem(testPiece("a", 30, 40, 50, 1), Pivot{}))
		require.True(t, box.putItem(testPiece("b", 20, 20, 20, 1), Pivot{30, 0, 0}))

		return box
	}

	firstPosition, firstRotation, found := bestPlacement(build(), testPiece("c", 25, 15, 35, 1))
	require.True(t, found)

	for range 20 {
		position, rotation, ok := bestPlacement(build(), testPiece("c", 25, 15, 35, 1))
		require.True(t, ok)
		require.Equal(t, firstPosition, position)
		require.Equal(t, firstRotation, rotation)
	}
}
