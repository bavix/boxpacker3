package boxpacker3_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/bavix/boxpacker3"
)

const (
	// BoxTypeF -- 220, 185, 50, 20000.
	BoxTypeF = "8ec81501-11a4-4b3f-9a52-7cd2f9c8370c"

	// BoxTypeE -- 165, 215, 100, 20000.
	BoxTypeE = "9c69baf8-1ca3-46a0-9fc2-6f15ad9fef9a"

	// BoxTypeG -- 265, 165, 190, 20000.
	BoxTypeG = "2c5279d3-48ad-451b-b673-f6d9be7fc6f6"

	// BoxTypeC -- 425, 165, 190, 20000.
	BoxTypeC = "7f1cc68f-d554-4094-8734-c68df5c13154"

	// BoxTypeB -- 425, 265, 190, 20000.
	BoxTypeB = "76cede41-86bb-4487-bfb0-9513f032d53e"

	// BoxTypeA -- 425, 265, 380, 20000.
	BoxTypeA = "8e10cebf-cee6-4136-b060-1587b993d083"

	// BoxTypeStd -- 530, 380, 265, 20000.
	BoxTypeStd = "ba973206-aa64-493b-b37a-c53192cde8fd"

	// BoxTypeNotStd1 -- 1000, 500, 500, 20000.
	BoxTypeNotStd1 = "cb1ed5b8-7405-48c5-bfd0-d86f75c99261"

	// BoxTypeNotStd2 -- 1000, 1000, 1000, 20000.
	BoxTypeNotStd2 = "d91e2661-aebb-4a55-bfb5-4ff9c6e3c008"

	// BoxTypeNotStd3 -- 2000, 500, 500, 20000.
	BoxTypeNotStd3 = "a0ecd730-375a-4313-bbe8-820710606b3d"

	// BoxTypeNotStd4 -- 2000, 2000, 2000, 20000.
	BoxTypeNotStd4 = "6dff37f0-4dd1-4143-abdc-c19ab94f2e68"

	// BoxTypeNotStd5 -- 2500, 2500, 2500, 20000.
	BoxTypeNotStd5 = "abac6d59-b51f-4d62-a338-42aca7afe1cc"

	// BoxTypeNotStd6 -- 3000, 3000, 3000, 20000.
	BoxTypeNotStd6 = "981ffb30-a7b9-4d9e-820e-04de2145763e"
)

var defaultBoxes = []struct {
	Type               string
	W, H, L, MaxWeight float64
}{
	{BoxTypeF, 220, 185, 50, 20000},
	{BoxTypeE, 165, 215, 100, 20000},
	{BoxTypeG, 265, 165, 190, 20000},
	{BoxTypeC, 425, 165, 190, 20000},
	{BoxTypeB, 425, 265, 190, 20000},
	{BoxTypeA, 425, 265, 380, 20000},
	{BoxTypeStd, 530, 380, 265, 20000},
	{BoxTypeNotStd1, 1000, 500, 500, 20000},
	{BoxTypeNotStd2, 1000, 1000, 1000, 20000},
	{BoxTypeNotStd3, 2000, 500, 500, 20000},
	{BoxTypeNotStd4, 2000, 2000, 2000, 20000},
	{BoxTypeNotStd5, 2500, 2500, 2500, 20000},
	{BoxTypeNotStd6, 3000, 3000, 3000, 20000},
}

// NewDefaultBoxList creates a list of default boxes based on the predefined box types and dimensions.
func NewDefaultBoxList() []*boxpacker3.Box {
	// Initialize an empty list of boxes
	boxes := make([]*boxpacker3.Box, 0, len(defaultBoxes))
	// Iterate over the default box configurations and create a new box for each
	for _, box := range defaultBoxes {
		boxes = append(boxes, boxpacker3.NewBox(box.Type, box.W, box.H, box.L, box.MaxWeight))
	}

	return boxes
}

// PackerSuit is a test suite for the boxpacker3 package.
//
// It contains a number of tests for the boxpacker3 package, which can be run
// using the go test command.
type PackerSuit struct {
	suite.Suite
}

// TestBoxPackerSuite runs the PackerSuit test suite.
//
// It is a test function that uses the testing package to run the PackerSuit
// test suite. This function is used to test the boxpacker3 package.
//
// The test suite contains a number of tests for the boxpacker3 package, which
// can be run using the go test command.
//
// This function takes a testing.T object as a parameter and calls the Run
// function of the suite.Suite type, passing in the testing.T object and a
// pointer to a new PackerSuit instance.
func TestBoxPackerSuite(t *testing.T) {
	// Run the test suite in parallel
	t.Parallel()

	// Run the PackerSuit test suite and pass in a new PackerSuit instance
	suite.Run(t, new(PackerSuit))
}

// TestEmptyBoxes tests the Packer.Pack function with an empty list of boxes and items.
//
// It creates a new Packer instance and calls the Pack function with nil for the
// boxes and items parameters. It then verifies that the Pack function returns a
// non-nil PackResult, that the PackResult.Boxes slice is empty, and that the
// PackResult.UnfitItems slice is also empty.
func (s *PackerSuit) TestEmptyBoxes() {
	// Get the testing.T instance
	t := s.T()

	// Create a new Packer instance
	packer := boxpacker3.NewPacker()

	// Call the Pack function with nil for the boxes and items parameters
	packResult := packer.Pack(nil, nil)

	// Verify the PackResult
	require.NotNil(t, packResult, "Pack function returned nil")
	require.Empty(t, packResult.Boxes, "PackResult.Boxes is not empty")
	require.Empty(t, packResult.UnfitItems, "PackResult.UnfitItems is not empty")
}

// TestEmptyItems tests the Packer.Pack function with an empty list of items.
//
// It creates a new Packer instance and calls the Pack function with a non-nil
// boxes parameter (representing a list of boxes) and a nil items parameter.
// It then verifies that the Pack function returns a non-nil PackResult, that
// the PackResult.Boxes slice has the same length as the boxes parameter, and
// that the PackResult.UnfitItems slice is empty.
func (s *PackerSuit) TestEmptyItems() {
	// Get the testing.T instance
	t := s.T()

	// Create a new Packer instance
	packer := boxpacker3.NewPacker()

	// Create a list of default boxes
	boxes := NewDefaultBoxList()

	// Call the Pack function with the boxes and nil items parameters
	packResult := packer.Pack(boxes, nil)

	// Verify the PackResult
	require.NotNil(t, packResult, "Pack function returned nil")
	// Verify that the number of boxes is correct (all boxes should be returned, even empty)
	require.Len(t, packResult.Boxes, len(boxes), "PackResult.Boxes has incorrect length")
	// Verify that the UnfitItems slice is empty
	require.Empty(t, packResult.UnfitItems, "PackResult.UnfitItems is not empty")
}

// TestMinBox tests the Packer.Pack function with a single item that is smaller
// than the smallest box in the list of boxes.
//
// It creates a new Packer instance and calls the Pack function with a non-nil
// boxes parameter (representing a list of boxes) and a single item that is
// smaller than the smallest box in the boxes parameter. It then verifies that
// the Pack function returns a non-nil PackResult, that the PackResult.Boxes
// slice has the same length as the boxes parameter, and that the PackResult.
// UnfitItems slice is empty. It also verifies that the number of items in each
// box in the PackResult.Boxes slice is correct.
func (s *PackerSuit) TestMinBox() {
	// Get the testing.T instance
	t := s.T()

	// Create a new Packer instance
	packer := boxpacker3.NewPacker()

	// Create a list of default boxes
	boxes := NewDefaultBoxList()

	// Create a single item that is smaller than the smallest box in the boxes
	// parameter
	item := boxpacker3.NewItem(
		uuid.New().String(),
		8,
		17,
		5,
		384)

	// Call the Pack function with the boxes and a single item
	packResult := packer.Pack(boxes, []*boxpacker3.Item{item})

	// Verify the PackResult
	require.NotNil(t, packResult, "Pack function returned nil")
	// Verify that the number of boxes is correct (all boxes should be returned, even empty)
	require.Len(t, packResult.Boxes, len(boxes), "PackResult.Boxes has incorrect length")
	// Verify that the UnfitItems slice is empty
	require.Empty(t, packResult.UnfitItems, "PackResult.UnfitItems is not empty")

	// Verify that the number of items in each box is correct
	checks := map[string]int{
		BoxTypeF: 1,
	}

	for i := range packResult.Boxes {
		if len(packResult.Boxes[i].GetItems()) > 0 {
			require.Len(t, packResult.Boxes[i].GetItems(), checks[packResult.Boxes[i].GetID()])
		}
	}
}

// TestRotate tests the Packer.Pack function with a list of items that are
// rotated among the boxes.
//
// It creates a new Packer instance and calls the Pack function with a list of
// boxes and items. It then verifies that the Pack function returns a non-nil
// PackResult, that the PackResult.Boxes slice has the same length as the boxes
// parameter, and that the PackResult. UnfitItems slice is empty. It also
// verifies that the number of items in each box in the PackResult.Boxes slice
// is correct.
func (s *PackerSuit) TestRotate() {
	t := s.T()

	// Use StrategyGreedy to match old StrategyMinimizeBoxes behavior (First Fit with ascending sort)
	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))
	boxes := NewDefaultBoxList()

	items := []*boxpacker3.Item{
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 380, 100, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 250, 380, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
	}

	packResult := packer.Pack(boxes, items)
	require.NotNil(t, packResult)

	checks := map[string]int{
		BoxTypeStd: 5,
	}

	require.Empty(t, packResult.UnfitItems)

	for i := range len(packResult.Boxes) {
		require.Len(t, packResult.Boxes[i].GetItems(), checks[packResult.Boxes[i].GetID()], packResult.Boxes[i].GetID())
	}
}

// TestStd tests the Packer.Pack function with a list of standard items.
//
// It creates a new Packer instance and calls the Pack function with a list of
// boxes and standard items. It then verifies that the Pack function returns a
// non-nil PackResult, that the PackResult.Boxes slice has the same length as
// the boxes parameter, and that the PackResult. UnfitItems slice is empty. It
// also verifies that the number of items in each box in the PackResult.Boxes
// slice is correct.
func (s *PackerSuit) TestStd() {
	t := s.T()

	// Create a new Packer instance
	// Use StrategyGreedy to match old StrategyMinimizeBoxes behavior (First Fit with ascending sort)
	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))

	// Create a list of default boxes
	boxes := NewDefaultBoxList()

	// Create a list of standard items
	items := []*boxpacker3.Item{
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
	}

	// Call the Pack function with the boxes and standard items
	packResult := packer.Pack(boxes, items)

	// Verify the PackResult
	require.NotNil(t, packResult, "Pack function returned nil")

	// Verify that the number of boxes is correct (all boxes should be returned, even empty)
	require.Len(t, packResult.Boxes, len(boxes), "PackResult.Boxes has incorrect length")

	// Verify that the UnfitItems slice is empty
	require.Empty(t, packResult.UnfitItems, "PackResult.UnfitItems is not empty")

	// Verify that the number of items in each box is correct
	checks := map[string]int{
		BoxTypeStd: 5,
	}

	for i := range packResult.Boxes {
		if len(packResult.Boxes[i].GetItems()) > 0 {
			require.Len(t, packResult.Boxes[i].GetItems(), checks[packResult.Boxes[i].GetID()], packResult.Boxes[i].GetID())
		}
	}
}

// TestBoxTypeF tests the Packer.Pack function with a list of items that can fit
// into box type F.
//
// It creates a new Packer instance and calls the Pack function with a list of
// boxes and items that can fit into box type F. It then verifies that the Pack
// function returns a non-nil PackResult, that the PackResult.Boxes slice has
// the same length as the boxes parameter, and that the PackResult.UnfitItems
// slice is empty. It also verifies that the number of items in each box in the
// PackResult.Boxes slice is correct.
func (s *PackerSuit) TestBoxTypeF() {
	t := s.T()

	// Create a new Packer instance
	packer := boxpacker3.NewPacker()

	// Create a list of default boxes
	boxes := NewDefaultBoxList()

	// Create a list of items that can fit into box type F
	items := []*boxpacker3.Item{
		// 5
		boxpacker3.NewItem(uuid.New().String(), 100, 100, 5, 2500), // 5
		boxpacker3.NewItem(uuid.New().String(), 100, 5, 100, 2500), // 6
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2500), // 7
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2500), // 8
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2500), // 9
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2500), // 10

		// 35
		boxpacker3.NewItem(uuid.New().String(), 35, 100, 100, 2500), // 11
		boxpacker3.NewItem(uuid.New().String(), 35, 100, 100, 2500), // 12
	}

	// Call the Pack function with the boxes and items
	packResult := packer.Pack(boxes, items)

	// Verify the PackResult
	require.NotNil(t, packResult, "Pack function returned nil")

	// Verify that the number of boxes is correct (all boxes should be returned, even empty)
	require.Len(t, packResult.Boxes, len(boxes), "PackResult.Boxes has incorrect length")

	// Verify that the UnfitItems slice is empty
	require.Empty(t, packResult.UnfitItems, "PackResult.UnfitItems is not empty")

	// Verify that the number of items in each box is correct
	checks := map[string]int{
		BoxTypeF: 8,
	}

	for i := range packResult.Boxes {
		if len(packResult.Boxes[i].GetItems()) > 0 {
			require.Len(t, packResult.Boxes[i].GetItems(), checks[packResult.Boxes[i].GetID()], packResult.Boxes[i].GetID())
		}
	}
}

// TestBoxTypeF_Weight tests the Packer.Pack function with a list of items that can fit
// into box type F, but with a weight greater than the weight of box type E.
//
// It creates a new Packer instance and calls the Pack function with a list of boxes
// and items that can fit into box type F, but with a weight greater than the weight
// of box type E. It then verifies that the Pack function returns a non-nil PackResult,
// that the PackResult.Boxes slice has the same length as the boxes parameter, and
// that the PackResult.UnfitItems slice is empty. It also verifies that the number
// of items in each box in the PackResult.Boxes slice is correct.
func (s *PackerSuit) TestBoxTypeF_Weight() {
	t := s.T()

	packer := boxpacker3.NewPacker()
	boxes := NewDefaultBoxList()

	items := []*boxpacker3.Item{
		// 5
		boxpacker3.NewItem(uuid.New().String(), 100, 100, 5, 2690), // 5
		boxpacker3.NewItem(uuid.New().String(), 100, 5, 100, 2690), // 6
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2690), // 7
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2690), // 8
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2690), // 9
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2690), // 10

		// 35
		boxpacker3.NewItem(uuid.New().String(), 35, 100, 100, 2690), // 11
		boxpacker3.NewItem(uuid.New().String(), 35, 100, 100, 2690), // 12
	}

	packResult := packer.Pack(boxes, items)
	require.NotNil(t, packResult, "Pack function returned nil")

	// Verify that the number of boxes is correct (all boxes should be returned, even empty)
	require.Len(t, packResult.Boxes, len(boxes), "PackResult.Boxes has incorrect length")

	// Verify that the UnfitItems slice is empty
	require.Empty(t, packResult.UnfitItems, "PackResult.UnfitItems is not empty")

	// Verify that the number of items in each box is correct
	checks := map[string]int{
		BoxTypeF: 7,
		BoxTypeE: 1,
	}

	for i := range packResult.Boxes {
		require.Len(t, packResult.Boxes[i].GetItems(), checks[packResult.Boxes[i].GetID()], packResult.Boxes[i].GetID())
	}
}

// TestPacker_AllBoxes tests the Packer.Pack function with a variety of different
// boxes and items. It verifies that the number of boxes, unfit items, and number
// of items in each box are correct.
//
// It uses a predefined list of boxes and items and checks that the packer correctly
// assigns each item to a box.
func (s *PackerSuit) TestPacker_AllBoxes() {
	t := s.T()

	// Create a new Packer instance
	packer := boxpacker3.NewPacker()

	// Create a list of boxes
	boxes := NewDefaultBoxList()

	// Create a reverse list of boxes (used to iterate over the boxes in reverse order)
	reverse := make([]*boxpacker3.Box, len(boxes))
	for i := range boxes {
		reverse[i] = boxes[len(boxes)-1-i]
	}

	// Create a list of items
	items := []*boxpacker3.Item{
		// Large items
		boxpacker3.NewItem(uuid.New().String(), 1000, 1000, 1000, 20000),
		boxpacker3.NewItem(uuid.New().String(), 2000, 500, 500, 20000),
		boxpacker3.NewItem(uuid.New().String(), 2000, 2000, 2000, 20000),
		boxpacker3.NewItem(uuid.New().String(), 2500, 2500, 2500, 20000),
		boxpacker3.NewItem(uuid.New().String(), 3000, 3000, 3000, 20000),

		// Small items
		boxpacker3.NewItem(uuid.New().String(), 220, 185, 50, 20000),
		boxpacker3.NewItem(uuid.New().String(), 165, 215, 100, 20000),
		boxpacker3.NewItem(uuid.New().String(), 265, 165, 190, 20000),
		boxpacker3.NewItem(uuid.New().String(), 425, 165, 190, 20000),
		boxpacker3.NewItem(uuid.New().String(), 425, 265, 190, 20000),
		boxpacker3.NewItem(uuid.New().String(), 425, 265, 380, 20000),
		boxpacker3.NewItem(uuid.New().String(), 530, 380, 265, 20000),
		boxpacker3.NewItem(uuid.New().String(), 1000, 500, 500, 20000),
	}

	// Pack the items into the boxes
	packResult := packer.Pack(boxes, items)

	// Verify that the packing was successful
	require.NotNil(t, packResult, "Pack function returned nil")

	// Verify that the number of boxes is correct
	require.Len(t, packResult.Boxes, len(boxes), "PackResult.Boxes has incorrect length")

	// Verify that there are no unfit items
	require.Empty(t, packResult.UnfitItems, "PackResult.UnfitItems is not empty")

	// Verify that the number of items in each box is correct
	for i := range packResult.Boxes {
		// Each box should contain exactly one item
		require.Len(t, packResult.Boxes[i].GetItems(), 1, packResult.Boxes[i].GetID())
	}
}

// TestPacker_UnfitItems tests the Packer.Pack function with items that don't
// fit into any of the boxes. It verifies that the number of unfit items is
// correct and that no items are packed into any boxes.
func (s *PackerSuit) TestPacker_UnfitItems() {
	t := s.T()

	// Create a new Packer instance
	packer := boxpacker3.NewPacker()

	// Create a list of boxes
	boxes := NewDefaultBoxList()

	// Create a list of items that don't fit into any of the boxes
	items := []*boxpacker3.Item{
		boxpacker3.NewItem(uuid.New().String(), 3001, 3000, 3000, 20000), // Too large in all dimensions
		boxpacker3.NewItem(uuid.New().String(), 3000, 3001, 3000, 20000), // Too large in one dimension
		boxpacker3.NewItem(uuid.New().String(), 3000, 3000, 3001, 20000), // Too large in another dimension
		boxpacker3.NewItem(uuid.New().String(), 3000, 3000, 3000, 20001), // Too heavy
	}

	// Pack the items into the boxes
	packResult := packer.Pack(boxes, items)

	// Verify that the packing was successful
	require.NotNil(t, packResult, "Pack function returned nil")

	// Verify that the number of unfit items is correct
	require.Len(t, packResult.UnfitItems, 4, "PackResult.UnfitItems has incorrect length")

	// Verify that no items are packed into any boxes
	for i := range packResult.Boxes {
		require.Empty(t, packResult.Boxes[i].GetItems(), packResult.Boxes[i].GetID(), "Box "+packResult.Boxes[i].GetID()+" contains items")
	}
}

// TestPacker_MinAndStd tests the Packer.Pack function with a combination of
// standard and minimum sized items. It verifies that the packing is correct
// and that the number of items in each box is correct.
func (s *PackerSuit) TestPacker_MinAndStd() {
	t := s.T()

	// Create a new Packer instance
	// Use StrategyGreedy to match old StrategyMinimizeBoxes behavior (First Fit with ascending sort)
	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))

	// Create a list of boxes
	boxes := NewDefaultBoxList()
	reverse := make([]*boxpacker3.Box, len(boxes))

	// Reverse the list of boxes to guarantee that the packer will
	// pack items into boxes based on their volume, not their dimensions
	for i := range boxes {
		reverse[i] = boxes[len(boxes)-1-i]
	}

	// Create a list of items
	items := []*boxpacker3.Item{
		// Standard size items
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690), // 1
		boxpacker3.NewItem(uuid.New().String(), 380, 100, 250, 2690), // 2
		boxpacker3.NewItem(uuid.New().String(), 250, 380, 100, 2690), // 3
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690), // 4
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690), // 5

		// Minimum size items
		boxpacker3.NewItem(uuid.New().String(), 220, 185, 50, 20000),   // 6. F
		boxpacker3.NewItem(uuid.New().String(), 165, 215, 100, 20000),  // 7. E
		boxpacker3.NewItem(uuid.New().String(), 265, 165, 190, 20000),  // 8. G
		boxpacker3.NewItem(uuid.New().String(), 425, 165, 190, 20000),  // 9. C
		boxpacker3.NewItem(uuid.New().String(), 425, 265, 190, 20000),  // 10. B
		boxpacker3.NewItem(uuid.New().String(), 425, 265, 380, 20000),  // 11. A
		boxpacker3.NewItem(uuid.New().String(), 530, 380, 265, 20000),  // 12. Std
		boxpacker3.NewItem(uuid.New().String(), 1000, 500, 500, 20000), // 13. NotStd1

		// Maximum size items
		boxpacker3.NewItem(uuid.New().String(), 3000, 3000, 3000, 20000), // 14. NotStd6
	}

	// Pack the items into the boxes
	packResult := packer.Pack(reverse, items)

	// Verify that the packing was successful
	require.NotNil(t, packResult, "Pack function returned nil")

	// Verify that there are no unfit items
	require.Empty(t, packResult.UnfitItems)

	// Define the expected number of items in each box
	checks := map[string]int{
		BoxTypeF:       1, // 1
		BoxTypeE:       1, // 2
		BoxTypeG:       1, // 3
		BoxTypeC:       1, // 4
		BoxTypeB:       1, // 5
		BoxTypeA:       3, // 8
		BoxTypeStd:     1, // 9
		BoxTypeNotStd1: 1, // 10
		BoxTypeNotStd2: 1, // 11
		BoxTypeNotStd3: 1, // 12
		BoxTypeNotStd4: 1, // 13
		BoxTypeNotStd6: 1, // 14
	}

	// Verify that the number of items in each box is correct
	for i := range packResult.Boxes {
		require.Len(t, packResult.Boxes[i].GetItems(), checks[packResult.Boxes[i].GetID()], packResult.Boxes[i].GetID())
	}
}
