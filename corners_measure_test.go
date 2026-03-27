//go:build measure

package boxpacker3_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

// Criterion, written down before the run. The free space corners were measured
// against the ordinary rule and gained 2.59 points with no support for three
// times the time, which is why they are an option and not the default. What is
// asked here is whether they are worth having under the search as well, where
// the corners a block is laid against already come from the same spaces:
//
//   - the combination is recommended in the README only if it beats the search
//     alone by at least 1.0 points at both support settings and costs no more
//     than twice the time;
//   - a gain at one support setting and a loss at the other is not a
//     recommendation, it is a note.
//
// Run with: go test -tags measure -run TestMeasure_SpaceCorners -timeout 180m.
func TestMeasure_SpaceCorners(t *testing.T) {
	t.Parallel()

	classes := []string{fileBR1, fileBR2, fileBR3, fileBR4, fileBR5, fileBR6, fileBR7}

	for _, support := range []float64{0, 1} {
		t.Logf("=== support %.0f", support)

		control := measureCorners(t, false, false, support, classes)
		reportCorners(t, "rule", control, control, classes)
		reportCorners(t, "rule+corners", measureCorners(t, false, true, support, classes), control, classes)

		search := measureCorners(t, true, false, support, classes)
		reportCorners(t, "search", search, control, classes)
		reportCorners(t, "search+corners", measureCorners(t, true, true, support, classes), search, classes)
	}
}

// cornersBudget is the node budget the search runs under here: the smaller of
// the two the table in TestMeasure_Search reads, so the two runs can be laid
// beside each other.
const cornersBudget = 128

func measureCorners(t *testing.T, searching, corners bool, support float64, classes []string) map[string]searchCell {
	t.Helper()

	var algorithm boxpacker3.Algorithm = boxpacker3.NewGreedy(
		boxpacker3.OrderDecreasing, boxpacker3.SelectFullestBox)

	if searching {
		search, err := boxpacker3.NewSearch(cornersBudget, searchMeasureBranching)
		require.NoError(t, err)

		algorithm = search
	}

	packer := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(algorithm),
		boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: support, FreeSpaceCorners: corners}),
	)

	readings := map[string]searchCell{}

	for _, class := range classes {
		problems := loadPublishedInstances(t, class)
		if len(problems) > searchMeasureInstances {
			problems = problems[:searchMeasureInstances]
		}

		started := time.Now()
		total := 0.0

		for _, problem := range problems {
			result, err := packer.Pack(t.Context(), []*boxpacker3.Box{problem.Box}, problem.Items)
			require.NoError(t, err)
			require.NotEmpty(t, result.Boxes)

			total += volumeUtilisation(result.Boxes[0])
		}

		readings[class] = searchCell{fill: total / float64(len(problems)), elapsed: time.Since(started)}
	}

	return readings
}

func reportCorners(t *testing.T, name string, cells, against map[string]searchCell, classes []string) {
	t.Helper()

	mean, base := searchMean(cells, classes), searchMean(against, classes)
	elapsed := time.Duration(0)
	worst, worstClass := 0.0, "-"

	for _, class := range classes {
		elapsed += cells[class].elapsed

		if drop := against[class].fill - cells[class].fill; drop > worst {
			worst, worstClass = drop, class
		}
	}

	t.Logf("%-16s mean %.4f  worst class %-8s -%.2f pp  %8s  %s",
		name, mean, worstClass, worst*100, elapsed.Round(time.Millisecond),
		fmt.Sprintf("%+.2f pp", (mean-base)*100))

	for _, class := range classes {
		t.Logf("    %-8s %.4f", class, cells[class].fill)
	}
}
