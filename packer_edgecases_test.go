package boxpacker3_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestPacker_Pack_NilInputs(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker()

	result, err := packer.Pack(context.Background(), nil, []*boxpacker3.Item{
		boxpacker3.NewItem("item1", 10, 10, 10, 1),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Boxes)
	require.NotNil(t, result.Unpacked)
	require.Len(t, result.Unpacked, 1, "Item should be in UnfitItems when no boxes provided")

	box := boxpacker3.NewBox("box1", 100, 100, 100, 100)
	result, err = packer.Pack(context.Background(), []*boxpacker3.Box{box}, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Boxes)
	require.NotNil(t, result.Unpacked)
	require.Empty(t, result.Unpacked, "No items should be in UnfitItems")

	result, err = packer.Pack(context.Background(), nil, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Boxes)
	require.NotNil(t, result.Unpacked)
	require.Empty(t, result.Boxes)
	require.Empty(t, result.Unpacked)
}

func TestPacker_Pack_NilElementsInSlices(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker()

	box := boxpacker3.NewBox("box1", 100, 100, 100, 100)
	item := boxpacker3.NewItem("item1", 10, 10, 10, 1)

	result, err := packer.Pack(context.Background(), []*boxpacker3.Box{box, nil, box}, []*boxpacker3.Item{item})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.GreaterOrEqual(t, len(result.Boxes), 1, "Should have at least one box")

	result, err = packer.Pack(context.Background(), []*boxpacker3.Box{box}, []*boxpacker3.Item{item, nil, item})
	require.NoError(t, err)
	require.NotNil(t, result)

	totalItems := len(result.Unpacked)
	for _, b := range result.Boxes {
		totalItems += len(b.Items)
	}

	require.GreaterOrEqual(t, totalItems, 2, "Should have at least 2 items (non-nil items)")
	require.LessOrEqual(t, totalItems, 3, "Should have at most 3 items (2 non-nil + possibly 1 nil in unpacked)")
}

func TestPacker_PackCtx_ContextCancellation(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker()

	items := make([]*boxpacker3.Item, 1000)
	for i := range items {
		items[i] = boxpacker3.NewItem(
			"item"+strconv.Itoa(i),
			float64(10+i%10),
			float64(10+i%10),
			float64(10+i%10),
			1,
		)
	}

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box1", 100, 100, 100, 1000),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := packer.Pack(ctx, boxes, items)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, context.Canceled, err)
}

func TestPacker_PackCtx_Timeout(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker()

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item1", 10, 10, 10, 1),
	}

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box1", 100, 100, 100, 100),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)

	result, err := packer.Pack(ctx, boxes, items)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, context.DeadlineExceeded, err)
}

func TestPacker_PackCtx_Success(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker()

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item1", 10, 10, 10, 1),
	}

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("box1", 100, 100, 100, 100),
	}

	ctx := context.Background()

	result, err := packer.Pack(ctx, boxes, items)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.GreaterOrEqual(t, len(result.Boxes), 1, "Should have at least one box")
}

func TestNewItem_NoValidation(t *testing.T) {
	t.Parallel()

	item := boxpacker3.NewItem("item1", 0, 0, 0, 0)
	require.NotNil(t, item)
	require.InDelta(t, 0.0, item.Width(), 0.0001, "Width should be exactly 0")
	require.InDelta(t, 0.0, item.Height(), 0.0001, "Height should be exactly 0")
	require.InDelta(t, 0.0, item.Depth(), 0.0001, "Depth should be exactly 0")
	require.InDelta(t, 0.0, item.Weight(), 0.0001, "Weight should be exactly 0")

	item = boxpacker3.NewItem("item2", -10, -20, -30, -5)
	require.NotNil(t, item)
	require.InDelta(t, -10.0, item.Width(), 0.0001, "Width should be exactly -10")
	require.InDelta(t, -20.0, item.Height(), 0.0001, "Height should be exactly -20")
	require.InDelta(t, -30.0, item.Depth(), 0.0001, "Depth should be exactly -30")
	require.InDelta(t, -5.0, item.Weight(), 0.0001, "Weight should be exactly -5")

	item = boxpacker3.NewItem("", 10, 10, 10, 1)
	require.NotNil(t, item)
	require.Empty(t, item.ID(), "ID should be empty string")
}

func TestNewBox_NoValidation(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("box1", 0, 0, 0, 0)
	require.NotNil(t, box)
	require.InDelta(t, 0.0, box.Width(), 0.0001, "Width should be exactly 0")
	require.InDelta(t, 0.0, box.Height(), 0.0001, "Height should be exactly 0")
	require.InDelta(t, 0.0, box.Depth(), 0.0001, "Depth should be exactly 0")
	require.InDelta(t, 0.0, box.MaxWeight(), 0.0001, "MaxWeight should be exactly 0")

	box = boxpacker3.NewBox("box2", -10, -20, -30, -5)
	require.NotNil(t, box)
	require.InDelta(t, -10.0, box.Width(), 0.0001, "Width should be exactly -10")
	require.InDelta(t, -20.0, box.Height(), 0.0001, "Height should be exactly -20")
	require.InDelta(t, -30.0, box.Depth(), 0.0001, "Depth should be exactly -30")
	require.InDelta(t, -5.0, box.MaxWeight(), 0.0001, "MaxWeight should be exactly -5")

	box = boxpacker3.NewBox("", 100, 100, 100, 100)
	require.NotNil(t, box)
	require.Empty(t, box.ID(), "ID should be empty string")
}

func TestPacker_AllStrategies_NilHandling(t *testing.T) {
	t.Parallel()

	strategies := []boxpacker3.RuleSettings{
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectFirstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectBestFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectNextFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectWorstFit},
		{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectAlmostWorstFit},
	}

	box := boxpacker3.NewBox("box1", 100, 100, 100, 100)
	item := boxpacker3.NewItem("item1", 10, 10, 10, 1)

	strategyNames := []string{
		"FirstFitDecreasing",
		"FirstFitIncreasing",
		"BestFitIncreasing",
		"BestFitDecreasing",
		"NextFitIncreasing",
		"WorstFitIncreasing",
		"AlmostWorstFitIncreasing",
	}

	for i, strategy := range strategies {
		t.Run(strategyNames[i], func(t *testing.T) {
			t.Parallel()

			packer := rulePacker(strategy.Order, strategy.Selection)

			result, err := packer.Pack(context.Background(), []*boxpacker3.Box{box}, []*boxpacker3.Item{item, nil, item})
			require.NoError(t, err)
			require.NotNil(t, result)

			result, err = packer.Pack(context.Background(), []*boxpacker3.Box{box, nil, box}, []*boxpacker3.Item{item})
			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
}
