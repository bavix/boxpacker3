package boxpacker3_test

import (
	"math"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

type criterionCheck struct {
	metric   func(*boxpacker3.Result) float64
	moreWins bool
}

func unfitCount(result *boxpacker3.Result) float64 {
	return float64(len(result.Unpacked))
}

func usedBoxCount(result *boxpacker3.Result) float64 {
	return float64(usedBoxCountOf(result))
}

func averageFill(result *boxpacker3.Result) float64 {
	total, count := 0.0, 0

	for _, box := range result.Boxes {
		items := box.Items
		if len(items) == 0 {
			continue
		}

		used := 0.0

		for _, item := range items {
			d := item.Dimension
			used += d[0] * d[1] * d[2]
		}

		total += used / (box.Box.Width() * box.Box.Height() * box.Box.Depth())
		count++
	}

	if count == 0 {
		return 0
	}

	return total / float64(count)
}

func packedVolume(result *boxpacker3.Result) float64 {
	total := 0.0

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			d := item.Dimension
			total += d[0] * d[1] * d[2]
		}
	}

	return total
}

func beats(rival, winner *boxpacker3.Result, criteria []criterionCheck) bool {
	for _, c := range criteria {
		a, b := c.metric(rival), c.metric(winner)
		if math.Abs(a-b) < 1e-9 {
			continue
		}

		if c.moreWins {
			return a > b
		}

		return a < b
	}

	return false
}

func parallelOver(t *testing.T, goalName string,
	boxes []*boxpacker3.Box, items []*boxpacker3.Item,
) *boxpacker3.Result {
	t.Helper()

	goal, ok := boxpacker3.GoalByName(goalName)
	if !ok {
		t.Fatalf("no goal named %s", goalName)
	}

	algorithms := make([]boxpacker3.Algorithm, 0, 6)
	for _, rule := range boxpacker3.BoxSelections() {
		algorithms = append(algorithms, boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, rule))
	}

	result, err := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(
		boxpacker3.NewPortfolio(goal, algorithms...))).Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	return result
}

func TestAudit_ParallelGoalPicksTheBest(t *testing.T) {
	t.Parallel()

	goals := map[string][]criterionCheck{
		"FewestBoxes": {
			{unfitCount, false},
			{usedBoxCount, false},
		},
		"MostItems": {
			{unfitCount, false},
		},
		"HighestFill": {
			{unfitCount, false},
			{averageFill, true},
		},
		"LeastVolume": {
			{unfitCount, false},
			{packedVolume, true},
		},
	}

	losses := 0

	for seed := range 80 {
		boxes, items := randomOrder(int64(seed))

		for name, criteria := range goals {
			winner := parallelOver(t, name, boxes, items)

			for _, rule := range boxpacker3.BoxSelections() {
				rival, err := rulePacker(boxpacker3.OrderDecreasing, rule).
					Pack(t.Context(), boxes, items)
				if err != nil {
					t.Fatal(err)
				}

				if beats(rival, winner, criteria) {
					losses++

					if losses <= 5 {
						t.Errorf("seed %d goal %s: %s beat the chosen packing", seed, name, rule)
					}
				}
			}
		}
	}

	t.Logf("goals beaten by a lone rule: %d", losses)
}
