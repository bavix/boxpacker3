package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestParallel_PickBestResult(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 1000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 1000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("medium-1", 60, 60, 60, 100),
		boxpacker3.NewItem("medium-2", 60, 60, 60, 100),
		boxpacker3.NewItem("small-1", 30, 30, 30, 100),
		boxpacker3.NewItem("small-2", 30, 30, 30, 100),
	}

	parallelAlgo := boxpacker3.NewPortfolio(boxpacker3.LeastVolume,
		boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit),
		boxpacker3.NewGreedy(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit),
	)

	packer := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(parallelAlgo))
	result, err := packer.Pack(context.Background(), boxes, items)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Empty(t, result.Unpacked, "Parallel strategy should have found a solution where all items fit")

	require.NotEmpty(t, result.Boxes)
}

func TestParallel_GoalSwitching(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("mock-box", 10, 10, 10, 100)}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("a", 1, 1, 1, 1),
		boxpacker3.NewItem("b", 1, 1, 1, 1),
		boxpacker3.NewItem("c", 1, 1, 1, 1),
	}

	algoA := &splitAlgo{name: "AlgoA", boxes: 1, leave: 1, pick: 0}

	algoB := &splitAlgo{name: "AlgoB", boxes: 2, leave: 0, pick: 0}

	strat1 := boxpacker3.NewPortfolio(boxpacker3.FewestBoxes, algoA, algoB)

	res1 := packed(t, boxpacker3.NewPacker(boxpacker3.WithAlgorithm(strat1)), boxes, items)
	require.Empty(t, res1.Unpacked, "Standard goal should prefer Algo B (0 unfit items)")
	require.Len(t, res1.Boxes, 2, "Standard goal should accept more boxes to fit all items")

	pureBoxCountGoal := boxpacker3.Lexicographic("PureBoxCount", boxpacker3.Criterion{
		Name:    "boxes",
		Measure: func(result *boxpacker3.Result) float64 { return float64(len(result.Boxes)) },
		Higher:  false,
	})

	strat2 := boxpacker3.NewPortfolio(pureBoxCountGoal, algoA, algoB)

	res2 := packed(t, boxpacker3.NewPacker(boxpacker3.WithAlgorithm(strat2)), boxes, items)
	require.Len(t, res2.Boxes, 1, "Custom goal should prefer Algo A (fewer boxes)")
	require.Len(t, res2.Unpacked, 1, "Custom goal accepted the result with unfit items")
}

type splitAlgo struct {
	name  string
	boxes int
	leave int
	pick  int
}

func (a *splitAlgo) Name() string { return a.name }

func (a *splitAlgo) Pack(_ context.Context, problem *boxpacker3.Problem) (*boxpacker3.Packing, error) {
	shelf := problem.Boxes()
	opened := make([]*boxpacker3.Container, 0, a.boxes)

	for i := range a.boxes {
		opened = append(opened, problem.Open(shelf[a.pick], i))
	}

	instances := problem.Instances()
	leftover := make([]boxpacker3.Instance, 0, a.leave)

	for i, instance := range instances {
		if i >= len(instances)-a.leave {
			leftover = append(leftover, instance)

			continue
		}

		if !opened[i%a.boxes].Fit(instance) {
			leftover = append(leftover, instance)
		}
	}

	return problem.Packing(opened, leftover), nil
}
