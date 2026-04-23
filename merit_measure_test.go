//go:build measure

package boxpacker3_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestMeasure_Merits(t *testing.T) {
	t.Parallel()

	classes := []string{fileBR1, fileBR2, fileBR3, fileBR4, fileBR5, fileBR6, fileBR7}
	supports := []float64{0, 1}

	readings := map[string]map[float64]map[string]meritCell{}

	for _, merit := range boxpacker3.Merits() {
		readings[merit.Name()] = map[float64]map[string]meritCell{}

		for _, support := range supports {
			readings[merit.Name()][support] = measureMerit(t, merit, support, classes)
		}
	}

	for _, support := range supports {
		t.Logf("=== support %.0f", support)
		reportMerits(t, readings, support, classes)
	}
}

const meritInstances = 30

type meritCell struct {
	fill    float64
	elapsed time.Duration
}

func measureMerit(t *testing.T, merit boxpacker3.Merit, support float64, classes []string) map[string]meritCell {
	t.Helper()

	packer := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectFullestBox)),
		boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: support}),
		boxpacker3.WithMerit(merit),
	)

	readings := map[string]meritCell{}

	for _, class := range classes {
		problems := loadPublishedInstances(t, class)
		if len(problems) > meritInstances {
			problems = problems[:meritInstances]
		}

		started := time.Now()

		total := 0.0

		for _, problem := range problems {
			result, err := packer.Pack(t.Context(), []*boxpacker3.Box{problem.Box}, problem.Items)
			require.NoError(t, err)
			require.NotEmpty(t, result.Boxes)

			total += volumeUtilisation(result.Boxes[0])
		}

		readings[class] = meritCell{fill: total / float64(len(problems)), elapsed: time.Since(started)}
	}

	return readings
}

func reportMerits(t *testing.T, readings map[string]map[float64]map[string]meritCell, support float64, classes []string) {
	t.Helper()

	control := boxpacker3.ContactFirst.Name()
	controlMean := meanFill(readings[control][support], classes)

	for _, merit := range boxpacker3.Merits() {
		name := merit.Name()
		cells := readings[name][support]

		mean := meanFill(cells, classes)
		elapsed := time.Duration(0)
		worst, worstClass := 0.0, ""

		for _, class := range classes {
			elapsed += cells[class].elapsed

			if drop := readings[control][support][class].fill - cells[class].fill; drop > worst {
				worst, worstClass = drop, class
			}
		}

		verdict := fmt.Sprintf("%+.2f pp against the control", (mean-controlMean)*100)
		if name == control {
			verdict = "control"
		}

		t.Logf("%-18s mean %.4f  worst class %-8s -%.2f pp  %8s  %s",
			name, mean, worstClass, worst*100, elapsed.Round(time.Millisecond), verdict)
	}
}

func meanFill(cells map[string]meritCell, classes []string) float64 {
	total := 0.0
	for _, class := range classes {
		total += cells[class].fill
	}

	return total / float64(len(classes))
}
