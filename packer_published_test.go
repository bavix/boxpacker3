package boxpacker3_test

import (
	"encoding/csv"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

const (
	baselineFile = "expected-utilisation.csv"
	historyFile  = "history.csv"
)

type baselineSettings struct {
	Instances int
	Merit     string
	Spaces    string
}

func currentSettings(instances int) baselineSettings {
	return baselineSettings{
		Instances: instances,
		Merit:     boxpacker3.ContactFirst.Name(),
		Spaces:    "extreme-points",
	}
}

func (s baselineSettings) String() string {
	return fmt.Sprintf("instances=%d merit=%s spaces=%s", s.Instances, s.Merit, s.Spaces)
}

var updateBaseline = flag.Bool("update-baseline", false,
	"rewrite testdata/expected-utilisation.csv from this run")

const baselineTolerance = 0.005

func publishedFiles() []string {
	return []string{
		fileBR1, fileBR2, fileBR3, fileBR4,
		fileBR5, fileBR6, fileBR7, "loh-nee.txt",
	}
}

func meanUtilisation(tb testing.TB, problems []publishedProblem, strategy boxpacker3.RuleSettings, support float64) float64 {
	tb.Helper()

	packer := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(strategy.Order, strategy.Selection)),
		boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: support}),
	)

	var total float64

	for _, problem := range problems {
		result, err := packer.Pack(tb.Context(),
			[]*boxpacker3.Box{problem.Box}, problem.Items)
		require.NoError(tb, err)
		require.NotEmpty(tb, result.Boxes)

		total += volumeUtilisation(result.Boxes[0])
	}

	if len(problems) == 0 {
		return 0
	}

	return total / float64(len(problems))
}

func readBaseline(tb testing.TB) (map[string]float64, string) {
	tb.Helper()

	file, err := os.Open(filepath.Join("testdata", baselineFile))
	if os.IsNotExist(err) {
		return map[string]float64{}, ""
	}

	require.NoError(tb, err)

	defer func() { require.NoError(tb, file.Close()) }()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	require.NoError(tb, err)

	baseline := make(map[string]float64, len(records))
	settings := ""

	for _, record := range records {
		if len(record) == 2 && record[0] == "#settings" {
			settings = record[1]

			continue
		}

		require.Len(tb, record, 2)

		value, convErr := strconv.ParseFloat(record[1], 64)
		require.NoError(tb, convErr)

		baseline[record[0]] = value
	}

	return baseline, settings
}

func writeBaseline(tb testing.TB, baseline map[string]float64, settings baselineSettings) {
	tb.Helper()

	keys := slices.Sorted(maps.Keys(baseline))

	file, err := os.Create(filepath.Join("testdata", baselineFile))
	require.NoError(tb, err)

	defer func() { require.NoError(tb, file.Close()) }()

	writer := csv.NewWriter(file)
	require.NoError(tb, writer.Write([]string{"#settings", settings.String()}))

	for _, key := range keys {
		require.NoError(tb, writer.Write([]string{key, fmt.Sprintf("%.4f", baseline[key])}))
	}

	writer.Flush()
	require.NoError(tb, writer.Error())
}

func appendHistory(tb testing.TB, measured map[string]float64, settings baselineSettings) {
	tb.Helper()

	file, err := os.OpenFile(filepath.Join("testdata", historyFile),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	require.NoError(tb, err)

	defer func() { require.NoError(tb, file.Close()) }()

	stamp := time.Now().UTC().Format(time.RFC3339)
	writer := csv.NewWriter(file)

	for _, key := range slices.Sorted(maps.Keys(measured)) {
		require.NoError(tb, writer.Write([]string{
			stamp, settings.String(), key, fmt.Sprintf("%.4f", measured[key]),
		}))
	}

	writer.Flush()
	require.NoError(tb, writer.Error())
}

func TestPublished_Utilisation(t *testing.T) {
	t.Parallel()

	t.Log("entries marked +support enforce full base support; published results " +
		"for that constraint live in their own table and are not comparable with the rest")

	configurations := []struct {
		name     string
		strategy boxpacker3.RuleSettings
		support  float64
	}{
		{nameFirstFitDecreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectFirstFit}, 0},
		{nameBestFitDecreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectBestFit}, 0},
		{
			nameFirstFitDecreasing + "+support",
			boxpacker3.RuleSettings{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectFirstFit},
			1,
		},
	}

	baseline, recorded := readBaseline(t)
	measured := make(map[string]float64, len(publishedFiles())*len(configurations))

	instances := 0

	for _, name := range publishedFiles() {
		problems := loadPublishedInstances(t, name)
		instances = max(instances, len(problems))

		for _, configuration := range configurations {
			key := name + "/" + configuration.name
			measured[key] = meanUtilisation(t, problems, configuration.strategy, configuration.support)

			t.Logf("%-32s mean utilisation %.4f over %d problems", key, measured[key], len(problems))
		}
	}

	settings := currentSettings(instances)

	if *updateBaseline {
		maps.Copy(baseline, measured)
		writeBaseline(t, baseline, settings)
		appendHistory(t, measured, settings)
		t.Logf("baseline rewritten with %d entries under %s", len(baseline), settings)

		return
	}

	require.NotEmpty(t, baseline,
		"no baseline recorded; run with -update-baseline to create one")

	require.Equal(t, settings.String(), recorded,
		"the baseline was taken under other settings; a reading under %q is not a baseline for one under %q",
		recorded, settings)

	for key, got := range measured {
		want, ok := baseline[key]
		require.True(t, ok, "no baseline for %s; run with -update-baseline to record one", key)
		require.GreaterOrEqual(t, got, want-baselineTolerance,
			"%s: utilisation dropped from %.4f to %.4f", key, want, got)
	}
}
