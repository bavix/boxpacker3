package boxpacker3_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

var errAlgoFailed = errors.New("algorithm failed")

type failingAlgo struct{ name string }

func (a *failingAlgo) Name() string { return a.name }

func (a *failingAlgo) Pack(_ context.Context, _ *boxpacker3.Problem) (*boxpacker3.Packing, error) {
	return nil, errAlgoFailed
}

func allStrategies() []boxpacker3.RuleSettings {
	return []boxpacker3.RuleSettings{
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectNextFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectWorstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectAlmostWorstFit},
	}
}

func TestPacker_NilBoxIsSkipped(t *testing.T) {
	t.Parallel()

	for _, strategy := range allStrategies() {
		t.Run(strategyName(strategy), func(t *testing.T) {
			t.Parallel()

			boxes := []*boxpacker3.Box{boxpacker3.NewBox("small", 10, 10, 10, 100), nil}
			items := []*boxpacker3.Item{boxpacker3.NewItem("oversized", 50, 50, 50, 1)}

			packer := rulePacker(strategy.Order, strategy.Selection)

			result, err := packer.Pack(t.Context(), boxes, items)
			require.NoError(t, err)
			require.Len(t, result.Unpacked, 1)

			for _, box := range result.Boxes {
				require.NotNil(t, box, "nil boxes must not reach the result")
			}
		})
	}
}

func TestGoals_NilBoxInResult(t *testing.T) {
	t.Parallel()

	parallel := boxpacker3.NewPortfolio(boxpacker3.FewestBoxes,
		boxpacker3.NewGreedy(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit),
		boxpacker3.NewGreedy(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit),
		boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit),
	)

	packer := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(parallel))

	result, err := packer.Pack(t.Context(),
		[]*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 1000), nil},
		[]*boxpacker3.Item{boxpacker3.NewItem("item", 10, 10, 10, 1)})

	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestGoals_NilCandidate(t *testing.T) {
	t.Parallel()

	goals := map[string]boxpacker3.Goal{}
	for _, goal := range boxpacker3.Goals() {
		goals[goal.Name()] = goal
	}

	for name, goal := range goals {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			incumbent := &boxpacker3.Result{Boxes: nil, Unpacked: nil, Report: boxpacker3.Report{}}

			require.Positive(t, goal.Compare(nil, incumbent), "an answer that is not there never wins")
			require.Negative(t, goal.Compare(incumbent, nil), "any answer beats none")
			require.Zero(t, goal.Compare(nil, nil))
		})
	}
}

func TestPacker_NilItemsNeverReachTheResult(t *testing.T) {
	t.Parallel()

	for _, strategy := range allStrategies() {
		t.Run(strategyName(strategy), func(t *testing.T) {
			t.Parallel()

			boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 1000)}
			items := []*boxpacker3.Item{
				boxpacker3.NewItem("a", 10, 10, 10, 1),
				nil,
				boxpacker3.NewItem("b", 10, 10, 10, 1),
				nil,
			}

			packer := rulePacker(strategy.Order, strategy.Selection)

			result, err := packer.Pack(t.Context(), boxes, items)
			require.NoError(t, err)

			packed := 0

			for _, box := range result.Boxes {
				for _, item := range box.Items {
					require.NotNil(t, item)

					packed++
				}
			}

			for _, item := range result.Unpacked {
				require.NotNil(t, item, "UnfitItems must never contain nil")
			}

			require.Equal(t, 2, packed+len(result.Unpacked),
				"every non-nil item is either packed or unfit")
		})
	}
}

func TestPacker_ExactFitWithFloatAccumulation(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("exact", 2.4, 1, 1, 100)}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("a", 0.8, 1, 1, 1),
		boxpacker3.NewItem("b", 0.8, 1, 1, 1),
		boxpacker3.NewItem("c", 0.8, 1, 1, 1),
	}

	packer := boxpacker3.NewPacker()

	result, err := packer.Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Empty(t, result.Unpacked, "an exact fill must not be rejected by rounding error")
}

func TestParallel_DeterministicTieBreak(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("first", 10, 10, 10, 100),
		boxpacker3.NewBox("second", 10, 10, 10, 100),
		boxpacker3.NewBox("third", 10, 10, 10, 100),
	}
	items := []*boxpacker3.Item{boxpacker3.NewItem("item", 1, 1, 1, 1)}

	strategy := boxpacker3.NewPortfolio(nil,
		&splitAlgo{name: "first", boxes: 1, leave: 0, pick: 0},
		&splitAlgo{name: "second", boxes: 1, leave: 0, pick: 1},
		&splitAlgo{name: "third", boxes: 1, leave: 0, pick: 2},
	)

	for range 200 {
		result := packed(t, boxpacker3.NewPacker(boxpacker3.WithAlgorithm(strategy)), boxes, items)
		require.Len(t, result.Boxes, 1)
		require.Equal(t, "first", result.Boxes[0].ID())
	}
}

func TestParallel_ReportsErrorsWhenNothingSucceeds(t *testing.T) {
	t.Parallel()

	strategy := boxpacker3.NewPortfolio(nil, &failingAlgo{name: "a"}, &failingAlgo{name: "b"})

	result, err := strategy.Pack(t.Context(), boxpacker3.NewProblem(nil, nil))
	require.Nil(t, result)
	require.ErrorIs(t, err, errAlgoFailed)
}

func TestParallel_KeepsResultWhenSomeAlgorithmsFail(t *testing.T) {
	t.Parallel()

	strategy := boxpacker3.NewPortfolio(nil, &failingAlgo{name: "broken"},
		boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit))

	result := packed(t, boxpacker3.NewPacker(boxpacker3.WithAlgorithm(strategy)),
		[]*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 1000)},
		[]*boxpacker3.Item{boxpacker3.NewItem("item", 10, 10, 10, 1)})

	require.Empty(t, result.Unpacked)
}

func TestStrategy_KeepsCallerItemPointers(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 10, 10, 10, 1000)}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("tall", 4, 4, 10, 10),
		boxpacker3.NewItem("wide", 10, 4, 4, 10),
		boxpacker3.NewItem("cube-1", 4, 4, 4, 10),
		boxpacker3.NewItem("cube-2", 4, 4, 4, 10),
		boxpacker3.NewItem("cube-3", 4, 4, 4, 10),
	}

	known := make(map[*boxpacker3.Item]bool, len(items))
	for _, item := range items {
		known[item] = true
	}

	result := packed(t, rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit), boxes, items)

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			require.True(t, known[item.Item], "packed item %s is not one the caller supplied", item.ID())
		}
	}

	for _, item := range result.Unpacked {
		require.True(t, known[item.Item], "unfit item %s is not one the caller supplied", item.ID())
	}
}

func TestPacker_AFailedRunHandsBackNoResult(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(&failingAlgo{name: "broken"}))

	items := []*boxpacker3.Item{boxpacker3.NewItem("item", 1, 1, 1, 1)}

	result, err := packer.Pack(t.Context(), []*boxpacker3.Box{boxpacker3.NewBox("box", 10, 10, 10, 100)}, items)
	require.Error(t, err)
	require.Nil(t, result)
}

func TestPacker_RepeatedRunsAreIdentical(t *testing.T) {
	t.Parallel()

	build := func() ([]*boxpacker3.Box, []*boxpacker3.Item) {
		boxes := []*boxpacker3.Box{
			boxpacker3.NewBox("small", 100, 100, 100, 10000),
			boxpacker3.NewBox("large", 200, 200, 200, 10000),
		}

		items := []*boxpacker3.Item{
			boxpacker3.NewItem("a", 60, 40, 30, 100),
			boxpacker3.NewItem("b", 40, 40, 40, 100),
			boxpacker3.NewItem("c", 90, 20, 20, 100),
			boxpacker3.NewItem("d", 30, 30, 30, 100),
			boxpacker3.NewItem("e", 70, 50, 25, 100),
			boxpacker3.NewItem("f", 40, 40, 40, 100),
		}

		return boxes, items
	}

	snapshot := func(strategy boxpacker3.RuleSettings) []recordedPlacement {
		boxes, items := build()

		result, err := rulePacker(strategy.Order, strategy.Selection).
			Pack(t.Context(), boxes, items)
		require.NoError(t, err)

		var placements []recordedPlacement

		for _, box := range result.Boxes {
			for _, item := range box.Items {
				position := item.Position
				dimension := item.Dimension
				placements = append(placements, recordedPlacement{
					box:  box.ID(),
					item: item.ID(),
					x:    position[0], y: position[1], z: position[2],
					w: dimension[0], h: dimension[1], d: dimension[2],
				})
			}
		}

		return placements
	}

	for _, strategy := range allStrategies() {
		t.Run(strategyName(strategy), func(t *testing.T) {
			t.Parallel()

			want := snapshot(strategy)
			require.NotEmpty(t, want)

			for range 20 {
				require.Equal(t, want, snapshot(strategy))
			}
		})
	}
}

type recordedPlacement struct {
	box  string
	item string
	x    float64
	y    float64
	z    float64
	w    float64
	h    float64
	d    float64
}
