package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/bavix/boxpacker3/v2"
)

const (
	BoxTypeF = "8ec81501-11a4-4b3f-9a52-7cd2f9c8370c"

	BoxTypeE = "9c69baf8-1ca3-46a0-9fc2-6f15ad9fef9a"

	BoxTypeG = "2c5279d3-48ad-451b-b673-f6d9be7fc6f6"

	BoxTypeC = "7f1cc68f-d554-4094-8734-c68df5c13154"

	BoxTypeB = "76cede41-86bb-4487-bfb0-9513f032d53e"

	BoxTypeA = "8e10cebf-cee6-4136-b060-1587b993d083"

	BoxTypeStd = "ba973206-aa64-493b-b37a-c53192cde8fd"

	BoxTypeNotStd1 = "cb1ed5b8-7405-48c5-bfd0-d86f75c99261"

	BoxTypeNotStd2 = "d91e2661-aebb-4a55-bfb5-4ff9c6e3c008"

	BoxTypeNotStd3 = "a0ecd730-375a-4313-bbe8-820710606b3d"

	BoxTypeNotStd4 = "6dff37f0-4dd1-4143-abdc-c19ab94f2e68"

	BoxTypeNotStd5 = "abac6d59-b51f-4d62-a338-42aca7afe1cc"

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

func NewDefaultBoxList() []*boxpacker3.Box {
	boxes := make([]*boxpacker3.Box, 0, len(defaultBoxes))

	for _, box := range defaultBoxes {
		boxes = append(boxes, boxpacker3.NewBox(box.Type, box.W, box.H, box.L, box.MaxWeight))
	}

	return boxes
}

type PackerSuit struct {
	suite.Suite
}

func TestBoxPackerSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(PackerSuit))
}

func (s *PackerSuit) TestEmptyBoxes() {
	t := s.T()

	packer := boxpacker3.NewPacker()

	packResult, err := packer.Pack(context.Background(), nil, nil)
	require.NoError(t, err)

	require.NotNil(t, packResult, "Pack function returned nil")
	require.Empty(t, packResult.Boxes, "PackResult.Boxes is not empty")
	require.Empty(t, packResult.Unpacked, "PackResult.Unpacked is not empty")
}

func (s *PackerSuit) TestEmptyItems() {
	t := s.T()

	packer := boxpacker3.NewPacker()

	boxes := NewDefaultBoxList()

	packResult, err := packer.Pack(context.Background(), boxes, nil)
	require.NoError(t, err)

	require.NotNil(t, packResult, "Pack function returned nil")

	require.Empty(t, packResult.Boxes, "no box should be used for an empty order")

	require.Empty(t, packResult.Unpacked, "PackResult.Unpacked is not empty")
}

func (s *PackerSuit) TestMinBox() {
	t := s.T()

	packer := boxpacker3.NewPacker()

	boxes := NewDefaultBoxList()

	item := boxpacker3.NewItem(
		uuid.New().String(),
		8,
		17,
		5,
		384)

	packResult, err := packer.Pack(context.Background(), boxes, []*boxpacker3.Item{item})
	require.NoError(t, err)

	require.NotNil(t, packResult, "Pack function returned nil")

	require.LessOrEqual(t, len(packResult.Boxes), len(boxes), "PackResult.Boxes has incorrect length")

	require.Empty(t, packResult.Unpacked, "PackResult.Unpacked is not empty")

	checks := map[string]int{
		BoxTypeF: 1,
	}

	for i := range packResult.Boxes {
		if len(packResult.Boxes[i].Items) > 0 {
			require.Len(t, packResult.Boxes[i].Items, checks[packResult.Boxes[i].ID()])
		}
	}
}

func (s *PackerSuit) TestRotate() {
	t := s.T()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)
	boxes := NewDefaultBoxList()

	items := []*boxpacker3.Item{
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 380, 100, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 250, 380, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
	}

	packResult, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, packResult)

	checks := map[string]int{
		BoxTypeStd: 5,
	}

	require.Empty(t, packResult.Unpacked)

	for i := range len(packResult.Boxes) {
		require.Len(t, packResult.Boxes[i].Items, checks[packResult.Boxes[i].ID()], packResult.Boxes[i].ID())
	}
}

func (s *PackerSuit) TestStd() {
	t := s.T()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)

	boxes := NewDefaultBoxList()

	items := []*boxpacker3.Item{
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
	}

	packResult, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)

	require.NotNil(t, packResult, "Pack function returned nil")

	require.LessOrEqual(t, len(packResult.Boxes), len(boxes), "PackResult.Boxes has incorrect length")

	require.Empty(t, packResult.Unpacked, "PackResult.Unpacked is not empty")

	checks := map[string]int{
		BoxTypeStd: 5,
	}

	for i := range packResult.Boxes {
		if len(packResult.Boxes[i].Items) > 0 {
			require.Len(t, packResult.Boxes[i].Items, checks[packResult.Boxes[i].ID()], packResult.Boxes[i].ID())
		}
	}
}

func (s *PackerSuit) TestBoxTypeF() {
	t := s.T()

	packer := boxpacker3.NewPacker()

	boxes := NewDefaultBoxList()

	items := []*boxpacker3.Item{
		boxpacker3.NewItem(uuid.New().String(), 100, 100, 5, 2500),
		boxpacker3.NewItem(uuid.New().String(), 100, 5, 100, 2500),
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2500),
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2500),
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2500),
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2500),

		boxpacker3.NewItem(uuid.New().String(), 35, 100, 100, 2500),
		boxpacker3.NewItem(uuid.New().String(), 35, 100, 100, 2500),
	}

	packResult, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)

	require.NotNil(t, packResult, "Pack function returned nil")

	require.LessOrEqual(t, len(packResult.Boxes), len(boxes), "PackResult.Boxes has incorrect length")

	require.Empty(t, packResult.Unpacked, "PackResult.Unpacked is not empty")

	checks := map[string]int{
		BoxTypeF: 8,
	}

	for i := range packResult.Boxes {
		if len(packResult.Boxes[i].Items) > 0 {
			require.Len(t, packResult.Boxes[i].Items, checks[packResult.Boxes[i].ID()], packResult.Boxes[i].ID())
		}
	}
}

func (s *PackerSuit) TestBoxTypeF_Weight() {
	t := s.T()

	packer := boxpacker3.NewPacker()
	boxes := NewDefaultBoxList()

	items := []*boxpacker3.Item{
		boxpacker3.NewItem(uuid.New().String(), 100, 100, 5, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 5, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 5, 100, 100, 2690),

		boxpacker3.NewItem(uuid.New().String(), 35, 100, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 35, 100, 100, 2690),
	}

	packResult, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)
	require.NotNil(t, packResult, "Pack function returned nil")

	require.LessOrEqual(t, len(packResult.Boxes), len(boxes), "PackResult.Boxes has incorrect length")

	require.Empty(t, packResult.Unpacked, "PackResult.Unpacked is not empty")

	packed := 0
	used := map[string]float64{}

	for _, box := range packResult.Boxes {
		count := len(box.Items)
		packed += count

		if count > 0 {
			used[box.ID()] = box.Stats.ItemsWeight
		}
	}

	require.Equal(t, len(items), packed)
	require.Len(t, used, 2)
	require.Contains(t, used, BoxTypeF)
	require.Contains(t, used, BoxTypeE)
	require.InDelta(t, used[BoxTypeF], used[BoxTypeE], 2690,
		"the weight balancing pass evens the two parcels out to within one item")
}

func (s *PackerSuit) TestPacker_AllBoxes() {
	t := s.T()

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

	packResult, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)

	require.NotNil(t, packResult, "Pack function returned nil")

	require.LessOrEqual(t, len(packResult.Boxes), len(boxes), "PackResult.Boxes has incorrect length")

	require.Empty(t, packResult.Unpacked, "PackResult.Unpacked is not empty")

	for i := range packResult.Boxes {
		require.Len(t, packResult.Boxes[i].Items, 1, packResult.Boxes[i].ID())
	}
}

func (s *PackerSuit) TestPacker_UnfitItems() {
	t := s.T()

	packer := boxpacker3.NewPacker()

	boxes := NewDefaultBoxList()

	items := []*boxpacker3.Item{
		boxpacker3.NewItem(uuid.New().String(), 3001, 3000, 3000, 20000),
		boxpacker3.NewItem(uuid.New().String(), 3000, 3001, 3000, 20000),
		boxpacker3.NewItem(uuid.New().String(), 3000, 3000, 3001, 20000),
		boxpacker3.NewItem(uuid.New().String(), 3000, 3000, 3000, 20001),
	}

	packResult, err := packer.Pack(context.Background(), boxes, items)
	require.NoError(t, err)

	require.NotNil(t, packResult, "Pack function returned nil")

	require.Len(t, packResult.Unpacked, 4, "PackResult.Unpacked has incorrect length")

	for i := range packResult.Boxes {
		require.Empty(t, packResult.Boxes[i].Items, packResult.Boxes[i].ID(), "Box "+packResult.Boxes[i].ID()+" contains items")
	}
}

func (s *PackerSuit) TestPacker_MinAndStd() {
	t := s.T()

	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)

	boxes := NewDefaultBoxList()
	reverse := make([]*boxpacker3.Box, len(boxes))

	for i := range boxes {
		reverse[i] = boxes[len(boxes)-1-i]
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 380, 100, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 250, 380, 100, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),
		boxpacker3.NewItem(uuid.New().String(), 100, 380, 250, 2690),

		boxpacker3.NewItem(uuid.New().String(), 220, 185, 50, 20000),
		boxpacker3.NewItem(uuid.New().String(), 165, 215, 100, 20000),
		boxpacker3.NewItem(uuid.New().String(), 265, 165, 190, 20000),
		boxpacker3.NewItem(uuid.New().String(), 425, 165, 190, 20000),
		boxpacker3.NewItem(uuid.New().String(), 425, 265, 190, 20000),
		boxpacker3.NewItem(uuid.New().String(), 425, 265, 380, 20000),
		boxpacker3.NewItem(uuid.New().String(), 530, 380, 265, 20000),
		boxpacker3.NewItem(uuid.New().String(), 1000, 500, 500, 20000),

		boxpacker3.NewItem(uuid.New().String(), 3000, 3000, 3000, 20000),
	}

	packResult, err := packer.Pack(context.Background(), reverse, items)
	require.NoError(t, err)

	require.NotNil(t, packResult, "Pack function returned nil")

	require.Empty(t, packResult.Unpacked)

	packed := 0
	used := 0

	for _, box := range packResult.Boxes {
		count := len(box.Items)
		packed += count

		if count > 0 {
			used++
		}
	}

	require.Equal(t, len(items), packed)
	require.LessOrEqual(t, used, 11)
}
