//go:build measure

package boxpacker3_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

type analysisReading struct {
	boxes   float64
	volume  float64
	fill    float64
	unfit   int
	gap     float64
	elapsed time.Duration
	atBound int
	orders  int
}

func analysisShelves() []struct {
	name  string
	build func() []*boxpacker3.Box
} {
	return []struct {
		name  string
		build func() []*boxpacker3.Box
	}{
		{"seven sizes", catalogue},
		{"one size, thirty of them", func() []*boxpacker3.Box {
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
}

type analysisWay struct {
	name    string
	options []boxpacker3.Option
}

func analysisWays(t *testing.T) []analysisWay {
	t.Helper()

	search, err := boxpacker3.NewSearch(128, 5)
	require.NoError(t, err)

	ways := []analysisWay{}

	for _, rule := range boxpacker3.EveryRuleSettings() {
		ways = append(ways, analysisWay{
			name: rule.Name(),
			options: []boxpacker3.Option{
				boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(rule.Order, rule.Selection)),
			},
		})
	}

	return append(ways, analysisExtras(t, search)...)
}

func analysisExtras(t *testing.T, search *boxpacker3.Search) []analysisWay {
	t.Helper()

	fullest := boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectFullestBox)
	ways := []analysisWay{}

	for _, goal := range boxpacker3.Goals() {
		answer, ok := boxpacker3.BestForName(goal.Name())
		require.True(t, ok)

		ways = append(ways, analysisWay{
			name:    "BestFor(" + goal.Name() + ")",
			options: []boxpacker3.Option{boxpacker3.WithAlgorithm(answer)},
		})
	}

	return append(ways,
		analysisWay{name: search.Name(), options: []boxpacker3.Option{boxpacker3.WithAlgorithm(search)}},
		analysisWay{name: "FullestBoxDecreasing + search filler", options: []boxpacker3.Option{
			boxpacker3.WithAlgorithm(fullest), boxpacker3.WithFiller(search),
		}},
		analysisWay{name: "FullestBoxDecreasing + space corners", options: []boxpacker3.Option{
			boxpacker3.WithAlgorithm(fullest),
			boxpacker3.WithRules(boxpacker3.Rules{FreeSpaceCorners: true}),
		}},
		analysisWay{name: "FullestBoxDecreasing + rehome", options: []boxpacker3.Option{
			boxpacker3.WithAlgorithm(fullest),
			boxpacker3.WithFinishers(boxpacker3.Rehome, boxpacker3.BalanceWeight{MaxBoxes: 12}),
		}},
	)
}

func TestMeasure_Analysis(t *testing.T) {
	t.Parallel()

	const orders = 120

	for _, shelf := range analysisShelves() {
		t.Logf("=== %s", shelf.name)
		t.Logf("%-38s %8s %10s %7s %6s %6s %9s", "", "boxes", "volume", "fill", "unfit", "atLB", "time")

		for _, way := range analysisWays(t) {
			reading := measureWay(t, way.options, shelf.build, orders)

			t.Logf("%-38s %8.2f %10.1f %6.1f%% %6d %5d%% %9s",
				way.name,
				reading.boxes/float64(reading.orders),
				reading.volume/float64(reading.orders)/1e6,
				reading.fill/float64(reading.orders)*100,
				reading.unfit,
				reading.atBound*100/reading.orders,
				(reading.elapsed / time.Duration(reading.orders)).Round(time.Microsecond))
		}
	}
}

func measureWay(
	t *testing.T, options []boxpacker3.Option, build func() []*boxpacker3.Box, orders int,
) analysisReading {
	t.Helper()

	packer := boxpacker3.NewPacker(options...)
	reading := analysisReading{
		boxes: 0, volume: 0, fill: 0, unfit: 0, gap: 0, elapsed: 0, atBound: 0, orders: orders,
	}

	for seed := range orders {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec
		items := randomOrderOf(random)

		started := time.Now()

		result, err := packer.Pack(t.Context(), build(), items)
		require.NoError(t, err)

		reading.elapsed += time.Since(started)
		reading.boxes += float64(len(result.Boxes))
		reading.unfit += len(result.Unpacked)
		reading.fill += result.Report.Fill
		reading.gap += float64(result.Report.GapBoxes)

		if result.Report.GapBoxes == 0 && len(result.Unpacked) == 0 {
			reading.atBound++
		}

		for _, box := range result.Boxes {
			reading.volume += box.Volume()
		}
	}

	return reading
}

func TestMeasure_DistanceToTheBound(t *testing.T) {
	t.Parallel()

	const orders = 200

	for _, shelf := range analysisShelves() {
		t.Logf("=== %s", shelf.name)

		for _, rule := range boxpacker3.EveryRuleSettings() {
			packer := boxpacker3.NewPacker(
				boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(rule.Order, rule.Selection)))

			atBound, gap, worst := 0, 0, 0

			for seed := range orders {
				random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

				result, err := packer.Pack(t.Context(), shelf.build(), randomOrderOf(random))
				require.NoError(t, err)

				if len(result.Unpacked) > 0 {
					continue
				}

				gap += result.Report.GapBoxes
				worst = max(worst, result.Report.GapBoxes)

				if result.Report.GapBoxes == 0 {
					atBound++
				}
			}

			t.Logf("%-28s meets the bound on %3d%% of orders, %.2f boxes above it on average, %d at worst",
				rule.Name(), atBound*100/orders, float64(gap)/float64(orders), worst)
		}
	}
}

func TestMeasure_GoalsAgainstTheirRules(t *testing.T) {
	t.Parallel()

	const orders = 100

	measures := map[string]func(*boxpacker3.Result) float64{
		nameFewestBoxes:    func(r *boxpacker3.Result) float64 { return float64(len(r.Boxes)) },
		nameLeastVolume:    func(r *boxpacker3.Result) float64 { return boxpacker3.MetricsOf(r).BoxVolume },
		nameMostItems:      func(r *boxpacker3.Result) float64 { return float64(len(r.Unpacked)) },
		nameHighestFill:    func(r *boxpacker3.Result) float64 { return -boxpacker3.MetricsOf(r).Fill },
		nameBalancedWeight: func(r *boxpacker3.Result) float64 { return boxpacker3.MetricsOf(r).WeightSpread },
	}

	for _, shelf := range analysisShelves() {
		t.Logf("=== %s", shelf.name)

		for _, goal := range boxpacker3.Goals() {
			answer, ok := boxpacker3.BestForName(goal.Name())
			require.True(t, ok)

			measure := measures[goal.Name()]
			beaten, level, better := 0, 0, 0

			for seed := range orders {
				random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec
				items := randomOrderOf(random)

				goalResult, err := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(answer)).
					Pack(t.Context(), shelf.build(), items)
				require.NoError(t, err)

				best := math.Inf(1)

				for _, rule := range boxpacker3.EveryRuleSettings() {
					ruleResult, err := boxpacker3.NewPacker(
						boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(rule.Order, rule.Selection))).
						Pack(t.Context(), shelf.build(), items)
					require.NoError(t, err)

					if len(ruleResult.Unpacked) == len(goalResult.Unpacked) {
						best = math.Min(best, measure(ruleResult))
					}
				}

				switch {
				case measure(goalResult) > best+1e-9:
					beaten++
				case measure(goalResult) < best-1e-9:
					better++
				default:
					level++
				}
			}

			t.Logf("%-18s beaten by a rule on %3d orders, level on %3d, better on %3d",
				goal.Name(), beaten, level, better)
		}
	}
}
