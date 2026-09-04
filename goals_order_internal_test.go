package boxpacker3

import (
	"testing"
)

const evenLoad = 10.0

func resultWith(boxes int, boxVolume, fillShare float64, unfit int) *Result {
	packed := make([]*Container, 0, boxes)

	for i := range boxes {
		box := openBox(string(rune('a'+i)), boxVolume, 1, 1, 1e9)
		box.itemsVolume = boxVolume * fillShare
		box.itemsWeight = evenLoad
		box.items = []*piece{testPiece("x", 1, 1, 1, evenLoad)}

		packed = append(packed, box)
	}

	left := make([]*piece, unfit)

	for i := range left {
		left[i] = testPiece("left", 1, 1, 1, 0)
	}

	return resultOf(packed, left...)
}

func TestGoals_DecideInTheDocumentedOrder(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		goal    Goal
		better  *Result
		worse   *Result
		because string
	}{
		{
			"MinimizeBoxes prefers fewer left behind", FewestBoxes,
			resultWith(4, 100, 0.5, 0), resultWith(1, 100, 0.5, 3),
			"unfit count comes first",
		},
		{
			"MinimizeBoxes then prefers fewer boxes", FewestBoxes,
			resultWith(2, 100, 0.5, 0), resultWith(3, 10, 0.5, 0),
			"box count comes before volume",
		},
		{
			"MinimizeBoxes then prefers less volume", FewestBoxes,
			resultWith(2, 50, 0.5, 0), resultWith(2, 100, 0.5, 0),
			"volume breaks a tie on box count",
		},
		{
			"MaximizeItems looks only at what is left behind", MostItems,
			resultWith(9, 900, 0.1, 0), resultWith(1, 10, 0.9, 1),
			"nothing but the unfit count",
		},
		{
			"TightestPacking prefers less volume before fewer boxes", LeastVolume,
			resultWith(3, 30, 0.5, 0), resultWith(2, 100, 0.5, 0),
			"volume comes before box count",
		},
		{
			"MaxAverageFillRate prefers the fuller boxes", HighestFill,
			resultWith(4, 100, 0.9, 0), resultWith(1, 100, 0.2, 0),
			"fill rate, not box count",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if test.goal.Compare(test.better, test.worse) >= 0 {
				t.Errorf("the goal did not prefer the better result: %s", test.because)
			}

			if test.goal.Compare(test.worse, test.better) < 0 {
				t.Errorf("the goal preferred the worse result too: %s", test.because)
			}
		})
	}
}

func TestGoals_BalancedPackingRanksOnTheSpread(t *testing.T) {
	t.Parallel()

	even := resultWith(2, 100, 0.5, 0)

	uneven := resultWith(2, 100, 0.5, 0)
	uneven.Boxes[0].Stats.ItemsWeight = 1
	uneven.Boxes[1].Stats.ItemsWeight = 100

	if BalancedWeight.Compare(even, uneven) >= 0 {
		t.Error("the even load was not preferred")
	}

	if BalancedWeight.Compare(uneven, even) < 0 {
		t.Error("the uneven load was preferred as well")
	}
}
