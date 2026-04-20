package boxpacker3

import "testing"

func benchmarkGoal(b *testing.B, goal Goal) {
	b.Helper()

	item := makeItem("i", 1, 1, 1, 1)
	candBox := makeBoxWithItems("cb", 2, 2, 2, 100, item)
	bestBox := makeBoxWithItems("bb", 4, 4, 4, 100, item)

	cand := resultOf([]*Container{candBox})
	best := resultOf([]*Container{bestBox})

	b.ResetTimer()

	for range b.N {
		_ = goal.Compare(cand, best)
	}
}

func BenchmarkGoal_FewestBoxes(b *testing.B)    { benchmarkGoal(b, FewestBoxes) }
func BenchmarkGoal_MostItems(b *testing.B)      { benchmarkGoal(b, MostItems) }
func BenchmarkGoal_LeastVolume(b *testing.B)    { benchmarkGoal(b, LeastVolume) }
func BenchmarkGoal_HighestFill(b *testing.B)    { benchmarkGoal(b, HighestFill) }
func BenchmarkGoal_BalancedWeight(b *testing.B) { benchmarkGoal(b, BalancedWeight) }

func BenchmarkGoal_Lexicographic(b *testing.B) {
	goal := Lexicographic("bench",
		Criterion{Name: "unpacked", Measure: unpackedCount, Higher: false},
		Criterion{Name: "boxes", Measure: boxCount, Higher: false},
		Criterion{Name: "box volume", Measure: boxVolume, Higher: false},
	)

	benchmarkGoal(b, goal)
}
