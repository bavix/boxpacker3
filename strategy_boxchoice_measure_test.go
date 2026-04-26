//go:build measure

package boxpacker3_test

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

type choice struct {
	containers float64
	volume     float64
	fill       float64
	spread     float64
	roomiest   int
	orders     int
}

func newChoice() *choice {
	return &choice{containers: 0, volume: 0, fill: 0, spread: 0, roomiest: 0, orders: 0}
}

func roomiestKind(boxes []*boxpacker3.Box) string {
	best, largest := "", 0.0

	for _, box := range boxes {
		if volume := box.Width() * box.Height() * box.Depth(); volume > largest {
			best, largest = box.ID(), volume
		}
	}

	return best
}

func (c *choice) record(result *boxpacker3.Result, roomiest string) {
	used := 0
	fill, weights, total := 0.0, make([]float64, 0, 4), 0.0

	for _, box := range result.Boxes {
		if len(box.Items) == 0 {
			continue
		}

		volume := box.Box.Width() * box.Box.Height() * box.Box.Depth()
		packed := 0.0

		for _, item := range box.Items {
			packed += item.Volume()
		}

		used++

		c.volume += volume
		fill += packed / volume

		weights = append(weights, box.Stats.ItemsWeight)
		total += box.Stats.ItemsWeight

		if box.ID() == roomiest {
			c.roomiest++
		}
	}

	c.containers += float64(used)
	c.orders++

	if used > 0 {
		c.fill += fill / float64(used)
	}

	c.spread += spreadOf(weights, total)
}

func spreadOf(weights []float64, total float64) float64 {
	if len(weights) < 2 {
		return 0
	}

	mean := total / float64(len(weights))
	variance := 0.0

	for _, weight := range weights {
		variance += (weight - mean) * (weight - mean)
	}

	return math.Sqrt(variance / float64(len(weights)))
}

//nolint:paralleltest // the harness times the work, so the runs must not overlap.
func TestMeasure_WhatEachRuleBuys(t *testing.T) {
	shelves := []struct {
		name  string
		build func() []*boxpacker3.Box
	}{
		{"a catalogue of seven sizes", catalogue},
		{"one kind of box, thirty of them", func() []*boxpacker3.Box {
			boxes := make([]*boxpacker3.Box, 0, 30)
			for i := range 30 {
				boxes = append(boxes, boxpacker3.NewBox(
					fmt.Sprintf("same-%d", i), 425, 265, 380, 20000))
			}

			return boxes
		}},

		{"equal bins, several needed", func() []*boxpacker3.Box {
			boxes := make([]*boxpacker3.Box, 0, 30)
			for i := range 30 {
				boxes = append(boxes, boxpacker3.NewBox(
					fmt.Sprintf("same-%d", i), 220, 180, 160, 20000))
			}

			return boxes
		}},
	}

	for _, shelf := range shelves {
		reportChoices(t, shelf.name, measureShelf(t, shelf.build))
	}
}

func randomOrderOf(random *rand.Rand) []*boxpacker3.Item {
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

func measureShelf(t *testing.T, build func() []*boxpacker3.Box) map[string]*choice {
	t.Helper()

	tally := map[string]*choice{}
	roomiest := roomiestKind(build())

	for seed := range 200 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec
		items := randomOrderOf(random)

		for _, rule := range boxpacker3.BoxSelections() {
			result, err := boxpacker3.NewPacker(
				boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, rule)),
			).Pack(t.Context(), build(), items)
			if err != nil {
				t.Fatal(err)
			}

			key := rule.String()
			if tally[key] == nil {
				tally[key] = newChoice()
			}

			tally[key].record(result, roomiest)
		}
	}

	return tally
}

//nolint:paralleltest // the harness times the work, so the runs must not overlap.
func TestMeasure_SpreadingRuleAgainstTheGoal(t *testing.T) {
	shelf := func() []*boxpacker3.Box {
		boxes := make([]*boxpacker3.Box, 0, 30)
		for i := range 30 {
			boxes = append(boxes, boxpacker3.NewBox(
				fmt.Sprintf("same-%d", i), 220, 180, 160, 20000))
		}

		return boxes
	}

	tally := map[string]*choice{}

	for seed := range 200 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec
		items := randomOrderOf(random)

		for _, way := range spreadingWays() {
			if tally[way.name] == nil {
				tally[way.name] = newChoice()
			}

			tally[way.name].record(way.pack(t, shelf(), items), "")
		}
	}

	reportChoices(t, "spreading the load: rule against goal", tally)
}

func reportChoices(t *testing.T, name string, tally map[string]*choice) {
	t.Helper()

	names := make([]string, 0, len(tally))
	for rule := range tally {
		names = append(names, rule)
	}

	sort.Slice(names, func(a, b int) bool {
		return tally[names[a]].volume < tally[names[b]].volume
	})

	t.Logf("--- %s ---", name)
	t.Log("rule                containers   volume paid   mean fill   weight spread   largest box")

	for _, rule := range names {
		c := tally[rule]
		orders := float64(c.orders)

		t.Logf("%-18s %10.2f %11.1f L %10.1f%% %15.0f %13.2f",
			rule, c.containers/orders, c.volume/orders/1e6,
			c.fill/orders*100, c.spread/orders, float64(c.roomiest)/orders)
	}
}

type packedBy struct {
	name string
	pack func(t *testing.T, boxes []*boxpacker3.Box, items []*boxpacker3.Item) *boxpacker3.Result
}

func byRule(selection boxpacker3.BoxSelection) func(*testing.T, []*boxpacker3.Box, []*boxpacker3.Item) *boxpacker3.Result {
	return func(t *testing.T, boxes []*boxpacker3.Box, items []*boxpacker3.Item) *boxpacker3.Result {
		t.Helper()

		result, err := boxpacker3.NewPacker(
			boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, selection)),
		).Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		return result
	}
}

func spreadingWays() []packedBy {
	return []packedBy{
		{name: "worst-fit", pack: byRule(boxpacker3.SelectWorstFit)},
		{name: "next-fit", pack: byRule(boxpacker3.SelectNextFit)},
		{name: "balanced goal", pack: func(
			t *testing.T, boxes []*boxpacker3.Box, items []*boxpacker3.Item,
		) *boxpacker3.Result {
			t.Helper()

			result, err := boxpacker3.NewPacker(
				boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.BalancedWeight)),
			).Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			return result
		}},
	}
}
