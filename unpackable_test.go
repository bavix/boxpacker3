package boxpacker3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func reasonsOf(result *boxpacker3.Result) map[string]boxpacker3.Reason {
	reasons := map[string]boxpacker3.Reason{}

	for _, item := range result.Unpacked {
		reasons[item.ID()] = item.Reason
	}

	return reasons
}

func TestUnpackable_OversizedItemIsReported(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("small", 10, 10, 10, 100)}
	items := []*boxpacker3.Item{boxpacker3.NewItem("oversized", 50, 50, 50, 1)}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Len(t, result.Unpacked, 1)
	require.Equal(t, boxpacker3.ReasonTooBig, result.Unpacked[0].Reason)
	require.Len(t, result.Unpackable(), 1)
	require.Equal(t, "oversized", result.Unpackable()[0].ID())
}

func TestUnpackable_TooHeavyItemIsReported(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 10)}
	items := []*boxpacker3.Item{boxpacker3.NewItem("anvil", 10, 10, 10, 1000)}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Equal(t, boxpacker3.ReasonTooHeavy, reasonsOf(result)["anvil"])
}

func TestUnpackable_TareCountsAgainstTheItem(t *testing.T) {
	t.Parallel()

	box := mustBox(t, boxpacker3.BoxSpec{
		ID:         boxCarton,
		OuterWidth: 100, OuterHeight: 100, OuterDepth: 100,
		EmptyWeight: 900,
		MaxWeight:   1000,
	})

	items := []*boxpacker3.Item{boxpacker3.NewItem("heavy", 10, 10, 10, 200)}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)
	require.Equal(t, boxpacker3.ReasonTooHeavy, reasonsOf(result)["heavy"],
		"only 100 of the 1000 gross limit is left for contents")
}

func TestUnpackable_ItemThatMerelyDidNotFitIsNoRoom(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 100000)}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("a", 100, 100, 60, 1),
		boxpacker3.NewItem("b", 100, 100, 60, 1),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Len(t, result.Unpacked, 1)
	require.Equal(t, boxpacker3.ReasonNoRoom, result.Unpacked[0].Reason,
		"the second item fits the box, the box just filled up")
	require.Empty(t, result.Unpackable())
}

func TestUnpackable_EveryAffectedItemIsCarried(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("small", 10, 10, 10, 100)}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("wide", 50, 5, 5, 1),
		boxpacker3.NewItem("tall", 5, 5, 50, 1),
		boxpacker3.NewItem("fits", 5, 5, 5, 1),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Len(t, result.Unpackable(), 2, "both oversized items are named, the one that fits is not")
}

func TestUnpackable_NothingLeftOverMeansNothingUnpackable(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("box", 100, 100, 100, 100000)}
	items := []*boxpacker3.Item{boxpacker3.NewItem("a", 10, 10, 10, 1)}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Empty(t, result.Unpacked)
	require.Empty(t, result.Unpackable())
}

func TestUnpackable_RotationRestrictionIsItsOwnReason(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("slot", 100, 100, 10, 100000)}

	items := []*boxpacker3.Item{mustItem(t, boxpacker3.ItemSpec{
		ID: "upright", Width: 10, Height: 10, Depth: 90, Weight: 1,
		Rotation: boxpacker3.RotationNever,
	})}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Equal(t, boxpacker3.ReasonNoOrientationFits, reasonsOf(result)["upright"],
		"the only orientation it permits does not fit, another would")
}

func TestUnpackable_DedicatedBoxRefusesUnlabelledGoods(t *testing.T) {
	t.Parallel()

	box := mustBox(t, boxpacker3.BoxSpec{
		ID: "food-only", OuterWidth: 100, OuterHeight: 100, OuterDepth: 100, MaxWeight: 1e6,
		Accepts: []string{"food"},
	})

	items := []*boxpacker3.Item{boxpacker3.NewItem("bolt", 10, 10, 10, 1)}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), []*boxpacker3.Box{box}, items)
	require.NoError(t, err)
	require.Equal(t, boxpacker3.ReasonRefusedByBox, reasonsOf(result)["bolt"])
}

func TestUnpackable_ReasonsHaveNames(t *testing.T) {
	t.Parallel()

	seen := map[string]bool{}

	for _, reason := range boxpacker3.Reasons() {
		require.NotEmpty(t, reason.String())
		require.False(t, seen[reason.String()], "two reasons read %q", reason.String())

		seen[reason.String()] = true
	}
}
