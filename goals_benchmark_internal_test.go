package boxpacker3

import "testing"

func benchResults() (*Result, *Result) {
	it := NewItem("i1", 2, 2, 2, 1)
	candBox := makeBoxWithItems("cb", 3, 3, 3, 100, it)
	bestBox := makeBoxWithItems("bb", 4, 4, 4, 100, it)

	cand := &Result{UnfitItems: itemSlice{}, Boxes: boxSlice{candBox}}
	best := &Result{UnfitItems: itemSlice{}, Boxes: boxSlice{bestBox}}

	return cand, best
}

func BenchmarkMinimizeBoxesGoal_Call(b *testing.B) {
	cand, best := benchResults()

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		MinimizeBoxesGoal(cand, best)
	}
}

func BenchmarkMinimizeBoxesGoal_Closure(b *testing.B) {
	// benchmark calling the comparator closure directly (avoids recreate of closure each call)
	comp := makeGoal(
		criterion{unfitCountMetric, lessIsBetter},
		criterion{boxCountMetric, lessIsBetter},
		criterion{totalVolumeMetric, lessIsBetter},
	)
	cand, best := benchResults()

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		comp(cand, best)
	}
}

func BenchmarkMaximizeItemsGoal(b *testing.B) {
	cand, best := benchResults()

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		MaximizeItemsGoal(cand, best)
	}
}

func BenchmarkTightestPackingGoal(b *testing.B) {
	cand, best := benchResults()

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		TightestPackingGoal(cand, best)
	}
}

func BenchmarkMaxAverageFillRateGoal(b *testing.B) {
	cand, best := benchResults()

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		MaxAverageFillRateGoal(cand, best)
	}
}

func BenchmarkBalancedPackingGoal(b *testing.B) {
	cand, best := benchResults()

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		BalancedPackingGoal(cand, best)
	}
}
