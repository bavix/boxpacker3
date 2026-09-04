//go:build measure

package boxpacker3_test

import (
	"cmp"
	"slices"
	"testing"
	"time"

	"github.com/bavix/boxpacker3/v2"
)

func orderings() []struct {
	name string
	less func(a, b *boxpacker3.Item) int
} {
	side := func(item *boxpacker3.Item) float64 {
		return max(item.Width(), item.Height(), item.Depth())
	}
	base := func(item *boxpacker3.Item) float64 {
		return item.Volume() / max(item.Width(), item.Height(), item.Depth())
	}

	return []struct {
		name string
		less func(a, b *boxpacker3.Item) int
	}{
		{"volume", func(a, b *boxpacker3.Item) int { return cmp.Compare(b.Volume(), a.Volume()) }},
		{"longest side", func(a, b *boxpacker3.Item) int { return cmp.Compare(side(b), side(a)) }},
		{"base area", func(a, b *boxpacker3.Item) int { return cmp.Compare(base(b), base(a)) }},
		{"height", func(a, b *boxpacker3.Item) int { return cmp.Compare(b.Depth(), a.Depth()) }},
		{"weight", func(a, b *boxpacker3.Item) int { return cmp.Compare(b.Weight(), a.Weight()) }},
	}
}

//nolint:paralleltest // the harness times the work, so the runs must not overlap.
func TestMeasure_MultiStart(t *testing.T) {
	problems := loadPublishedInstances(t, "br1.txt")
	if len(problems) > 30 {
		problems = problems[:30]
	}

	all := orderings()

	fillOf := func(problem publishedProblem, order int) float64 {
		box := problem.Box
		items := slices.Clone(problem.Items)
		slices.SortStableFunc(items, all[order].less)

		result, err := boxpacker3.NewPacker(
			boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderAsGiven, boxpacker3.SelectFullestBox))).
			Pack(t.Context(), []*boxpacker3.Box{box}, items)
		if err != nil {
			t.Fatal(err)
		}

		used := 0.0

		for _, packed := range result.Boxes {
			for _, item := range packed.Items {
				used += item.Volume()
			}
		}

		return used / box.Volume()
	}

	best := make([]float64, len(problems))

	for order := range all {
		total := 0.0
		started := time.Now()

		for i, problem := range problems {
			fill := fillOf(problem, order)
			total += fill
			best[i] = max(best[i], fill)
		}

		running := 0.0
		for _, fill := range best {
			running += fill
		}

		t.Logf("%-14s alone %.4f, best of the first %d %.4f, %s",
			all[order].name, total/float64(len(problems)),
			order+1, running/float64(len(problems)), time.Since(started).Round(time.Second))
	}
}
