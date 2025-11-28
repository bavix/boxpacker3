package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

// TestParallel_PickBestResult verifies that the parallel runner actually picks the best outcome
// from the available algorithms using real packing logic.
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

	parallelAlgo := boxpacker3.NewParallelStrategy(
		boxpacker3.WithAlgorithms(
			boxpacker3.NewMinimizeBoxesStrategy(),
			boxpacker3.NewBestFitStrategy(),
		),
		boxpacker3.WithGoal(boxpacker3.TightestPackingGoal),
	)

	packer := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(parallelAlgo))
	result, err := packer.PackCtx(context.Background(), boxes, items)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Empty(t, result.UnfitItems, "Parallel strategy should have found a solution where all items fit")

	require.NotEmpty(t, result.Boxes)
}

// TestParallel_GoalSwitching verifies that changing the Goal (ComparatorFunc)
// changes which result is selected as the winner.
func TestParallel_GoalSwitching(t *testing.T) {
	t.Parallel()

	// Setup Mocks:
	// Algo A: Uses 1 Box, but fails to pack 1 item (1 Unfit).
	algoA := &MockAlgo{
		name: "AlgoA",
		res: &boxpacker3.Result{
			Boxes:      makeMockBoxes(1),
			UnfitItems: makeMockItems(1),
		},
	}

	// Algo B: Uses 2 Boxes, but packs ALL items (0 Unfit).
	algoB := &MockAlgo{
		name: "AlgoB",
		res: &boxpacker3.Result{
			Boxes:      makeMockBoxes(2),
			UnfitItems: makeMockItems(0),
		},
	}

	// Case 1: Standard Goal "MinimizeBoxes".
	strat1 := boxpacker3.NewParallelStrategy(
		boxpacker3.WithAlgorithms(algoA, algoB),
		boxpacker3.WithGoal(boxpacker3.MinimizeBoxesGoal),
	)

	res1, _ := strat1.Pack(context.Background(), nil, nil)
	require.Empty(t, res1.UnfitItems, "Standard goal should prefer Algo B (0 unfit items)")
	require.Len(t, res1.Boxes, 2, "Standard goal should accept more boxes to fit all items")

	// Case 2: Custom Goal "Pure Box Count".
	pureBoxCountGoal := func(cand, best *boxpacker3.Result) bool {
		if best == nil {
			return true
		}

		return len(cand.Boxes) < len(best.Boxes)
	}

	strat2 := boxpacker3.NewParallelStrategy(
		boxpacker3.WithAlgorithms(algoA, algoB),
		boxpacker3.WithGoal(pureBoxCountGoal),
	)

	res2, _ := strat2.Pack(context.Background(), nil, nil)
	require.Len(t, res2.Boxes, 1, "Custom goal should prefer Algo A (fewer boxes)")
	require.Len(t, res2.UnfitItems, 1, "Custom goal accepted the result with unfit items")
}

type MockAlgo struct {
	name string
	res  *boxpacker3.Result
}

func (m *MockAlgo) Name() string { return m.name }

// Pack implements PackingAlgorithm.
func (m *MockAlgo) Pack(_ context.Context, _ []*boxpacker3.Box, _ []*boxpacker3.Item) (*boxpacker3.Result, error) {
	return m.res, nil
}

func makeMockBoxes(n int) []*boxpacker3.Box {
	boxes := make([]*boxpacker3.Box, n)
	for i := range n {
		b := boxpacker3.NewBox("mock-box", 10, 10, 10, 100)
		b.PutItem(boxpacker3.NewItem("mock-item", 1, 1, 1, 1), boxpacker3.Pivot{})
		boxes[i] = b
	}

	return boxes
}

func makeMockItems(n int) []*boxpacker3.Item {
	items := make([]*boxpacker3.Item, n)
	for i := range n {
		items[i] = boxpacker3.NewItem("unfit-item", 1, 1, 1, 1)
	}

	return items
}
