package boxpacker3_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

func loadSpread(result *boxpacker3.Result) (float64, int) {
	weights := make([]float64, 0, len(result.Boxes))

	for _, box := range result.Boxes {
		if len(box.Items) == 0 {
			continue
		}

		total := 0.0

		for _, item := range box.Items {
			total += item.Weight()
		}

		weights = append(weights, total)
	}

	if len(weights) < 2 {
		return 0, len(weights)
	}

	mean := 0.0

	for _, w := range weights {
		mean += w
	}

	mean /= float64(len(weights))

	variance := 0.0

	for _, w := range weights {
		variance += (w - mean) * (w - mean)
	}

	return math.Sqrt(variance / float64(len(weights))), len(weights)
}

func TestWorstFit_StaysInBoundAndSpreadsTheLoad(t *testing.T) {
	t.Parallel()

	worstSpread, firstSpread := 0.0, 0.0
	worstRuns, firstRuns := 0, 0
	overBound := 0

	for seed := range 300 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := identicalShelf(12, 300, 250, 200, 1e9)
		items := cargoOf(random, 10+random.Intn(18), 40, 90, 100, 900)

		worst := packBy(t, boxpacker3.SelectWorstFit, boxes, items)
		first := packBy(t, boxpacker3.SelectFirstFit, boxes, items)

		spread, worstBoxes := loadSpread(worst)
		if worstBoxes >= 2 {
			worstSpread += spread
			worstRuns++
		}

		spread, firstBoxes := loadSpread(first)
		if firstBoxes >= 2 {
			firstSpread += spread
			firstRuns++
		}

		if firstBoxes > 0 && worstBoxes > firstBoxes*2 {
			overBound++

			if overBound <= 3 {
				t.Errorf("seed %d: worst fit used %d boxes, first fit %d",
					seed, worstBoxes, firstBoxes)
			}
		}
	}

	t.Logf("runs past the bound: %d of 300", overBound)
	reportSpread(t, worstSpread, worstRuns, firstSpread, firstRuns)
}

func packBy(t *testing.T, rule boxpacker3.BoxSelection,
	boxes []*boxpacker3.Box, items []*boxpacker3.Item,
) *boxpacker3.Result {
	t.Helper()

	result, err := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, rule)),
	).Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	return result
}

func reportSpread(t *testing.T, worstSpread float64, worstRuns int, firstSpread float64, firstRuns int) {
	t.Helper()

	if worstRuns == 0 || firstRuns == 0 {
		return
	}

	worstMean := worstSpread / float64(worstRuns)
	firstMean := firstSpread / float64(firstRuns)

	t.Logf("weight spread: worst fit %.0f g, first fit %.0f g", worstMean, firstMean)

	if worstMean > firstMean {
		t.Error("worst fit spread the load less evenly than first fit, which is what it is for")
	}
}
