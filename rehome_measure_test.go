//go:build measure

package boxpacker3_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestMeasure_Rehome(t *testing.T) {
	t.Parallel()

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
	}

	for _, shelf := range shelves {
		t.Logf("=== %s", shelf.name)

		for _, rule := range boxpacker3.EveryRuleSettings() {
			plain := measureShelfRun(t, shelf.build, rule, false)
			rehomed := measureShelfRun(t, shelf.build, rule, true)

			t.Logf("%-26s containers %5.2f → %5.2f, volume paid %8.0f → %8.0f (%+.1f%%), unfit %d → %d",
				rule.Name(),
				plain.containers, rehomed.containers,
				plain.volume, rehomed.volume,
				(rehomed.volume-plain.volume)/plain.volume*100,
				plain.unfit, rehomed.unfit)
		}
	}
}

type rehomeReading struct {
	containers float64
	volume     float64
	unfit      int
}

func measureShelfRun(
	t *testing.T, build func() []*boxpacker3.Box, rule boxpacker3.RuleSettings, rehome bool,
) rehomeReading {
	t.Helper()

	finishers := []boxpacker3.Finisher{}
	if rehome {
		finishers = append(finishers, boxpacker3.Rehome)
	}

	packer := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(rule.Order, rule.Selection)),
		boxpacker3.WithFinishers(finishers...),
	)

	const orders = 200

	reading := rehomeReading{containers: 0, volume: 0, unfit: 0}

	for seed := range orders {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		result, err := packer.Pack(t.Context(), build(), randomOrderOf(random))
		require.NoError(t, err)

		reading.containers += float64(len(result.Boxes))
		reading.unfit += len(result.Unpacked)

		for _, box := range result.Boxes {
			reading.volume += box.Volume()
		}
	}

	reading.containers /= orders

	return reading
}
