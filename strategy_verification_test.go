package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

//nolint:funlen
func TestStrategy_AgreeOnIdenticalBinsAndDifferOnACatalogue(t *testing.T) {
	t.Parallel()

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("large-1", 60, 60, 60, 1000),
		boxpacker3.NewItem("large-2", 60, 60, 60, 1000),
		boxpacker3.NewItem("medium-1", 40, 40, 40, 500),
		boxpacker3.NewItem("medium-2", 40, 40, 40, 500),
		boxpacker3.NewItem("small-1", 20, 20, 20, 200),
		boxpacker3.NewItem("small-2", 20, 20, 20, 200),
		boxpacker3.NewItem("small-3", 20, 20, 20, 200),
	}

	counts := func(boxes []*boxpacker3.Box) map[string]int {
		used := map[string]int{}

		for _, strategy := range boxpacker3.EveryRuleSettings() {
			result, err := rulePacker(strategy.Order, strategy.Selection).
				Pack(t.Context(), boxes, items)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Empty(t, result.Unpacked,
				"%s left goods behind that fit", strategy)

			for _, box := range result.Boxes {
				if len(box.Items) > 0 {
					used[strategy.Name()]++
				}
			}
		}

		return used
	}

	identical := counts([]*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 5000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 5000),
		boxpacker3.NewBox("box-3", 100, 100, 100, 5000),
	})

	fewest := 0
	for _, count := range identical {
		if fewest == 0 || count < fewest {
			fewest = count
		}
	}

	for name, count := range identical {
		require.LessOrEqual(t, count, fewest*2,
			"%s used %d boxes where %d is enough, past its own bound", name, count, fewest)
	}

	catalogue := counts([]*boxpacker3.Box{
		boxpacker3.NewBox("tiny", 65, 65, 65, 5000),
		boxpacker3.NewBox("small", 85, 85, 85, 5000),
		boxpacker3.NewBox("roomy", 150, 150, 150, 5000),
		boxpacker3.NewBox("roomy-2", 150, 150, 150, 5000),
	})

	distinct := map[int]bool{}
	for _, count := range catalogue {
		distinct[count] = true
	}

	require.GreaterOrEqual(t, len(distinct), 2,
		"the rules made the same choice on a catalogue of sizes: %v", catalogue)
}

func TestStrategy_SortingOrder(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 200, 200, 200, 10000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("small", 10, 10, 10, 100),
		boxpacker3.NewItem("large", 50, 50, 50, 500),
		boxpacker3.NewItem("medium", 30, 30, 30, 300),
	}

	descendingStrategies := []boxpacker3.RuleSettings{
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectBestFit},
	}

	ascendingStrategies := []boxpacker3.RuleSettings{
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectNextFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectWorstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectAlmostWorstFit},
	}

	for _, strategy := range descendingStrategies {
		packer := rulePacker(strategy.Order, strategy.Selection)
		result, err := packer.Pack(context.Background(), boxes, items)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotEmpty(t, result.Boxes)

		firstBox := result.Boxes[0]
		if len(firstBox.Items) > 0 {
			firstItemID := firstBox.Items[0].ID()
			require.Equal(t, "large", firstItemID,
				"Strategy %d should pack large item first (descending sort)", strategy)
		}
	}

	for _, strategy := range ascendingStrategies {
		packer := rulePacker(strategy.Order, strategy.Selection)
		result, err := packer.Pack(context.Background(), boxes, items)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotEmpty(t, result.Boxes)

		firstBox := result.Boxes[0]
		if len(firstBox.Items) > 0 {
			firstItemID := firstBox.Items[0].ID()
			require.Equal(t, "small", firstItemID,
				"Strategy %d should pack small item first (ascending sort)", strategy)
		}
	}
}

func TestStrategy_NextFit_UsesCurrentBox(t *testing.T) {
	t.Parallel()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectNextFit)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box-1", 100, 100, 100, 5000),
		boxpacker3.NewBox("box-2", 100, 100, 100, 5000),
		boxpacker3.NewBox("box-3", 100, 100, 100, 5000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 60, 60, 60, 1000),
		boxpacker3.NewItem("item-2", 60, 60, 60, 1000),
		boxpacker3.NewItem("item-3", 60, 60, 60, 1000),
	}

	result, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Empty(t, result.Unpacked)

	usedBoxes := usedBoxCountOf(result)

	require.GreaterOrEqual(t, usedBoxes, 2, "Next Fit should use multiple boxes sequentially")
}
