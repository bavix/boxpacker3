package boxpacker3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

const boxCarton = "carton"

func TestItemSpec_Valid(t *testing.T) {
	t.Parallel()

	item, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
		ID:     "widget",
		Width:  10,
		Height: 20,
		Depth:  30,
		Weight: 5,
	})
	require.NoError(t, err)
	require.Equal(t, "widget", item.ID())
	require.InDelta(t, 6000.0, item.Volume(), 1e-9)
	require.Equal(t, boxpacker3.RotationBestFit, item.Rotation())
}

func TestItemSpec_Rejects(t *testing.T) {
	t.Parallel()

	base := boxpacker3.ItemSpec{ID: "widget", Width: 10, Height: 10, Depth: 10, Weight: 1}

	cases := map[string]struct {
		mutate func(boxpacker3.ItemSpec) boxpacker3.ItemSpec
		want   error
	}{
		"negative width": {func(s boxpacker3.ItemSpec) boxpacker3.ItemSpec {
			s.Width = -1

			return s
		}, boxpacker3.ErrInvalidDimension},
		"zero height": {func(s boxpacker3.ItemSpec) boxpacker3.ItemSpec {
			s.Height = 0

			return s
		}, boxpacker3.ErrInvalidDimension},
		"zero depth": {func(s boxpacker3.ItemSpec) boxpacker3.ItemSpec {
			s.Depth = 0

			return s
		}, boxpacker3.ErrInvalidDimension},
		"negative weight": {func(s boxpacker3.ItemSpec) boxpacker3.ItemSpec {
			s.Weight = -1

			return s
		}, boxpacker3.ErrInvalidWeight},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			item, err := boxpacker3.NewItemFromSpec(testCase.mutate(base))
			require.ErrorIs(t, err, testCase.want)
			require.Nil(t, item)
		})
	}
}

func TestBoxSpec_InnerDefaultsToOuter(t *testing.T) {
	t.Parallel()

	box, err := boxpacker3.NewBoxFromSpec(boxpacker3.BoxSpec{
		ID:         boxCarton,
		OuterWidth: 100, OuterHeight: 80, OuterDepth: 60,
		MaxWeight: 1000,
	})
	require.NoError(t, err)

	require.InDelta(t, 100.0, box.Width(), 1e-9)
	require.InDelta(t, 100.0, box.OuterWidth(), 1e-9)
	require.InDelta(t, 80.0, box.OuterHeight(), 1e-9)
	require.InDelta(t, 60.0, box.OuterDepth(), 1e-9)
	require.InDelta(t, 0.0, box.EmptyWeight(), 1e-9)
}

func TestBoxSpec_WallsTakeSpace(t *testing.T) {
	t.Parallel()

	box, err := boxpacker3.NewBoxFromSpec(boxpacker3.BoxSpec{
		ID:         boxCarton,
		OuterWidth: 100, OuterHeight: 100, OuterDepth: 100,
		InnerWidth: 90, InnerHeight: 90, InnerDepth: 90,
		EmptyWeight: 250,
		MaxWeight:   1000,
	})
	require.NoError(t, err)

	require.InDelta(t, 90.0, box.Width(), 1e-9, "packing uses the inner dimensions")
	require.InDelta(t, 100.0, box.OuterWidth(), 1e-9)
	require.InDelta(t, 90.0*90*90, box.Volume(), 1e-9)

	items := []*boxpacker3.Item{boxpacker3.NewItem("snug", 95, 50, 50, 1)}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)
	require.Len(t, result.Unpacked, 1, "an item wider than the inner width does not fit")
}

func TestBoxSpec_TareCountsAgainstTheLimit(t *testing.T) {
	t.Parallel()

	box, err := boxpacker3.NewBoxFromSpec(boxpacker3.BoxSpec{
		ID:         boxCarton,
		OuterWidth: 100, OuterHeight: 100, OuterDepth: 100,
		EmptyWeight: 200,
		MaxWeight:   1000,
	})
	require.NoError(t, err)

	require.InDelta(t, 200.0, box.EmptyWeight(), 1e-9, "an empty box weighs its tare")
	require.InDelta(t, 800.0, box.MaxWeight()-box.EmptyWeight(), 1e-9)

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("a", 10, 10, 10, 500),
		boxpacker3.NewItem("b", 10, 10, 10, 400),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)

	packed := result.Boxes[0]
	require.InDelta(t, 500.0, packed.Stats.ItemsWeight, 1e-9)
	require.InDelta(t, 700.0, packed.Stats.GrossWeight, 1e-9, "tare plus contents")
	require.Len(t, result.Unpacked, 1, "the second item would exceed the gross limit")
}

func TestBoxSpec_Rejects(t *testing.T) {
	t.Parallel()

	base := boxpacker3.BoxSpec{
		ID:         boxCarton,
		OuterWidth: 100, OuterHeight: 100, OuterDepth: 100,
		MaxWeight: 1000,
	}

	cases := map[string]struct {
		mutate func(boxpacker3.BoxSpec) boxpacker3.BoxSpec
		want   error
	}{
		"zero outer width": {func(s boxpacker3.BoxSpec) boxpacker3.BoxSpec {
			s.OuterWidth = 0

			return s
		}, boxpacker3.ErrInvalidDimension},
		"negative outer depth": {func(s boxpacker3.BoxSpec) boxpacker3.BoxSpec {
			s.OuterDepth = -1

			return s
		}, boxpacker3.ErrInvalidDimension},
		"inner wider than outer": {func(s boxpacker3.BoxSpec) boxpacker3.BoxSpec {
			s.InnerWidth = 200

			return s
		}, boxpacker3.ErrInnerExceedsOuter},
		"negative tare": {func(s boxpacker3.BoxSpec) boxpacker3.BoxSpec {
			s.EmptyWeight = -1

			return s
		}, boxpacker3.ErrInvalidWeight},
		"limit below tare": {func(s boxpacker3.BoxSpec) boxpacker3.BoxSpec {
			s.EmptyWeight = 2000

			return s
		}, boxpacker3.ErrInvalidWeight},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			box, err := boxpacker3.NewBoxFromSpec(testCase.mutate(base))
			require.ErrorIs(t, err, testCase.want)
			require.Nil(t, box)
		})
	}
}

func TestRotation_OrientationCounts(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		spec boxpacker3.ItemSpec
		want int
	}{
		"free item": {boxpacker3.ItemSpec{
			ID: "a", Width: 1, Height: 2, Depth: 3, Weight: 1,
			Rotation: boxpacker3.RotationBestFit,
		}, 6},
		"keep flat": {boxpacker3.ItemSpec{
			ID: "b", Width: 1, Height: 2, Depth: 3, Weight: 1,
			Rotation: boxpacker3.RotationKeepFlat,
		}, 2},
		"never": {boxpacker3.ItemSpec{
			ID: "c", Width: 1, Height: 2, Depth: 3, Weight: 1,
			Rotation: boxpacker3.RotationNever,
		}, 1},
		"cube collapses": {boxpacker3.ItemSpec{
			ID: "d", Width: 2, Height: 2, Depth: 2, Weight: 1,
			Rotation: boxpacker3.RotationBestFit,
		}, 1},
		"square face collapses": {boxpacker3.ItemSpec{
			ID: "e", Width: 2, Height: 2, Depth: 5, Weight: 1,
			Rotation: boxpacker3.RotationBestFit,
		}, 3},
		"two vertical axes": {boxpacker3.ItemSpec{
			ID: "f", Width: 1, Height: 2, Depth: 3, Weight: 1,
			VerticalAxes: []boxpacker3.Axis{boxpacker3.WidthAxis, boxpacker3.DepthAxis},
		}, 4},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			item, err := boxpacker3.NewItemFromSpec(testCase.spec)
			require.NoError(t, err)
			require.Len(t, item.Orientations(), testCase.want)
		})
	}
}

func TestRotation_VerticalAxesRestrictWhichEdgePointsUp(t *testing.T) {
	t.Parallel()

	item, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
		ID: "plank", Width: 10, Height: 20, Depth: 30, Weight: 1,
		VerticalAxes: []boxpacker3.Axis{boxpacker3.WidthAxis},
	})
	require.NoError(t, err)

	for _, dimension := range item.Orientations() {
		require.InDelta(t, 10.0, dimension[2], 1e-9,
			"only the width edge may point upwards")
	}
}

func TestBox_DirectPlacementRespectsRotation(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("deep", 100, 100, 100, 10000)
	sheet := boxpacker3.NewItem2D("sheet", 80, 60, 10)

	result, err := boxpacker3.NewPacker().Pack(t.Context(), []*boxpacker3.Box{box}, []*boxpacker3.Item{sheet})
	require.NoError(t, err)
	require.Len(t, result.Boxes[0].Items, 1)
	require.InDelta(t, 1.0, result.Boxes[0].Items[0].Dimension[2], 1e-9,
		"a flat item keeps its depth even in a deep box")
}

func TestItemSpec_QuantityLaysOutEveryCopy(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("crate", 100, 100, 100, 1e9)
	mug := mustItem(t, boxpacker3.ItemSpec{
		ID: "mug", Width: 45, Height: 45, Depth: 50, Weight: 350, Quantity: 12,
	})

	require.Equal(t, 12, mug.Quantity())

	result, err := boxpacker3.NewPacker().Pack(t.Context(), []*boxpacker3.Box{box}, []*boxpacker3.Item{mug})
	require.NoError(t, err)

	indexes := map[int]int{}

	for _, packed := range result.Boxes {
		for _, item := range packed.Items {
			require.Same(t, mug, item.Item)

			indexes[item.Index]++
		}
	}

	for _, item := range result.Unpacked {
		require.Same(t, mug, item.Item)

		indexes[item.Index]++
	}

	require.Len(t, indexes, 12, "twelve copies, twelve indexes")

	for index := range 12 {
		require.Equal(t, 1, indexes[index], "copy %d appears once", index)
	}

	_, err = boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
		ID: "bad", Width: 1, Height: 1, Depth: 1, Weight: 1, Quantity: -1,
	})
	require.ErrorIs(t, err, boxpacker3.ErrInvalidQuantity)
}
