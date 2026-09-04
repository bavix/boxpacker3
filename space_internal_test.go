package boxpacker3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBox_PointsStartAtTheOrigin(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	require.Equal(t, []Pivot{{}}, box.points)
}

func TestBox_PlacementConsumesThePoint(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	require.True(t, box.putItem(testPiece("a", 40, 40, 40, 1), Pivot{}))
	require.NotContains(t, box.points, Pivot{})
	require.Contains(t, box.points, Pivot{40, 0, 0})
	require.Contains(t, box.points, Pivot{0, 40, 0})
	require.Contains(t, box.points, Pivot{0, 0, 40})
}

func TestBox_PointsStayInsideTheBox(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	require.True(t, box.putItem(testPiece("a", 100, 40, 40, 1), Pivot{}))

	for _, point := range box.points {
		require.Less(t, point[WidthAxis], 100.0)
	}
}

func TestBox_PointsAreOrderedBottomFirst(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	require.True(t, box.putItem(testPiece("a", 40, 40, 40, 1), Pivot{}))
	require.True(t, box.putItem(testPiece("b", 40, 40, 40, 1), Pivot{40, 0, 0}))

	for i := 1; i < len(box.points); i++ {
		require.LessOrEqual(t, comparePoints(box.points[i-1], box.points[i]), 0)
	}
}

func TestBox_ProjectRestsOnTheNearestFace(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	require.True(t, box.putItem(testPiece("shelf", 50, 50, 20, 1), Pivot{0, 0, 50}))

	require.InDelta(t, 70.0, box.project(Pivot{10, 10, 90}, DepthAxis), 1e-9)
	require.InDelta(t, 0.0, box.project(Pivot{90, 90, 90}, DepthAxis), 1e-9)
}

func TestBox_ResidualGapMeasuresFreeSpace(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	item := testPiece("blocker", 20, 100, 100, 1)
	require.True(t, box.putItem(item, Pivot{60, 0, 0}))

	require.InDelta(t, 40.0, box.residualGap(Pivot{}, Dimension{20, 20, 20}, WidthAxis), 1e-9)
	require.InDelta(t, 80.0, box.residualGap(Pivot{}, Dimension{20, 20, 20}, HeightAxis), 1e-9)
}

func TestBox_FitsAtRejectsOverlapAndOverflow(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	require.True(t, box.putItem(testPiece("a", 50, 50, 50, 1), Pivot{}))

	require.False(t, box.fitsAt(Pivot{}, Dimension{10, 10, 10}), "overlaps the placed item")
	require.False(t, box.fitsAt(Pivot{95, 0, 0}, Dimension{10, 10, 10}), "sticks out of the box")
	require.False(t, box.fitsAt(Pivot{-1, 0, 0}, Dimension{10, 10, 10}), "starts outside the box")
	require.True(t, box.fitsAt(Pivot{50, 0, 0}, Dimension{10, 10, 10}), "sits against the placed item")
}

func TestOverlaps_TouchingFacesDoNotOverlap(t *testing.T) {
	t.Parallel()

	require.False(t, overlaps(Pivot{}, Dimension{10, 10, 10}, Pivot{10, 0, 0}, Dimension{10, 10, 10}))
	require.True(t, overlaps(Pivot{}, Dimension{10, 10, 10}, Pivot{9, 0, 0}, Dimension{10, 10, 10}))
	require.False(t, overlaps(Pivot{}, Dimension{0, 10, 10}, Pivot{}, Dimension{10, 10, 10}),
		"a degenerate item never overlaps")
}

func TestPointInside(t *testing.T) {
	t.Parallel()

	require.True(t, pointInside(Pivot{5, 5, 5}, Pivot{}, Dimension{10, 10, 10}))
	require.True(t, pointInside(Pivot{}, Pivot{}, Dimension{10, 10, 10}))
	require.False(t, pointInside(Pivot{10, 5, 5}, Pivot{}, Dimension{10, 10, 10}),
		"a point on the far face is outside")
}

func TestBox_ResetRestoresTheOrigin(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	require.True(t, box.putItem(testPiece("a", 40, 40, 40, 1), Pivot{}))

	box.reset()
	require.Equal(t, []Pivot{{}}, box.points)
}

func TestBox_CloneCopiesThePoints(t *testing.T) {
	t.Parallel()

	box := openBox("box", 100, 100, 100, 1000)
	require.True(t, box.putItem(testPiece("a", 40, 40, 40, 1), Pivot{}))

	clone := clonePtr(box)
	clone.reset()

	require.NotEqual(t, []Pivot{{}}, box.points, "the original keeps its points")
}

func TestBox_CopyRemembersLoadLimits(t *testing.T) {
	t.Parallel()

	box := openBox("crate", 100, 100, 400, 1e9)

	carrier, err := NewItemFromSpec(ItemSpec{
		ID: "carrier", Width: 100, Height: 100, Depth: 100, Weight: 1000, MaxLoadOnTop: 500,
	})
	if err != nil {
		t.Fatal(err)
	}

	box.place(pieceOf(carrier), Pivot{0, 0, 0}, OrientationWHD)

	copied := clonePtr(box)
	if !copied.loadLimited {
		t.Fatal("the copy forgot that its goods state a load limit")
	}

	heavy := testPiece("heavy", 100, 100, 100, 900)
	if copied.acceptsPlacement(heavy, Pivot{0, 0, 100}, Dimension{100, 100, 100}, true) {
		t.Error("the copy took a load its goods cannot carry")
	}
}
