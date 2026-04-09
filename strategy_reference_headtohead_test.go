//go:build measure

package boxpacker3_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

type referenceRow struct {
	Order      string  `json:"order"`
	Containers int     `json:"containers"`
	Volume     float64 `json:"volume"`
	Unfit      int     `json:"unfit"`
}

func packingRuns(t *testing.T, set fixture) map[string]func(fixtureOrder) (*boxpacker3.Result, error) {
	t.Helper()

	runs := map[string]func(fixtureOrder) (*boxpacker3.Result, error){}

	for _, rule := range []boxpacker3.BoxSelection{
		boxpacker3.SelectFullestBox, boxpacker3.SelectFirstFit,
	} {
		runs[rule.String()] = func(order fixtureOrder) (*boxpacker3.Result, error) {
			return boxpacker3.NewPacker(
				boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, rule)),
			).Pack(t.Context(), boxesOf(set), itemsOf(order))
		}
	}

	runs["search/4/3"] = func(order fixtureOrder) (*boxpacker3.Result, error) {
		search, err := boxpacker3.NewSearch(128, 3)
		if err != nil {
			return nil, err
		}

		return boxpacker3.NewPacker(boxpacker3.WithAlgorithm(search)).
			Pack(t.Context(), boxesOf(set), itemsOf(order))
	}

	return runs
}

func readFixture(t *testing.T, name string) (fixture, []referenceRow) {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(referenceDir, name+".ref.json"))
	if err != nil {
		t.Skipf("no reference output for %s: run the PHP harness first", name)
	}

	var theirs []referenceRow

	err = json.Unmarshal(raw, &theirs)
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := os.ReadFile(filepath.Join(referenceDir, name+".json"))
	if err != nil {
		t.Fatal(err)
	}

	var set fixture

	err = json.Unmarshal(encoded, &set)
	if err != nil {
		t.Fatal(err)
	}

	return set, theirs
}

func compareRule(t *testing.T, name, rule string,
	set fixture, theirs []referenceRow,
	run func(fixtureOrder) (*boxpacker3.Result, error),
) {
	t.Helper()

	ours, better, worse, level := 0, 0, 0, 0

	for at, order := range set.Orders {
		result, err := run(order)
		if err != nil {
			t.Fatal(err)
		}

		used := usedBoxCountOf(result)
		ours += used

		switch {
		case used < theirs[at].Containers:
			better++
		case used > theirs[at].Containers:
			worse++
		default:
			level++
		}
	}

	total := 0
	for _, row := range theirs {
		total += row.Containers
	}

	t.Logf("%-18s %-12s ours %4d containers, reference %4d — better on %d orders, worse on %d, level on %d",
		name, rule, ours, total, better, worse, level)
}

func TestReference_HeadToHead(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"identical-cartons", "seven-sizes", "bookshop"} {
		set, theirs := readFixture(t, name)
		runs := packingRuns(t, set)

		for _, rule := range []string{"fullest-box", "first-fit", "search/4/3"} {
			compareRule(t, name, rule, set, theirs, runs[rule])
		}
	}
}
