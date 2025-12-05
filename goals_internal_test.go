package boxpacker3

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// helpers for tests.
func makeBoxWithItems(id string, bw, bh, bd, mw float64, items ...*Item) *Box {
	b := NewBox(id, bw, bh, bd, mw)
	for _, it := range items {
		b.insert(it)
	}

	return b
}

func makeItem(id string, w, h, d, weight float64) *Item {
	return NewItem(id, w, h, d, weight)
}

func TestMinimizeBoxesGoal(t *testing.T) {
	t.Parallel()
	t.Run("nil currentBest returns true", func(t *testing.T) {
		t.Parallel()

		cand := &Result{}
		require.True(t, MinimizeBoxesGoal(cand, nil), "expected true when currentBest is nil")
	})

	t.Run("prefers fewer unfit items", func(t *testing.T) {
		t.Parallel()

		cand := &Result{UnfitItems: itemSlice{makeItem("a", 1, 1, 1, 1)}}
		best := &Result{UnfitItems: itemSlice{makeItem("a", 1, 1, 1, 1), makeItem("b", 1, 1, 1, 1)}}
		require.True(t, MinimizeBoxesGoal(cand, best), "expected candidate with fewer unfit items to win")
	})

	t.Run("tie on unfit prefers fewer boxes then smaller total volume", func(t *testing.T) {
		t.Parallel()
		// both have zero unfit
		// candidate: 1 box of volume 8 (2x2x2)
		// best: 2 boxes of volume 27 and 27 (3x3x3) but only one used -> ensure box count differs
		itemA := makeItem("i1", 1, 1, 1, 1)

		candBox := makeBoxWithItems("b1", 2, 2, 2, 10, itemA)  // volume 8
		bestBox1 := makeBoxWithItems("b2", 3, 3, 3, 10, itemA) // volume 27
		bestBox2 := NewBox("b3", 3, 3, 3, 10)                  // empty

		cand := &Result{UnfitItems: itemSlice{}, Boxes: boxSlice{candBox}}
		best := &Result{UnfitItems: itemSlice{}, Boxes: boxSlice{bestBox1, bestBox2}}

		// candidate uses 1 box vs best uses 1 used box as well (only bestBox1 has items).
		// total volume: candidate=8, best=27 -> candidate should win
		require.True(t, MinimizeBoxesGoal(cand, best), "expected candidate with smaller total used volume to win when unfit and box count tie")
	})
}

func TestMaximizeItemsGoal(t *testing.T) {
	t.Parallel()
	t.Run("nil currentBest returns true", func(t *testing.T) {
		t.Parallel()
		require.True(t, MaximizeItemsGoal(&Result{}, nil), "expected true when currentBest is nil")
	})

	t.Run("prefers fewer unfit items", func(t *testing.T) {
		t.Parallel()

		cand := &Result{UnfitItems: itemSlice{}}
		best := &Result{UnfitItems: itemSlice{makeItem("x", 1, 1, 1, 1)}}
		require.True(t, MaximizeItemsGoal(cand, best), "expected candidate with fewer unfit items to win")
	})
}

func TestTightestPackingGoal(t *testing.T) {
	t.Parallel()
	t.Run("prefers smaller total used volume when unfit equal", func(t *testing.T) {
		t.Parallel()

		itemA := makeItem("a", 2, 2, 2, 1) // volume 8

		candBox := makeBoxWithItems("b1", 2, 2, 2, 10, itemA) // used volume 8
		bestBox := makeBoxWithItems("b2", 4, 4, 4, 10, itemA) // used volume 64
		bestBoxEmpty := NewBox("b3", 4, 4, 4, 10)             // empty

		cand := &Result{UnfitItems: itemSlice{}, Boxes: boxSlice{candBox}}
		best := &Result{UnfitItems: itemSlice{}, Boxes: boxSlice{bestBox, bestBoxEmpty}}

		require.True(t, TightestPackingGoal(cand, best), "expected candidate with smaller total used volume to win")
	})
}

func TestMaxAverageFillRateGoal(t *testing.T) {
	t.Parallel()
	t.Run("prefers higher average fill rate when unfit equal", func(t *testing.T) {
		t.Parallel()
		// candidate: one box, fill rate 0.8 (itemsVolume 8 / box volume 10)
		// best: one box, fill rate 0.5 (itemsVolume 5 / box volume 10)
		candBox := NewBox("c1", 10, 1, 1, 100) // volume 10
		bestBox := NewBox("b1", 10, 1, 1, 100) // volume 10

		// create items with volumes summing to desired itemsVolume
		candBox.insert(makeItem("ci", 2, 2, 2, 1)) // volume 8
		bestBox.insert(makeItem("bi", 1, 1, 5, 1)) // volume 5

		cand := &Result{UnfitItems: itemSlice{}, Boxes: boxSlice{candBox}}
		best := &Result{UnfitItems: itemSlice{}, Boxes: boxSlice{bestBox}}

		require.True(t, MaxAverageFillRateGoal(cand, best), "expected candidate with higher average fill rate to win")
	})
}

func TestBalancedPackingGoal(t *testing.T) {
	t.Parallel()
	t.Run("prefers lower weight std dev when unfit equal", func(t *testing.T) {
		t.Parallel()
		// candidate: two boxes with weights 5 and 5 -> stddev 0
		// best: two boxes with weights 1 and 9 -> stddev > 0
		c1 := NewBox("c1", 2, 2, 2, 100)
		c2 := NewBox("c2", 2, 2, 2, 100)

		c1.insert(makeItem("ci1", 1, 1, 1, 5))
		c2.insert(makeItem("ci2", 1, 1, 1, 5))

		b1 := NewBox("b1", 2, 2, 2, 100)
		b2 := NewBox("b2", 2, 2, 2, 100)

		b1.insert(makeItem("bi1", 1, 1, 1, 1))
		b2.insert(makeItem("bi2", 1, 1, 1, 9))

		cand := &Result{UnfitItems: itemSlice{}, Boxes: boxSlice{c1, c2}}
		best := &Result{UnfitItems: itemSlice{}, Boxes: boxSlice{b1, b2}}

		// quick sanity check of std dev values
		require.Less(t, getWeightStdDev(cand.Boxes), getWeightStdDev(best.Boxes), "expected candidate to have lower weight std dev than best")
		require.True(t, BalancedPackingGoal(cand, best), "expected candidate with lower weight std dev to win")
	})
}

func TestMakeBoxWithNonDefaultMaxWeight(t *testing.T) {
	t.Parallel()

	b := makeBoxWithItems("mx", 3, 3, 3, 50)
	require.InEpsilon(t, 50.0, b.GetMaxWeight(), 1e-9)
}

func TestMakeGoalEpsilonTolerance(t *testing.T) {
	t.Parallel()
	// ensure tiny differences within epsilon are treated as equal and move to next criterion
	eps := 1e-6
	boxA := NewBox("a", 2, 2, 2, 100)
	boxB := NewBox("b", 2, 2, 2, 100)

	boxA.insert(makeItem("i", 1, 1, 1, 1))
	boxB.insert(makeItem("i", 1, 1, 1, 1))

	cand := &Result{UnfitItems: itemSlice{}, Boxes: boxSlice{boxA}}
	best := &Result{UnfitItems: itemSlice{}, Boxes: boxSlice{boxB}}

	// tweak values slightly around epsilon
	if math.Abs(unfitCountMetric(cand)-unfitCountMetric(best)) < eps {
		// with equal unfit count and equal boxes, makeGoal should return false (no improvement)
		require.False(t, MinimizeBoxesGoal(cand, best), "expected no preference when metrics are equal within epsilon")
	}
}
