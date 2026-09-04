package boxpacker3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

const classGlass = "glass"

func TestFilter_CanFitVolume(t *testing.T) {
	t.Parallel()

	box := openBox("b", 10, 10, 10, 100)

	require.True(t, box.canFitVolume(testPiece("exact", 10, 10, 10, 1)))
	require.False(t, box.canFitVolume(testPiece("over", 10, 10, 10.1, 1)))

	box.insert(testPiece("half", 10, 10, 5, 1))

	require.True(t, box.canFitVolume(testPiece("rest", 10, 10, 5, 1)))
	require.False(t, box.canFitVolume(testPiece("toomuch", 10, 10, 5.01, 1)))
}

func TestFilter_CanFitWeight_CountsTheTare(t *testing.T) {
	t.Parallel()

	boxSpec, err := NewBoxFromSpec(BoxSpec{
		ID: "b", OuterWidth: 10, OuterHeight: 10, OuterDepth: 10, MaxWeight: 10, EmptyWeight: 3,
	})
	require.NoError(t, err)

	box := openUnder(boxSpec)

	require.True(t, box.canFitWeight(testPiece("fits", 1, 1, 1, 7)))
	require.False(t, box.canFitWeight(testPiece("over", 1, 1, 1, 7.1)))

	box.insert(testPiece("first", 1, 1, 1, 4))

	require.True(t, box.canFitWeight(testPiece("rest", 1, 1, 1, 3)))
	require.False(t, box.canFitWeight(testPiece("toomuch", 1, 1, 1, 3.1)))
}

func TestFilter_TakesTheClass_DedicatedBox(t *testing.T) {
	t.Parallel()

	boxSpec, err := NewBoxFromSpec(BoxSpec{
		ID: "cold", OuterWidth: 10, OuterHeight: 10, OuterDepth: 10, MaxWeight: 100, Accepts: []string{"frozen"},
	})
	require.NoError(t, err)

	box := openUnder(boxSpec)

	frozen := classed(t, "f", "frozen", nil)
	dry := classed(t, "d", "dry", nil)
	plain := testPiece("p", 1, 1, 1, 1)

	require.True(t, box.takesTheClass(frozen))
	require.False(t, box.takesTheClass(dry))
	require.False(t, box.takesTheClass(plain))
}

func TestFilter_TakesTheClass_SeparationRunsBothWays(t *testing.T) {
	t.Parallel()

	box := openBox("b", 10, 10, 10, 100)
	box.insert(classed(t, "food", "food", []string{"solvent"}))

	require.False(t, box.takesTheClass(classed(t, "chem", "chemical", []string{"food"})))
	require.False(t, box.takesTheClass(classed(t, "solv", "solvent", nil)))
	require.True(t, box.takesTheClass(classed(t, "paper", "paper", nil)))
	require.True(t, box.takesTheClass(testPiece("plain", 1, 1, 1, 1)))
}

func TestFilter_CouldEverHold(t *testing.T) {
	t.Parallel()

	boxSpec, err := NewBoxFromSpec(BoxSpec{
		ID: "b", OuterWidth: 10, OuterHeight: 20, OuterDepth: 5, MaxWeight: 0.3, EmptyWeight: 0.1,
	})
	require.NoError(t, err)

	box := openUnder(boxSpec)

	require.True(t, box.couldEverHold(testPiece("exact-weight", 1, 1, 1, 0.2)))
	require.False(t, box.couldEverHold(testPiece("over-weight", 1, 1, 1, 0.21)))

	turned, err := NewItemFromSpec(ItemSpec{ID: "turned", Width: 5, Height: 10, Depth: 20, Weight: 0.1})
	require.NoError(t, err)
	require.True(t, box.couldEverHold(pieceOf(turned)))

	upright, err := NewItemFromSpec(ItemSpec{
		ID: "upright", Width: 5, Height: 10, Depth: 20, Weight: 0.1, Rotation: RotationNever,
	})
	require.NoError(t, err)
	require.False(t, box.couldEverHold(pieceOf(upright)))

	dedicatedSpec, err := NewBoxFromSpec(BoxSpec{
		ID: "d", OuterWidth: 10, OuterHeight: 10, OuterDepth: 10, MaxWeight: 100, Accepts: []string{classGlass},
	})
	require.NoError(t, err)

	dedicated := openUnder(dedicatedSpec)
	require.False(t, dedicated.couldEverHold(testPiece("plain", 1, 1, 1, 1)))
	require.True(t, dedicated.couldEverHold(classed(t, "g", classGlass, nil)))
}

func TestFilter_CouldEverHoldAgreesWithCanFitWeight(t *testing.T) {
	t.Parallel()

	for _, tare := range []float64{0.1, 0.3, 0.7, 1.1} {
		boxSpec, err := NewBoxFromSpec(BoxSpec{
			ID: "b", OuterWidth: 10, OuterHeight: 10, OuterDepth: 10, MaxWeight: tare + 0.2, EmptyWeight: tare,
		})
		require.NoError(t, err)

		box := openUnder(boxSpec)

		item := testPiece("i", 1, 1, 1, 0.2)
		require.Equal(t, box.canFitWeight(item), box.couldEverHold(item), "tare %g", tare)
	}
}

func TestFilter_BearsTheLoad_CapOnTheItemBelow(t *testing.T) {
	t.Parallel()

	box := openBox("b", 10, 10, 30, 100)
	base := loadLimited(t, "base", 5, false)
	box.place(base, Pivot{}, OrientationWHD)

	top := Pivot{0, 0, 10}
	cube := Dimension{10, 10, 10}

	require.True(t, box.bearsTheLoad(testPiece("five", 10, 10, 10, 5), top, cube))
	require.False(t, box.bearsTheLoad(testPiece("six", 10, 10, 10, 6), top, cube))

	box.place(testPiece("three", 10, 10, 10, 3), top, OrientationWHD)

	higher := Pivot{0, 0, 20}
	require.True(t, box.bearsTheLoad(testPiece("two", 10, 10, 10, 2), higher, cube))
	require.False(t, box.bearsTheLoad(testPiece("two-and-a-bit", 10, 10, 10, 2.1), higher, cube))
}

func TestFilter_BearsTheLoad_NothingOnTop(t *testing.T) {
	t.Parallel()

	box := openBox("b", 10, 10, 30, 100)
	box.place(loadLimited(t, "clear", 0, true), Pivot{}, OrientationWHD)

	require.False(t, box.bearsTheLoad(testPiece("weightless", 10, 10, 10, 0), Pivot{0, 0, 10}, Dimension{10, 10, 10}))
	require.True(t, box.bearsTheLoad(testPiece("beside", 10, 10, 10, 50), Pivot{0, 0, 0}, Dimension{10, 10, 10}))
}

func TestFilter_BearsTheLoad_PlacingUnderneath(t *testing.T) {
	t.Parallel()

	box := openBox("b", 10, 10, 30, 100)
	box.place(testPiece("above", 10, 10, 10, 8), Pivot{0, 0, 10}, OrientationWHD)

	floor := Pivot{}
	cube := Dimension{10, 10, 10}

	require.False(t, box.bearsTheLoad(loadLimited(t, "weak", 5, false), floor, cube))
	require.True(t, box.bearsTheLoad(loadLimited(t, "strong", 8, false), floor, cube))
	require.False(t, box.bearsTheLoad(loadLimited(t, "clear", 0, true), floor, cube))
	require.True(t, box.bearsTheLoad(testPiece("plain", 10, 10, 10, 1), floor, cube))
}

type gateCase struct {
	name  string
	build func(t *testing.T) (*Container, *piece, Pivot, Dimension)
}

func closedGates() []gateCase {
	return append(geometryGates(), cargoGates()...)
}

func geometryGates() []gateCase {
	cube := Dimension{10, 10, 10}
	onTheFloor := Pivot{}

	return []gateCase{
		{"unstable orientation while a stable one fits", func(t *testing.T) (*Container, *piece, Pivot, Dimension) {
			t.Helper()

			return openBox("b", 30, 30, 30, 100), testPiece("tall", 2, 2, 20, 1), onTheFloor, Dimension{2, 2, 20}
		}},
		{"outside the box", func(t *testing.T) (*Container, *piece, Pivot, Dimension) {
			t.Helper()

			return openBox("b", 15, 15, 15, 100), testPiece("c", 10, 10, 10, 1), Pivot{6, 0, 0}, cube
		}},
		{"overlapping what is there", func(t *testing.T) (*Container, *piece, Pivot, Dimension) {
			t.Helper()

			box := openBox("b", 30, 30, 30, 100)
			box.place(testPiece("first", 10, 10, 10, 1), onTheFloor, OrientationWHD)

			return box, testPiece("c", 10, 10, 10, 1), Pivot{5, 0, 0}, cube
		}},
		{"too little support", func(t *testing.T) (*Container, *piece, Pivot, Dimension) {
			t.Helper()

			box := openBox("b", 30, 30, 30, 100)
			box.rules.MinSupportRatio = 0.5
			box.place(testPiece("first", 10, 10, 10, 1), onTheFloor, OrientationWHD)

			return box, testPiece("c", 10, 10, 10, 1), Pivot{5, 5, 10}, cube
		}},
	}
}

func cargoGates() []gateCase {
	cube := Dimension{10, 10, 10}
	onTheFloor := Pivot{}

	return []gateCase{
		{"too heavy for what is below", func(t *testing.T) (*Container, *piece, Pivot, Dimension) {
			t.Helper()

			box := openBox("b", 30, 30, 30, 100)
			box.place(loadLimited(t, "clear", 0, true), onTheFloor, OrientationWHD)

			return box, testPiece("c", 10, 10, 10, 1), Pivot{0, 0, 10}, cube
		}},
		{"a class the box refuses", func(t *testing.T) (*Container, *piece, Pivot, Dimension) {
			t.Helper()

			boxSpec, err := NewBoxFromSpec(BoxSpec{
				ID: "b", OuterWidth: 30, OuterHeight: 30, OuterDepth: 30, MaxWeight: 100, Accepts: []string{classGlass},
			})
			require.NoError(t, err)

			box := openUnder(boxSpec)

			return box, testPiece("c", 10, 10, 10, 1), onTheFloor, cube
		}},
		{"a class kept apart from what is there", func(t *testing.T) (*Container, *piece, Pivot, Dimension) {
			t.Helper()

			box := openBox("b", 30, 30, 30, 100)
			box.place(classed(t, "food", "food", nil), onTheFloor, OrientationWHD)

			return box, classed(t, "chem", "chemical", []string{"food"}), Pivot{10, 0, 0}, cube
		}},
		{"a constraint that says no", func(t *testing.T) (*Container, *piece, Pivot, Dimension) {
			t.Helper()

			box := openBox("b", 30, 30, 30, 100)
			box.rules.Placement = []PlacementRule{
				PlacementRuleFunc(func(*Container, *Item, Pivot, Dimension) bool { return false }),
			}

			return box, testPiece("c", 10, 10, 10, 1), onTheFloor, cube
		}},
	}
}

func TestFilter_AcceptsPlacement_EachGateAlone(t *testing.T) {
	t.Parallel()

	for _, tc := range closedGates() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			box, item, position, dimension := tc.build(t)
			require.False(t, box.acceptsPlacement(item, position, dimension, false))
		})
	}

	t.Run("nothing in the way", func(t *testing.T) {
		t.Parallel()

		box := openBox("b", 30, 30, 30, 100)
		box.rules.MinSupportRatio = 0.5
		box.place(testPiece("first", 10, 10, 10, 1), Pivot{}, OrientationWHD)

		require.True(t, box.acceptsPlacement(testPiece("c", 10, 10, 10, 1), Pivot{0, 0, 10}, Dimension{10, 10, 10}, false))
	})
}

func TestFilter_DirectPlacementRunsTheSameGate(t *testing.T) {
	t.Parallel()

	t.Run("refuses to rest on a clear item", func(t *testing.T) {
		t.Parallel()

		box := openBox("b", 10, 10, 30, 100)
		require.True(t, box.putItem(loadLimited(t, "clear", 0, true), Pivot{}))
		require.False(t, box.putItem(testPiece("c", 10, 10, 10, 1), Pivot{0, 0, 10}))
	})

	t.Run("refuses an unsupported placement", func(t *testing.T) {
		t.Parallel()

		box := openBox("b", 30, 30, 30, 100)
		box.rules.MinSupportRatio = 0.5
		require.True(t, box.putItem(testPiece("first", 10, 10, 10, 1), Pivot{}))
		require.False(t, box.putItem(testPiece("c", 10, 10, 10, 1), Pivot{5, 5, 10}))
		require.True(t, box.putItem(testPiece("d", 10, 10, 10, 1), Pivot{0, 0, 10}))
	})

	t.Run("stands a tall item on its side", func(t *testing.T) {
		t.Parallel()

		box := openBox("b", 30, 30, 30, 100)
		tall := testPiece("tall", 2, 2, 20, 1)
		require.True(t, box.putItem(tall, Pivot{}))
		require.True(t, box.isStableOrientation(tall.dimension()))
	})

	t.Run("refuses a class the box does not take", func(t *testing.T) {
		t.Parallel()

		boxSpec, err := NewBoxFromSpec(BoxSpec{
			ID: "b", OuterWidth: 30, OuterHeight: 30, OuterDepth: 30, MaxWeight: 100, Accepts: []string{classGlass},
		})
		require.NoError(t, err)

		box := openUnder(boxSpec)
		require.False(t, box.putItem(testPiece("c", 10, 10, 10, 1), Pivot{}))
		require.True(t, box.putItem(classed(t, "g", classGlass, nil), Pivot{}))
	})
}

func TestFullestBox_TellsBoxesApartByKindNotID(t *testing.T) {
	t.Parallel()

	big, err := NewBoxFromSpec(BoxSpec{
		ID: "same", OuterWidth: 100, OuterHeight: 100, OuterDepth: 100, MaxWeight: 100, Quantity: 2,
	})
	require.NoError(t, err)

	boxes := []*Box{NewBox("same", 10, 10, 10, 100), NewBox("other", 60, 60, 60, 100), big}
	items := []*piece{
		testPiece("big-1", 70, 70, 70, 1), testPiece("big-2", 70, 70, 70, 1), testPiece("tiny", 5, 5, 5, 1),
	}

	result, err := NewGreedy(OrderDecreasing, SelectFullestBox).Pack(context.Background(), problemOf(boxes, items))
	require.NoError(t, err)
	require.Empty(t, result.unfit)
	require.Equal(t, 2, countUsedBoxes(result.boxes))
}

func classed(t *testing.T, id, class string, separateFrom []string) *piece {
	t.Helper()

	item, err := NewItemFromSpec(ItemSpec{
		ID: id, Width: 10, Height: 10, Depth: 10, Weight: 1, Class: class, SeparateFrom: separateFrom,
	})
	require.NoError(t, err)

	return pieceOf(item)
}

func loadLimited(t *testing.T, id string, maxLoadOnTop float64, nothingOnTop bool) *piece {
	t.Helper()

	item, err := NewItemFromSpec(ItemSpec{
		ID: id, Width: 10, Height: 10, Depth: 10, Weight: 1, MaxLoadOnTop: maxLoadOnTop, NothingOnTop: nothingOnTop,
	})
	require.NoError(t, err)

	return pieceOf(item)
}

func TestFilter_SeededFillRunsTheGate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("refuses an unstable seed while a stable one exists", func(t *testing.T) {
		t.Parallel()

		box := openBox("b", 30, 30, 30, 100)
		items := []*piece{testPiece("tall", 2, 2, 20, 1), testPiece("c", 5, 5, 5, 1)}

		fill, err := seededFillAttempt(ctx, box, items, OrientationWHD)
		require.NoError(t, err)
		require.Nil(t, fill)

		fill, err = seededFillAttempt(ctx, box, items, OrientationDHW)
		require.NoError(t, err)
		require.NotNil(t, fill)
	})

	t.Run("refuses a seed the caller's rule forbids", func(t *testing.T) {
		t.Parallel()

		box := openBox("b", 30, 30, 30, 100)
		box.rules.Placement = []PlacementRule{
			PlacementRuleFunc(func(_ *Container, item *Item, _ Pivot, _ Dimension) bool {
				return item.id != "banned"
			}),
		}
		items := []*piece{testPiece("banned", 10, 10, 10, 1), testPiece("c", 5, 5, 5, 1)}

		fill, err := seededFillAttempt(ctx, box, items, OrientationWHD)
		require.NoError(t, err)
		require.Nil(t, fill)
	})
}
