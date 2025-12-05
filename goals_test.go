package boxpacker3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

func TestGoals_ConflictScenarios(t *testing.T) {
	t.Parallel()

	// A: 1 Huge Box (100L capacity).
	//   - Items inside take 10L.
	//   - Fill rate: 10%.
	//   - Total Volume: 100L.
	//   - Box Count: 1.
	//
	// B: 2 Small Boxes (10L capacity each).
	//   - Items inside take 10L (5L per box).
	//   - Fill rate: 50% per box.
	//   - Total Volume: 20L.
	//   - Box Count: 2.

	resA := &boxpacker3.Result{
		Boxes: []*boxpacker3.Box{
			makeBoxWithProps(100, 10, 10),
		},
		UnfitItems: []*boxpacker3.Item{},
	}

	resB := &boxpacker3.Result{
		Boxes: []*boxpacker3.Box{
			makeBoxWithProps(10, 5, 5),
			makeBoxWithProps(10, 5, 5),
		},
		UnfitItems: []*boxpacker3.Item{},
	}

	// MinimizeBoxes should prefer A (1 box < 2 boxes).
	require.True(t, boxpacker3.MinimizeBoxesGoal(resA, resB),
		"MinimizeBoxes should prefer 1 huge box over 2 small ones")

	// TightestPacking should prefer B (20L total volume < 100L total volume).
	require.True(t, boxpacker3.TightestPackingGoal(resB, resA),
		"TightestPacking should prefer 2 small boxes (20L) over 1 huge box (100L)")

	// MaxAverageFillRate should prefer B (50% fill vs 10% fill).
	require.True(t, boxpacker3.MaxAverageFillRateGoal(resB, resA),
		"MaxAverageFillRate should prefer higher density")
}

func TestGoals_BalancedPacking(t *testing.T) {
	t.Parallel()

	// Candidate A: Balanced (10kg, 10kg). StdDev = 0.
	resA := &boxpacker3.Result{
		Boxes: []*boxpacker3.Box{
			makeBoxWithProps(20, 10, 10), // 10 kg
			makeBoxWithProps(20, 10, 10), // 10 kg
		},
	}

	// Candidate B: Unbalanced (1kg, 19kg). High StdDev.
	resB := &boxpacker3.Result{
		Boxes: []*boxpacker3.Box{
			makeBoxWithProps(20, 10, 1),  // 1 kg
			makeBoxWithProps(20, 10, 19), // 19 kg
		},
	}

	require.True(t, boxpacker3.BalancedPackingGoal(resA, resB),
		"BalancedPacking should prefer equal weights")
}

func TestGoals_TieBreaker(t *testing.T) {
	t.Parallel()

	// Both use 1 box.
	// A is smaller (10L). B is larger (20L).
	resA := &boxpacker3.Result{Boxes: []*boxpacker3.Box{makeBoxWithProps(10, 5, 5)}}
	resB := &boxpacker3.Result{Boxes: []*boxpacker3.Box{makeBoxWithProps(20, 5, 5)}}

	// MinimizeBoxes should fall back to volume check if counts are equal
	require.True(t, boxpacker3.MinimizeBoxesGoal(resA, resB),
		"MinimizeBoxes should prefer smaller volume if box counts are equal")
}

// makeBoxWithProps creates a real Box struct and populates it with an item
// to simulate volume and weight usage for testing goals.
func makeBoxWithProps(volume, itemsVolume, itemsWeight float64) *boxpacker3.Box {
	// Create a box with W=volume, H=1, D=1 => Volume = volume
	b := boxpacker3.NewBox("mock", volume, 1, 1, 1000)

	// Create an item with matching props to populate the box stats
	item := boxpacker3.NewItem("mock-item", itemsVolume, 1, 1, itemsWeight)

	// PutItem updates internal itemsVolume and itemsWeight
	b.PutItem(item, boxpacker3.Pivot{})

	return b
}
