//nolint:dupl
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

type PackerSuit struct {
	suite.Suite
}

func TestBoxPackerSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(PackerSuit))
}

func NewDefaultBoxList() []*boxpacker3.Box {
	return []*boxpacker3.Box{
		boxpacker3.NewBox(BoxTypeF, 220, 185, 50, 20000),           // 0
		boxpacker3.NewBox(BoxTypeE, 165, 215, 100, 20000),          // 1
		boxpacker3.NewBox(BoxTypeG, 265, 165, 190, 20000),          // 2
		boxpacker3.NewBox(BoxTypeC, 425, 165, 190, 20000),          // 3
		boxpacker3.NewBox(BoxTypeB, 425, 265, 190, 20000),          // 4
		boxpacker3.NewBox(BoxTypeA, 425, 265, 380, 20000),          // 5
		boxpacker3.NewBox(BoxTypeStd, 530, 380, 265, 20000),        // 6
		boxpacker3.NewBox(BoxTypeNotStd1, 1000, 500, 500, 20000),   // 7
		boxpacker3.NewBox(BoxTypeNotStd2, 1000, 1000, 1000, 20000), // 8
		boxpacker3.NewBox(BoxTypeNotStd3, 2000, 500, 500, 20000),   // 9
		boxpacker3.NewBox(BoxTypeNotStd4, 2000, 2000, 2000, 20000), // 10
		boxpacker3.NewBox(BoxTypeNotStd5, 2500, 2500, 2500, 20000), // 11
		boxpacker3.NewBox(BoxTypeNotStd6, 3000, 3000, 3000, 20000), // 12
	}
}

func (s *PackerSuit) TestMinBox() {
	t := s.T()
	t.Parallel()

	packer := boxpacker3.NewPacker()
	boxes := NewDefaultBoxList()
	item := boxpacker3.NewItem(
		uuid.New().String(),
		8,
		17,
		5,
		384)

	packResult, err := packer.Pack(boxes, []*boxpacker3.Item{item})
	require.NoError(t, err)
	require.NotNil(t, packResult)

	checks := map[string]int{
		BoxTypeF: 1,
	}

	require.Len(t, packResult.UnfitItems, 0)

	for i := 0; i < len(packResult.Boxes); i++ {
		require.Len(t, packResult.Boxes[i].GetItems(), checks[packResult.Boxes[i].GetID()])
	}
}

func (s *PackerSuit) TestRotate() {
	t := s.T()
	t.Parallel()

	packer := boxpacker3.NewPacker()
	boxes := NewDefaultBoxList()

	items := []*boxpacker3.Item{
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 380, 100, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 250, 380, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
	}

	packResult, err := packer.Pack(boxes, items)
	require.NoError(t, err)
	require.NotNil(t, packResult)

	checks := map[string]int{
		BoxTypeStd: 5,
	}

	require.Len(t, packResult.UnfitItems, 0)

	for i := 0; i < len(packResult.Boxes); i++ {
		require.Len(t, packResult.Boxes[i].GetItems(), checks[packResult.Boxes[i].GetID()], packResult.Boxes[i].GetID())
	}
}

func (s *PackerSuit) TestStd() {
	t := s.T()
	t.Parallel()

	packer := boxpacker3.NewPacker()
	boxes := NewDefaultBoxList()

	items := []*boxpacker3.Item{
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
	}

	packResult, err := packer.Pack(boxes, items)
	require.NoError(t, err)
	require.NotNil(t, packResult)

	checks := map[string]int{
		BoxTypeStd: 5,
	}

	require.Len(t, packResult.UnfitItems, 0)

	for i := 0; i < len(packResult.Boxes); i++ {
		require.Len(t, packResult.Boxes[i].GetItems(), checks[packResult.Boxes[i].GetID()], packResult.Boxes[i].GetID())
	}
}

func (s *PackerSuit) TestPacker_AllBoxes() {
	t := s.T()
	t.Parallel()

	packer := boxpacker3.NewPacker()

	boxes := NewDefaultBoxList()
	reverse := make([]*boxpacker3.Box, len(boxes))

	for i := range boxes {
		reverse[i] = boxes[len(boxes)-1-i]
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem(uuid.New().String(), 1000, 1000, 1000, 20000),
		boxpacker3.NewItem(uuid.New().String(), 2000, 500, 500, 20000),
		boxpacker3.NewItem(uuid.New().String(), 2000, 2000, 2000, 20000),
		boxpacker3.NewItem(uuid.New().String(), 2500, 2500, 2500, 20000),
		boxpacker3.NewItem(uuid.New().String(), 3000, 3000, 3000, 20000),

		boxpacker3.NewItem(uuid.New().String(), 220, 185, 50, 20000),
		boxpacker3.NewItem(uuid.New().String(), 165, 215, 100, 20000),
		boxpacker3.NewItem(uuid.New().String(), 265, 165, 190, 20000),
		boxpacker3.NewItem(uuid.New().String(), 425, 165, 190, 20000),
		boxpacker3.NewItem(uuid.New().String(), 425, 265, 190, 20000),
		boxpacker3.NewItem(uuid.New().String(), 425, 265, 380, 20000),
		boxpacker3.NewItem(uuid.New().String(), 530, 380, 265, 20000),
		boxpacker3.NewItem(uuid.New().String(), 1000, 500, 500, 20000),
	}

	packResult, err := packer.Pack(reverse, items)
	require.NoError(t, err)
	require.NotNil(t, packResult)

	require.Len(t, packResult.UnfitItems, 0)

	for i := 0; i < len(packResult.Boxes); i++ {
		require.Len(t, packResult.Boxes[i].GetItems(), 1, packResult.Boxes[i].GetID())
	}
}

func (s *PackerSuit) TestPacker_MinAndStd() {
	t := s.T()
	t.Parallel()

	packer := boxpacker3.NewPacker()

	boxes := NewDefaultBoxList()
	reverse := make([]*boxpacker3.Box, len(boxes))

	for i := range boxes {
		reverse[i] = boxes[len(boxes)-1-i]
	}

	items := []*boxpacker3.Item{
		// std
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690), // 1
		boxpacker3.NewItem(uuid.New().String(), 380, 100, 250, 2690), // 2
		boxpacker3.NewItem(uuid.New().String(), 250, 380, 100, 2690), // 3
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690), // 4
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690), // 5

		// min
		boxpacker3.NewItem(uuid.New().String(), 220, 185, 50, 20000),   // 6. F
		boxpacker3.NewItem(uuid.New().String(), 165, 215, 100, 20000),  // 7. E
		boxpacker3.NewItem(uuid.New().String(), 265, 165, 190, 20000),  // 8. G
		boxpacker3.NewItem(uuid.New().String(), 425, 165, 190, 20000),  // 9. C
		boxpacker3.NewItem(uuid.New().String(), 425, 265, 190, 20000),  // 10. B
		boxpacker3.NewItem(uuid.New().String(), 425, 265, 380, 20000),  // 11. A
		boxpacker3.NewItem(uuid.New().String(), 530, 380, 265, 20000),  // 12. Std
		boxpacker3.NewItem(uuid.New().String(), 1000, 500, 500, 20000), // 13. NotStd1

		// max
		boxpacker3.NewItem(uuid.New().String(), 3000, 3000, 3000, 20000), // 14. NotStd6
	}

	packResult, err := packer.Pack(reverse, items)
	require.NoError(t, err)
	require.NotNil(t, packResult)

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

	require.Len(t, packResult.UnfitItems, 0)

	for i := 0; i < len(packResult.Boxes); i++ {
		require.Len(t, packResult.Boxes[i].GetItems(), checks[packResult.Boxes[i].GetID()], packResult.Boxes[i].GetID())
	}
}
