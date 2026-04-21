package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestGoals_ConflictScenarios(t *testing.T) {
	t.Parallel()

	resA := &boxpacker3.Result{
		Boxes: []boxpacker3.PackedBox{
			makeBoxWithProps(100, 10, 10),
		},
		Unpacked: nil,
	}

	resB := &boxpacker3.Result{
		Boxes: []boxpacker3.PackedBox{
			makeBoxWithProps(10, 5, 5),
			makeBoxWithProps(10, 5, 5),
		},
		Unpacked: nil,
	}

	require.Negative(t, boxpacker3.FewestBoxes.Compare(resA, resB),
		"MinimizeBoxes should prefer 1 huge box over 2 small ones")

	require.Negative(t, boxpacker3.LeastVolume.Compare(resB, resA),
		"TightestPacking should prefer 2 small boxes (20L) over 1 huge box (100L)")

	require.Negative(t, boxpacker3.HighestFill.Compare(resB, resA),
		"MaxAverageFillRate should prefer higher density")
}

func TestGoals_BalancedPacking(t *testing.T) {
	t.Parallel()

	resA := &boxpacker3.Result{
		Boxes: []boxpacker3.PackedBox{
			makeBoxWithProps(20, 10, 10),
			makeBoxWithProps(20, 10, 10),
		},
		Unpacked: nil,
	}

	resB := &boxpacker3.Result{
		Boxes: []boxpacker3.PackedBox{
			makeBoxWithProps(20, 10, 1),
			makeBoxWithProps(20, 10, 19),
		},
		Unpacked: nil,
	}

	require.Negative(t, boxpacker3.BalancedWeight.Compare(resA, resB),
		"BalancedPacking should prefer equal weights")
}

func TestGoals_TieBreaker(t *testing.T) {
	t.Parallel()

	resA := &boxpacker3.Result{
		Boxes:    []boxpacker3.PackedBox{makeBoxWithProps(10, 5, 5)},
		Unpacked: nil,
	}
	resB := &boxpacker3.Result{
		Boxes:    []boxpacker3.PackedBox{makeBoxWithProps(20, 5, 5)},
		Unpacked: nil,
	}

	require.Negative(t, boxpacker3.FewestBoxes.Compare(resA, resB),
		"MinimizeBoxes should prefer smaller volume if box counts are equal")
}

func makeBoxWithProps(volume, itemsVolume, itemsWeight float64) boxpacker3.PackedBox {
	b := boxpacker3.NewBox("mock", volume, 1, 1, 1000)

	item := boxpacker3.NewItem("mock-item", itemsVolume, 1, 1, itemsWeight)

	result, err := boxpacker3.NewPacker().Pack(
		context.Background(), []*boxpacker3.Box{b}, []*boxpacker3.Item{item})
	if err != nil || len(result.Boxes) == 0 {
		panic("the mock item must pack")
	}

	return result.Boxes[0]
}
