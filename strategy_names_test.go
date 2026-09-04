package boxpacker3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestStrategyNames(t *testing.T) {
	t.Parallel()

	named := map[string]boxpacker3.Algorithm{
		nameFirstFitDecreasing:          boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit),
		nameFirstFitIncreasing:          boxpacker3.NewGreedy(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit),
		nameBestFitIncreasing:           boxpacker3.NewGreedy(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit),
		nameBestFitDecreasing:           boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectBestFit),
		nameNextFitIncreasing:           boxpacker3.NewGreedy(boxpacker3.OrderIncreasing, boxpacker3.SelectNextFit),
		nameWorstFitIncreasing:          boxpacker3.NewGreedy(boxpacker3.OrderIncreasing, boxpacker3.SelectWorstFit),
		nameAlmostWorstFitIncreasing:    boxpacker3.NewGreedy(boxpacker3.OrderIncreasing, boxpacker3.SelectAlmostWorstFit),
		"BestFor(FewestBoxes: nothing)": boxpacker3.NewPortfolio(nil),
	}

	for want, algo := range named {
		require.Equal(t, want, algo.Name())
	}
}

func TestSettingsAreDiscoverable(t *testing.T) {
	t.Parallel()

	require.Len(t, boxpacker3.ItemOrders(), 5)
	require.Len(t, boxpacker3.BoxSelections(), 6)
	require.Len(t, boxpacker3.Rotations(), 3)
	require.Len(t, boxpacker3.GoalNames(), 5)

	labels := map[string]bool{}

	for _, order := range boxpacker3.ItemOrders() {
		require.NotEmpty(t, order.Name())
		labels[order.Name()] = true
	}

	require.Len(t, labels, 5, "every order has its own label")

	labels = map[string]bool{}

	for _, selection := range boxpacker3.BoxSelections() {
		require.NotEmpty(t, selection.String())
		labels[selection.String()] = true
	}

	require.Len(t, labels, 6, "every selection rule has its own label")

	labels = map[string]bool{}

	for _, rotation := range boxpacker3.Rotations() {
		require.NotEmpty(t, rotation.String())
		labels[rotation.String()] = true
	}

	require.Len(t, labels, 3, "every rotation mode has its own label")
}

func TestGoalByName(t *testing.T) {
	t.Parallel()

	for _, name := range boxpacker3.GoalNames() {
		goal, ok := boxpacker3.GoalByName(name)
		require.True(t, ok, name)
		require.NotNil(t, goal)
	}

	goal, ok := boxpacker3.GoalByName("nonsense")
	require.False(t, ok)
	require.Nil(t, goal)
}

func TestParallelName_CarriesTheAlgorithmsAndTheGoal(t *testing.T) {
	t.Parallel()

	runner := boxpacker3.NewPortfolio(boxpacker3.BalancedWeight,
		boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit),
		boxpacker3.NewGreedy(boxpacker3.OrderIncreasing, boxpacker3.SelectNextFit),
	)

	require.Equal(t, "BestFor(BalancedWeight: FirstFitDecreasing, NextFitIncreasing)", runner.Name())
	require.Len(t, runner.Algorithms(), 2)
	require.Equal(t, "BalancedWeight", runner.Goal().Name())

	goal, known := boxpacker3.GoalByName("BalancedWeight")
	require.True(t, known)
	require.Equal(t, boxpacker3.BalancedWeight, goal)
}

func TestParallelName_ACustomGoalNamesItself(t *testing.T) {
	t.Parallel()

	mine := boxpacker3.Lexicographic("MyOwnGoal", boxpacker3.Criterion{
		Name:    "unpacked",
		Measure: func(result *boxpacker3.Result) float64 { return float64(len(result.Unpacked)) },
		Higher:  false,
	})

	runner := boxpacker3.NewPortfolio(mine, boxpacker3.NewGreedy(boxpacker3.OrderAsGiven, boxpacker3.SelectBestFit))

	require.Equal(t, "BestFor(MyOwnGoal: BestFit)", runner.Name())

	_, known := boxpacker3.GoalByName("MyOwnGoal")
	require.False(t, known, "a goal the caller wrote is not one of the ready made ones")
}

func TestParallelGoal_NilIsIgnored(t *testing.T) {
	t.Parallel()

	runner := boxpacker3.NewPortfolio(nil)
	require.Equal(t, "BestFor(FewestBoxes: nothing)", runner.Name())

	boxes, items := randomOrder(7)

	result, err := runner.Pack(t.Context(), boxpacker3.NewProblem(boxes, items))
	require.NoError(t, err)
	require.NotNil(t, result)
}
