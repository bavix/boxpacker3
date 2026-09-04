package boxpacker3

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func makeBoxWithItems(id string, bw, bh, bd, mw float64, items ...*piece) *Container {
	b := openBox(id, bw, bh, bd, mw)
	for _, it := range items {
		b.insert(it)
	}

	return b
}

func makeItem(id string, w, h, d, weight float64) *piece {
	return testPiece(id, w, h, d, weight)
}

func TestMinimizeBoxesGoal(t *testing.T) {
	t.Parallel()
	t.Run("nil currentBest returns true", func(t *testing.T) {
		t.Parallel()

		cand := resultOf(nil)
		require.Negative(t, FewestBoxes.Compare(cand, nil), "expected true when currentBest is nil")
	})

	t.Run("prefers fewer unfit items", func(t *testing.T) {
		t.Parallel()

		cand := resultOf(nil, makeItem("a", 1, 1, 1, 1))
		best := resultOf(nil, makeItem("a", 1, 1, 1, 1), makeItem("b", 1, 1, 1, 1))
		require.Negative(t, FewestBoxes.Compare(cand, best), "expected candidate with fewer unfit items to win")
	})

	t.Run("tie on unfit prefers fewer boxes then smaller total volume", func(t *testing.T) {
		t.Parallel()

		itemA := makeItem("i1", 1, 1, 1, 1)

		candBox := makeBoxWithItems("b1", 2, 2, 2, 10, itemA)
		bestBox1 := makeBoxWithItems("b2", 3, 3, 3, 10, itemA)
		bestBox2 := openBox("b3", 3, 3, 3, 10)

		cand := resultOf([]*Container{candBox})
		best := resultOf([]*Container{bestBox1, bestBox2})

		require.Negative(t, FewestBoxes.Compare(cand, best),
			"the smaller total volume wins when unfit and box count tie")
	})
}

func TestMaximizeItemsGoal(t *testing.T) {
	t.Parallel()
	t.Run("nil currentBest returns true", func(t *testing.T) {
		t.Parallel()
		require.Negative(t, MostItems.Compare(resultOf(nil), nil), "expected true when currentBest is nil")
	})

	t.Run("prefers fewer unfit items", func(t *testing.T) {
		t.Parallel()

		cand := resultOf(nil)
		best := resultOf(nil, makeItem("x", 1, 1, 1, 1))
		require.Negative(t, MostItems.Compare(cand, best), "expected candidate with fewer unfit items to win")
	})
}

func TestTightestPackingGoal(t *testing.T) {
	t.Parallel()
	t.Run("prefers smaller total used volume when unfit equal", func(t *testing.T) {
		t.Parallel()

		itemA := makeItem("a", 2, 2, 2, 1)

		candBox := makeBoxWithItems("b1", 2, 2, 2, 10, itemA)
		bestBox := makeBoxWithItems("b2", 4, 4, 4, 10, itemA)
		bestBoxEmpty := openBox("b3", 4, 4, 4, 10)

		cand := resultOf([]*Container{candBox})
		best := resultOf([]*Container{bestBox, bestBoxEmpty})

		require.Negative(t, LeastVolume.Compare(cand, best), "expected candidate with smaller total used volume to win")
	})
}

func TestMaxAverageFillRateGoal(t *testing.T) {
	t.Parallel()
	t.Run("prefers higher average fill rate when unfit equal", func(t *testing.T) {
		t.Parallel()

		candBox := openBox("c1", 10, 1, 1, 100)
		bestBox := openBox("b1", 10, 1, 1, 100)

		candBox.insert(makeItem("ci", 2, 2, 2, 1))
		bestBox.insert(makeItem("bi", 1, 1, 5, 1))

		cand := resultOf([]*Container{candBox})
		best := resultOf([]*Container{bestBox})

		require.Negative(t, HighestFill.Compare(cand, best), "expected candidate with higher average fill rate to win")
	})
}

func TestBalancedPackingGoal(t *testing.T) {
	t.Parallel()
	t.Run("prefers lower weight std dev when unfit equal", func(t *testing.T) {
		t.Parallel()

		c1 := openBox("c1", 2, 2, 2, 100)
		c2 := openBox("c2", 2, 2, 2, 100)

		c1.insert(makeItem("ci1", 1, 1, 1, 5))
		c2.insert(makeItem("ci2", 1, 1, 1, 5))

		b1 := openBox("b1", 2, 2, 2, 100)
		b2 := openBox("b2", 2, 2, 2, 100)

		b1.insert(makeItem("bi1", 1, 1, 1, 1))
		b2.insert(makeItem("bi2", 1, 1, 1, 9))

		cand := resultOf([]*Container{c1, c2})
		best := resultOf([]*Container{b1, b2})

		require.Less(t, weightSpread(cand.Boxes), weightSpread(best.Boxes), "expected candidate to have lower weight std dev than best")
		require.Negative(t, BalancedWeight.Compare(cand, best), "expected candidate with lower weight std dev to win")
	})
}

func TestMakeBoxWithNonDefaultMaxWeight(t *testing.T) {
	t.Parallel()

	b := makeBoxWithItems("mx", 3, 3, 3, 50)
	require.InEpsilon(t, 50.0, b.MaxWeight(), 1e-9)
}

func TestMakeGoalEpsilonTolerance(t *testing.T) {
	t.Parallel()

	eps := 1e-6
	boxA := openBox("a", 2, 2, 2, 100)
	boxB := openBox("b", 2, 2, 2, 100)

	boxA.insert(makeItem("i", 1, 1, 1, 1))
	boxB.insert(makeItem("i", 1, 1, 1, 1))

	cand := resultOf([]*Container{boxA})
	best := resultOf([]*Container{boxB})

	if math.Abs(unpackedCount(cand)-unpackedCount(best)) < eps {
		require.GreaterOrEqual(t, FewestBoxes.Compare(cand, best), 0, "expected no preference when metrics are equal within epsilon")
	}
}

func TestMakeGoalToleranceIsRelative(t *testing.T) {
	t.Parallel()

	smaller := openBox("smaller", 0.1, 0.1, 0.1, 100)
	larger := openBox("larger", 0.1, 0.1, 0.1008, 100)

	smaller.insert(makeItem("i", 0.05, 0.05, 0.05, 1))
	larger.insert(makeItem("i", 0.05, 0.05, 0.05, 1))

	tight := resultOf([]*Container{smaller})
	loose := resultOf([]*Container{larger})

	require.Negative(t, LeastVolume.Compare(tight, loose))
	require.GreaterOrEqual(t, LeastVolume.Compare(loose, tight), 0)
}
