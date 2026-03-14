package boxpacker3_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

const groupKit = "kit"

func mustItem(t *testing.T, spec boxpacker3.ItemSpec) *boxpacker3.Item {
	t.Helper()

	item, err := boxpacker3.NewItemFromSpec(spec)
	require.NoError(t, err)

	return item
}

func mustBox(t *testing.T, spec boxpacker3.BoxSpec) *boxpacker3.Box {
	t.Helper()

	box, err := boxpacker3.NewBoxFromSpec(spec)
	require.NoError(t, err)

	return box
}

func boxOf(result *boxpacker3.Result, id string) *boxpacker3.PackedBox {
	for i := range result.Boxes {
		for _, item := range result.Boxes[i].Items {
			if item.ID() == id {
				return &result.Boxes[i]
			}
		}
	}

	return nil
}

func TestSupply_StockIsRespected(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{mustBox(t, boxpacker3.BoxSpec{
		ID:         boxCarton,
		OuterWidth: 100, OuterHeight: 100, OuterDepth: 100,
		MaxWeight: 100000,
		Quantity:  2,
	})}

	items := make([]*boxpacker3.Item, 0, 3)
	for i := range 3 {
		items = append(items, boxpacker3.NewItem(fmt.Sprintf("slab-%d", i), 100, 100, 60, 1))
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)

	used := usedBoxCountOf(result)

	require.Equal(t, 2, used, "only two cartons are available")
	require.Len(t, result.Unpacked, 1)
}

func TestSupply_DefaultsToOne(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{mustBox(t, boxpacker3.BoxSpec{
		ID:         boxCarton,
		OuterWidth: 100, OuterHeight: 100, OuterDepth: 100,
		MaxWeight: 100000,
	})}

	require.Equal(t, 1, boxes[0].Quantity())

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("a", 100, 100, 60, 1),
		boxpacker3.NewItem("b", 100, 100, 60, 1),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Len(t, result.Unpacked, 1)
}

func TestSupply_ShortConstructorIsOne(t *testing.T) {
	t.Parallel()

	require.Equal(t, 1, boxpacker3.NewBox("plain", 10, 10, 10, 100).Quantity())
}

func TestSupply_RejectsNegativeQuantity(t *testing.T) {
	t.Parallel()

	box, err := boxpacker3.NewBoxFromSpec(boxpacker3.BoxSpec{
		ID:         boxCarton,
		OuterWidth: 10, OuterHeight: 10, OuterDepth: 10,
		MaxWeight: 100,
		Quantity:  -1,
	})
	require.ErrorIs(t, err, boxpacker3.ErrInvalidQuantity)
	require.Nil(t, box)
}

func TestGroup_MembersTravelTogether(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small", 60, 60, 60, 100000),
		boxpacker3.NewBox("large", 200, 200, 200, 100000),
	}

	items := []*boxpacker3.Item{
		mustItem(t, boxpacker3.ItemSpec{ID: "product", Width: 50, Height: 50, Depth: 50, Weight: 1, Group: "order-42"}),
		mustItem(t, boxpacker3.ItemSpec{ID: "accessory", Width: 50, Height: 50, Depth: 50, Weight: 1, Group: "order-42"}),
		boxpacker3.NewItem("filler", 40, 40, 40, 1),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)

	product := boxOf(result, "product")
	accessory := boxOf(result, "accessory")

	require.NotNil(t, product)
	require.Same(t, product, accessory, "the group shares one box")
}

func TestGroup_UnpackableGroupIsReportedWhole(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("small", 60, 60, 60, 100000)}

	items := []*boxpacker3.Item{
		mustItem(t, boxpacker3.ItemSpec{ID: "a", Width: 50, Height: 50, Depth: 50, Weight: 1, Group: groupKit}),
		mustItem(t, boxpacker3.ItemSpec{ID: "b", Width: 50, Height: 50, Depth: 50, Weight: 1, Group: groupKit}),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)

	require.Len(t, result.Unpacked, 2, "neither member ships without the other")
}

func TestGroup_FreedSpaceGoesToOtherItems(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 60, 100000)}

	items := []*boxpacker3.Item{
		mustItem(t, boxpacker3.ItemSpec{ID: "kit-a", Width: 100, Height: 100, Depth: 40, Weight: 1, Group: groupKit}),
		mustItem(t, boxpacker3.ItemSpec{ID: "kit-b", Width: 100, Height: 100, Depth: 40, Weight: 1, Group: groupKit}),
		boxpacker3.NewItem("loose", 100, 100, 20, 1),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)

	require.NotNil(t, boxOf(result, "loose"), "the space the broken group gave up is reused")
	require.Len(t, result.Unpacked, 2)
}

func TestGroup_UngroupedItemsAreUnaffected(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 100000)}

	items := []*boxpacker3.Item{
		mustItem(t, boxpacker3.ItemSpec{ID: "empty-group", Width: 40, Height: 40, Depth: 40, Weight: 1, Group: ""}),
		boxpacker3.NewItem("plain", 40, 40, 40, 1),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)
	require.Empty(t, result.Unpacked)
}

func TestConstraint_RejectsPlacement(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 100000)}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("battery-1", 30, 30, 30, 1),
		boxpacker3.NewItem("battery-2", 30, 30, 30, 1),
		boxpacker3.NewItem("battery-3", 30, 30, 30, 1),
	}

	atMostTwo := boxpacker3.PlacementRuleFunc(
		func(box *boxpacker3.Container, _ *boxpacker3.Item, _ boxpacker3.Pivot, _ boxpacker3.Dimension) bool {
			return len(box.Contents()) < 2
		})

	result, err := boxpacker3.NewPacker(boxpacker3.WithRules(boxpacker3.Rules{Placement: []boxpacker3.PlacementRule{atMostTwo}})).
		Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	requireInvariants(t, items, result)

	require.Len(t, result.Boxes[0].Items, 2)
	require.Len(t, result.Unpacked, 1)
}

func TestConstraint_SeesTheProposedPlacement(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 100000)}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("a", 40, 40, 40, 1),
		boxpacker3.NewItem("b", 40, 40, 40, 1),
	}

	groundFloorOnly := boxpacker3.PlacementRuleFunc(
		func(_ *boxpacker3.Container, _ *boxpacker3.Item, position boxpacker3.Pivot, _ boxpacker3.Dimension) bool {
			return position[2] == 0
		})

	result, err := boxpacker3.NewPacker(boxpacker3.WithRules(boxpacker3.Rules{Placement: []boxpacker3.PlacementRule{groundFloorOnly}})).
		Pack(t.Context(), boxes, items)
	require.NoError(t, err)

	for _, item := range result.Boxes[0].Items {
		require.InDelta(t, 0.0, item.Position[2], 1e-9)
	}
}

func TestConstraint_NoneConfiguredIsNotConsulted(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 100000)}
	items := []*boxpacker3.Item{boxpacker3.NewItem("a", 40, 40, 40, 1)}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Empty(t, result.Unpacked)
}

func TestConstraint_AllMustAgree(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 100000)}
	items := []*boxpacker3.Item{boxpacker3.NewItem("a", 40, 40, 40, 1)}

	permit := boxpacker3.PlacementRuleFunc(
		func(*boxpacker3.Container, *boxpacker3.Item, boxpacker3.Pivot, boxpacker3.Dimension) bool {
			return true
		})
	refuse := boxpacker3.PlacementRuleFunc(
		func(*boxpacker3.Container, *boxpacker3.Item, boxpacker3.Pivot, boxpacker3.Dimension) bool {
			return false
		})

	result, err := boxpacker3.NewPacker(boxpacker3.WithRules(boxpacker3.Rules{Placement: []boxpacker3.PlacementRule{permit, refuse}})).
		Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Len(t, result.Unpacked, 1)
}
