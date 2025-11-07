package boxpacker3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3"
)

// TestPacker_StrategyMinimizeBoxes_MixedPackaging tests the specific case where
// StrategyMinimizeBoxes should pack all items into a single box (the medium one),
// but currently packs into two boxes incorrectly.
func TestPacker_StrategyMinimizeBoxes_MixedPackaging(t *testing.T) {
	t.Parallel()

	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyMinimizeBoxes))

	// Boxes: small, medium, large
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small", 220, 185, 50, 20000),
		boxpacker3.NewBox("medium", 425, 265, 190, 20000),
		boxpacker3.NewBox("large", 530, 380, 265, 20000),
	}

	// Products
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("item-1", 50, 30, 20, 250),
		boxpacker3.NewItem("item-2", 80, 60, 40, 500),
		boxpacker3.NewItem("item-3", 100, 70, 50, 600),
		boxpacker3.NewItem("item-4", 120, 80, 60, 800),
		boxpacker3.NewItem("item-5", 150, 100, 70, 1000),
		boxpacker3.NewItem("item-6", 180, 120, 80, 1200),
		boxpacker3.NewItem("item-7", 200, 150, 100, 1500),
		boxpacker3.NewItem("item-8", 90, 65, 45, 450),
		boxpacker3.NewItem("item-9", 110, 75, 55, 550),
		boxpacker3.NewItem("item-10", 130, 85, 60, 700),
		boxpacker3.NewItem("item-11", 140, 90, 65, 850),
		boxpacker3.NewItem("item-12", 160, 100, 70, 950),
		boxpacker3.NewItem("item-13", 170, 110, 75, 1100),
		boxpacker3.NewItem("item-14", 190, 130, 90, 1400),
		boxpacker3.NewItem("item-15", 210, 140, 100, 1600),
	}

	result := packer.Pack(boxes, items)

	require.NotNil(t, result)
	require.Empty(t, result.UnfitItems, "All items should be packed")

	// StrategyMinimizeBoxes should pack all items into a single box (medium)
	// FirstFitDecreasing can do it, so MinimizeBoxes should too
	// Count only boxes with items (empty boxes are still returned but don't count)
	usedBoxes := 0

	for _, box := range result.Boxes {
		if len(box.GetItems()) > 0 {
			usedBoxes++
		}
	}

	require.LessOrEqual(t, usedBoxes, 1,
		"StrategyMinimizeBoxes should pack all items into a single box (medium), but packed into %d boxes", usedBoxes)

	// Verify all items are packed
	totalPacked := 0
	for _, box := range result.Boxes {
		totalPacked += len(box.GetItems())
	}

	require.Equal(t, len(items), totalPacked, "All items should be packed")
}
