package boxpacker3_test

import (
	"math/rand"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

func TestAlmostWorstFit_IsDistinguishableFromWorstFit(t *testing.T) {
	t.Parallel()

	same, total := 0, 0
	sameCount := 0

	for seed := range 400 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 3+random.Intn(6), 120, 160, 1e9)
		items := cargoOf(random, 8+random.Intn(18), 30, 90, 100, 600)

		run := func(rule boxpacker3.BoxSelection) *boxpacker3.Result {
			result, err := boxpacker3.NewPacker(
				boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, rule)),
			).Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			return result
		}

		worst := run(boxpacker3.SelectWorstFit)
		almost := run(boxpacker3.SelectAlmostWorstFit)

		total++

		if fingerprint(worst) == fingerprint(almost) {
			same++
		}

		countOf := func(result *boxpacker3.Result) int {
			used := usedBoxCountOf(result)

			return used
		}

		if countOf(worst) == countOf(almost) {
			sameCount++
		}
	}

	t.Logf("identical packings: %d of %d", same, total)
	t.Logf("identical box counts: %d of %d", sameCount, total)

	if same*4 >= total {
		t.Errorf("the two rules produced the same packing %d times in %d, so the second name earns nothing",
			same, total)
	}
}
