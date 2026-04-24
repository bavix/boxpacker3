//go:build measure

package boxpacker3_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestMeasure_Search(t *testing.T) {
	t.Parallel()

	classes := []string{fileBR1, fileBR2, fileBR3, fileBR4, fileBR5, fileBR6, fileBR7}
	budgets := []int{128, 512}
	supports := []float64{0, 1}

	for _, support := range supports {
		control := measureSearch(t, nil, support, classes)

		t.Logf("=== support %.0f", support)
		reportSearch(t, "rule", control, control, classes)

		for _, nodes := range budgets {
			search, err := boxpacker3.NewSearch(nodes, searchMeasureBranching)
			require.NoError(t, err)

			reportSearch(t, search.Name(), measureSearch(t, search, support, classes), control, classes)
		}
	}
}

const (
	searchMeasureInstances = 30

	searchMeasureBranching = 3
)

type searchCell struct {
	fill    float64
	elapsed time.Duration
}

func measureSearch(t *testing.T, search *boxpacker3.Search, support float64, classes []string) map[string]searchCell {
	t.Helper()

	var algorithm boxpacker3.Algorithm = boxpacker3.NewGreedy(
		boxpacker3.OrderDecreasing, boxpacker3.SelectFullestBox)
	if search != nil {
		algorithm = search
	}

	packer := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(algorithm),
		boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: support}),
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

func reportSearch(t *testing.T, name string, cells, control map[string]searchCell, classes []string) {
	t.Helper()

	mean, controlMean := searchMean(cells, classes), searchMean(control, classes)
	elapsed := time.Duration(0)
	worst, worstClass := 0.0, ""

	for _, class := range classes {
		elapsed += cells[class].elapsed

		if drop := control[class].fill - cells[class].fill; drop > worst {
			worst, worstClass = drop, class
		}
	}

	verdict := fmt.Sprintf("%+.2f pp against the rule", (mean-controlMean)*100)

	if worstClass == "" {
		worstClass = "-"
	}

	t.Logf("%-16s mean %.4f  worst class %-8s -%.2f pp  %8s  %s",
		name, mean, worstClass, worst*100, elapsed.Round(time.Millisecond), verdict)

	for _, class := range classes {
		t.Logf("    %-8s %.4f", class, cells[class].fill)
	}
}

func searchMean(cells map[string]searchCell, classes []string) float64 {
	total := 0.0
	for _, class := range classes {
		total += cells[class].fill
	}

	return total / float64(len(classes))
}
