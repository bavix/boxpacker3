//go:build measure

package boxpacker3_test

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

func TestMeasure_FloatingItems(t *testing.T) {
	t.Parallel()

	for _, support := range []float64{0, 0.3, 0.6, 1} {
		floating, total, boxesWith := 0, 0, 0

		for seed := range 300 {
			random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

			boxes := make([]*boxpacker3.Box, 0, 4)

			for i := range 2 + random.Intn(3) {
				side := 120 + float64(random.Intn(180))
				boxes = append(boxes, boxpacker3.NewBox(
					fmt.Sprintf("box-%d", i), side, side*0.8, side*0.7, 1e9))
			}

			items := make([]*boxpacker3.Item, 0, 24)
			for i := range 8 + random.Intn(16) {
				items = append(items, boxpacker3.NewItem(
					fmt.Sprintf("item-%d", i),
					20+float64(random.Intn(80)),
					20+float64(random.Intn(80)),
					20+float64(random.Intn(80)), 100))
			}

			result, err := boxpacker3.NewPacker(
				boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: support})).Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			for _, box := range result.Boxes {
				hanging := 0

				for _, item := range box.Items {
					total++

					if supportUnder(box, item) <= 1e-9 {
						floating++
						hanging++
					}
				}

				if hanging > 0 {
					boxesWith++
				}
			}
		}

		t.Logf("support %.1f: %d of %d pieces rest on nothing, in %d containers",
			support, floating, total, boxesWith)
	}
}

func catalogue() []*boxpacker3.Box {
	sizes := [][4]float64{
		{220, 185, 50, 20000},
		{165, 215, 100, 20000},
		{265, 165, 190, 20000},
		{425, 165, 190, 20000},
		{425, 265, 190, 20000},
		{425, 265, 380, 20000},
		{530, 380, 265, 20000},
	}

	boxes := make([]*boxpacker3.Box, 0, len(sizes))
	for i, size := range sizes {
		boxes = append(boxes, boxpacker3.NewBox(
			fmt.Sprintf("kind-%d", i), size[0], size[1], size[2], size[3]))
	}

	return boxes
}

type score struct {
	containers float64
	volume     float64
	unfit      float64
	runs       int
}

func TestMeasure_DoTheNamesHold(t *testing.T) {
	t.Parallel()

	tally := map[string]*score{}

	for seed := range 250 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		items := make([]*boxpacker3.Item, 0, 40)
		for i := range 10 + random.Intn(30) {
			items = append(items, boxpacker3.NewItem(
				fmt.Sprintf("item-%d", i),
				30+float64(random.Intn(150)),
				30+float64(random.Intn(150)),
				20+float64(random.Intn(120)),
				float64(100+random.Intn(800))))
		}

		for _, strategy := range boxpacker3.EveryRuleSettings() {
			result, err := rulePacker(strategy.Order, strategy.Selection).
				Pack(t.Context(), catalogue(), items)
			if err != nil {
				t.Fatal(err)
			}

			name := strategy.Name()
			if tally[name] == nil {
				tally[name] = &score{containers: 0, volume: 0, unfit: 0, runs: 0}
			}

			used, volume := 0, 0.0

			for _, box := range result.Boxes {
				if len(box.Items) == 0 {
					continue
				}

				used++
				volume += box.Box.Width() * box.Box.Height() * box.Box.Depth()
			}

			tally[name].containers += float64(used)
			tally[name].volume += volume
			tally[name].unfit += float64(len(result.Unpacked))
			tally[name].runs++
		}
	}

	reportScores(t, tally)
}

func reportScores(t *testing.T, tally map[string]*score) {
	t.Helper()

	names := make([]string, 0, len(tally))
	for name := range tally {
		names = append(names, name)
	}

	sort.Slice(names, func(a, b int) bool {
		return tally[names[a]].containers/float64(tally[names[a]].runs) <
			tally[names[b]].containers/float64(tally[names[b]].runs)
	})

	t.Log("strategy               containers   container volume   left behind")

	for _, name := range names {
		s := tally[name]
		runs := float64(s.runs)
		t.Logf("%-22s %10.2f %14.1f L %13.2f",
			name, s.containers/runs, s.volume/runs/1e6, s.unfit/runs)
	}
}

type binRow struct{ containers, volume float64 }

func randomOrderFor(random *rand.Rand) []*boxpacker3.Item {
	items := make([]*boxpacker3.Item, 0, 40)

	for i := range 12 + random.Intn(26) {
		items = append(items, boxpacker3.NewItem(
			fmt.Sprintf("item-%d", i),
			30+float64(random.Intn(140)),
			30+float64(random.Intn(140)),
			20+float64(random.Intn(110)),
			float64(100+random.Intn(500))))
	}

	return items
}

func measureBins(t *testing.T, name string, boxesFor func() []*boxpacker3.Box) {
	t.Helper()

	tally := map[string]*binRow{}

	for seed := range 200 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec
		items := randomOrderFor(random)

		for _, rule := range boxpacker3.BoxSelections() {
			result, err := boxpacker3.NewPacker(
				boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, rule)),
			).Pack(t.Context(), boxesFor(), items)
			if err != nil {
				t.Fatal(err)
			}

			key := rule.String()
			if tally[key] == nil {
				tally[key] = &binRow{containers: 0, volume: 0}
			}

			for _, box := range result.Boxes {
				if len(box.Items) == 0 {
					continue
				}

				tally[key].containers++
				tally[key].volume += box.Box.Width() * box.Box.Height() * box.Box.Depth()
			}
		}
	}

	t.Logf("--- %s ---", name)

	for _, rule := range boxpacker3.BoxSelections() {
		r := tally[rule.String()]
		t.Logf("%-18s %6.2f containers %10.1f L", rule, r.containers/200, r.volume/200/1e6)
	}
}

func TestMeasure_IdenticalBinsVersusCatalogue(t *testing.T) {
	t.Parallel()

	measureBins(t, "one kind of box, thirty of them", func() []*boxpacker3.Box {
		boxes := make([]*boxpacker3.Box, 0, 30)
		for i := range 30 {
			boxes = append(boxes, boxpacker3.NewBox(
				fmt.Sprintf("same-%d", i), 425, 265, 380, 20000))
		}

		return boxes
	})

	measureBins(t, "a catalogue of seven sizes", catalogue)
}
