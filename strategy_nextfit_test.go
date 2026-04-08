package boxpacker3_test

import (
	"math/rand"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

func TestNextFit_LeavesBehindWhatTheDefinitionSaysItMust(t *testing.T) {
	t.Parallel()

	missedOnPurpose, ordersWithMisses := 0, 0

	for seed := range 400 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 3+random.Intn(4), 110, 150, 1e9)
		items := cargoOf(random, 8+random.Intn(16), 30, 90, 1, 1)

		result, err := boxpacker3.NewPacker(
			boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectNextFit)),
		).Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		misses := missesBySkippedBoxes(boxes, result)
		missedOnPurpose += misses

		if misses > 0 {
			ordersWithMisses++
		}
	}

	t.Logf("goods left behind that a skipped box would have taken: %d, across %d of 400 orders",
		missedOnPurpose, ordersWithMisses)

	if missedOnPurpose == 0 {
		t.Error("the restriction never bit, so this fixture does not exercise Next Fit at all")
	}
}

func missesBySkippedBoxes(shelf []*boxpacker3.Box, result *boxpacker3.Result) int {
	used := map[*boxpacker3.Box]int{}
	for _, box := range result.Boxes {
		used[box.Box]++
	}

	misses := 0

	for _, item := range result.Unpacked {
		for _, box := range shelf {
			if used[box] >= box.Quantity() {
				continue
			}

			if packsAlone(box, item.Item) {
				misses++

				break
			}
		}
	}

	return misses
}
