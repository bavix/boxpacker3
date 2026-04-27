//go:build measure

package boxpacker3_test

import (
	"testing"
	"time"

	"github.com/bavix/boxpacker3/v2"
)

//nolint:paralleltest // the harness times the work, so the runs must not overlap.
func TestMeasure_Fill(t *testing.T) {
	problems := loadPublishedInstances(t, "br1.txt")
	if len(problems) > 30 {
		problems = problems[:30]
	}

	for _, setting := range []struct {
		name    string
		support float64
		search  [2]int
	}{
		{"plain", 0, [2]int{0, 0}},
		{"support", 1, [2]int{0, 0}},
		{"search 1/1", 0, [2]int{1, 1}},
		{"search 2/2", 0, [2]int{2, 2}},
		{"search 4/3", 0, [2]int{4, 3}},
	} {
		total := 0.0
		started := time.Now()

		for _, problem := range problems {
			box := problem.Box

			options := []boxpacker3.Option{
				boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit)),
				boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: setting.support}),
			}

			if setting.search[0] > 0 {
				search, err := boxpacker3.NewSearch(setting.search[0]*32, setting.search[1])
				if err != nil {
					t.Fatal(err)
				}

				options = append(options, boxpacker3.WithAlgorithm(search))
			}

			result, err := boxpacker3.NewPacker(options...).
				Pack(t.Context(), []*boxpacker3.Box{box}, problem.Items)
			if err != nil {
				t.Fatal(err)
			}

			for _, packed := range result.Boxes {
				used := 0.0

				for _, item := range packed.Items {
					dimension := item.Dimension
					used += dimension[0] * dimension[1] * dimension[2]
				}

				total += used / (packed.Box.Width() * packed.Box.Height() * packed.Box.Depth())
			}
		}

		t.Logf("%-8s %d problems, mean utilisation %.4f, %.1fs",
			setting.name, len(problems), total/float64(len(problems)),
			time.Since(started).Seconds())
	}
}
