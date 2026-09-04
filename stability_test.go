package boxpacker3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestPacker_SupportRatioIsRespected(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 200, 200, 200, 100000)}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("a", 90, 90, 40, 100),
		boxpacker3.NewItem("b", 80, 80, 30, 100),
		boxpacker3.NewItem("c", 70, 70, 30, 100),
		boxpacker3.NewItem("d", 60, 60, 25, 100),
		boxpacker3.NewItem("e", 50, 50, 20, 100),
		boxpacker3.NewItem("f", 40, 40, 20, 100),
		boxpacker3.NewItem("g", 30, 30, 15, 100),
	}

	packer := boxpacker3.NewPacker(boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: 1}))

	result, err := packer.Pack(t.Context(), boxes, items)
	require.NoError(t, err)

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			require.GreaterOrEqual(t, supportRatioOf(box, item), 1-1e-6,
				"item %s rests on %.3f of its base", item.ID(), supportRatioOf(box, item))
		}
	}
}

func TestPacker_SupportRatioIsClamped(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 10000)}
	items := []*boxpacker3.Item{boxpacker3.NewItem("a", 50, 50, 50, 100)}

	for _, ratio := range []float64{-5, 0, 0.5, 1, 7} {
		result, err := boxpacker3.NewPacker(boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: ratio})).
			Pack(t.Context(), boxes, items)
		require.NoError(t, err)
		require.Empty(t, result.Unpacked)
	}
}

func TestPacker_UnstableOrientationsAreAvoided(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 200, 200, 200, 10000)}
	items := []*boxpacker3.Item{boxpacker3.NewItem("slab", 120, 100, 15, 100)}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Empty(t, result.Unpacked)

	packed := result.Boxes[0].Items
	require.Len(t, packed, 1)
	require.InDelta(t, 15.0, packed[0].Dimension[2], 1e-9, "the slab lies flat")
}

func supportRatioOf(box boxpacker3.PackedBox, item boxpacker3.PackedItem) float64 {
	position := item.Position
	dimension := item.Dimension

	base := dimension[0] * dimension[1]
	if base <= 0 {
		return 1
	}

	if position[2] <= 1e-9 {
		return 1
	}

	supported := 0.0

	for _, other := range box.Items {
		if other.Item == item.Item && other.Index == item.Index {
			continue
		}

		otherPosition := other.Position
		otherDimension := other.Dimension

		if diff := otherPosition[2] + otherDimension[2] - position[2]; diff > 1e-6 || diff < -1e-6 {
			continue
		}

		width := overlapOf(position[0], dimension[0], otherPosition[0], otherDimension[0])
		height := overlapOf(position[1], dimension[1], otherPosition[1], otherDimension[1])
		supported += width * height
	}

	return supported / base
}

func overlapOf(aStart, aSize, bStart, bSize float64) float64 {
	return max(0, min(aStart+aSize, bStart+bSize)-max(aStart, bStart))
}

func TestPacker_Property_SupportRatio(t *testing.T) {
	t.Parallel()

	for _, strategy := range allStrategies() {
		t.Run(strategyName(strategy), func(t *testing.T) {
			t.Parallel()

			boxes := []*boxpacker3.Box{
				boxpacker3.NewBox("small", 120, 120, 120, 100000),
				boxpacker3.NewBox("large", 240, 240, 240, 100000),
			}

			items := []*boxpacker3.Item{
				boxpacker3.NewItem("a", 100, 90, 40, 100),
				boxpacker3.NewItem("b", 80, 80, 30, 100),
				boxpacker3.NewItem("c", 70, 60, 30, 100),
				boxpacker3.NewItem("d", 60, 50, 25, 100),
				boxpacker3.NewItem("e", 50, 40, 20, 100),
				boxpacker3.NewItem("f", 40, 30, 20, 100),
				boxpacker3.NewItem("g", 30, 20, 15, 100),
				boxpacker3.NewItem("h", 25, 25, 25, 100),
			}

			packer := boxpacker3.NewPacker(
				boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(strategy.Order, strategy.Selection)),
				boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: 1}),
			)

			result, err := packer.Pack(t.Context(), boxes, items)
			require.NoError(t, err)

			validatePackingInvariants(t, result)

			for _, box := range result.Boxes {
				for _, item := range box.Items {
					require.GreaterOrEqual(t, supportRatioOf(box, item), 1-1e-6,
						"item %s in box %s rests on %.3f of its base",
						item.ID(), box.ID(), supportRatioOf(box, item))
				}
			}
		})
	}
}
