package boxpacker3_test

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

const (
	nameSearchFiller = "Search as filler"
	settingPlain     = "plain"
)

func everyWayToPack(t *testing.T) map[string]boxpacker3.Algorithm {
	t.Helper()

	algorithms := map[string]boxpacker3.Algorithm{}

	for _, rule := range boxpacker3.EveryRuleSettings() {
		algorithms[rule.Name()] = boxpacker3.NewGreedy(rule.Order, rule.Selection)
	}

	for _, goal := range boxpacker3.Goals() {
		answer, ok := boxpacker3.BestForName(goal.Name())
		require.True(t, ok)

		algorithms["BestFor("+goal.Name()+")"] = answer
	}

	search, err := boxpacker3.NewSearch(64, 3)
	require.NoError(t, err)

	algorithms[search.Name()] = search
	algorithms["SingleContainer"] = boxpacker3.SingleContainer(nil)
	algorithms[nameSearchFiller] = boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectFullestBox)

	return algorithms
}

func everySetting() map[string]boxpacker3.Rules {
	settings := map[string]boxpacker3.Rules{
		settingPlain:    {},
		"support 1":     {MinSupportRatio: 1},
		"support 0.7":   {MinSupportRatio: 0.7},
		"space corners": {FreeSpaceCorners: true},
	}

	for _, merit := range boxpacker3.Merits() {
		settings["merit "+merit.Name()] = boxpacker3.Rules{Merit: merit}
	}

	return settings
}

func auditOrderV2(t *testing.T, seed int) ([]*boxpacker3.Box, []*boxpacker3.Item) {
	t.Helper()

	random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

	boxes := make([]*boxpacker3.Box, 0, 5)

	for i := range 3 + random.Intn(3) {
		var accepts []string
		if random.Intn(4) == 0 {
			accepts = []string{classFood}
		}

		boxes = append(boxes, mustBox(t, boxpacker3.BoxSpec{
			ID:          fmt.Sprintf("box-%d", i),
			OuterWidth:  float64(240 + random.Intn(300)),
			OuterHeight: float64(240 + random.Intn(300)),
			OuterDepth:  float64(240 + random.Intn(300)),
			EmptyWeight: float64(random.Intn(400)),
			MaxWeight:   float64(6000 + random.Intn(14000)),
			Quantity:    1 + random.Intn(3),
			Accepts:     accepts,
		}))
	}

	rotations := boxpacker3.Rotations()
	classes := []string{"", classFood, classChemicals, "glass"}

	items := make([]*boxpacker3.Item, 0, 24)

	for i := range 10 + random.Intn(14) {
		class := classes[random.Intn(len(classes))]

		var separate []string
		if class == classChemicals {
			separate = []string{classFood}
		}

		group := ""
		if random.Intn(5) == 0 {
			group = fmt.Sprintf("group-%d", random.Intn(3))
		}

		items = append(items, mustItem(t, boxpacker3.ItemSpec{
			ID:           fmt.Sprintf("item-%d", i),
			Width:        float64(40 + random.Intn(160)),
			Height:       float64(40 + random.Intn(160)),
			Depth:        float64(40 + random.Intn(160)),
			Weight:       float64(50 + random.Intn(900)),
			Rotation:     rotations[random.Intn(len(rotations))],
			Group:        group,
			Quantity:     1 + random.Intn(3),
			MaxLoadOnTop: float64(random.Intn(2) * 500),
			NothingOnTop: random.Intn(8) == 0,
			Class:        class,
			SeparateFrom: separate,
		}))
	}

	return boxes, items
}

func TestAuditV2_EveryAlgorithmUnderEverySetting(t *testing.T) {
	t.Parallel()

	settings := everySetting()

	for name, algorithm := range everyWayToPack(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for settingName, rules := range settings {
				for seed := range 12 {
					boxes, items := auditOrderV2(t, seed)

					options := []boxpacker3.Option{
						boxpacker3.WithAlgorithm(algorithm),
						boxpacker3.WithRules(rules),
					}

					if name == nameSearchFiller {
						search, err := boxpacker3.NewSearch(32, 3)
						require.NoError(t, err)

						options = append(options, boxpacker3.WithFiller(search))
					}

					result, err := boxpacker3.NewPacker(options...).Pack(t.Context(), boxes, items)
					require.NoError(t, err, "%s/%s seed %d", name, settingName, seed)

					where := fmt.Sprintf("%s/%s seed %d", name, settingName, seed)
					auditResultV2(t, where, boxes, items, result, rules)
				}
			}
		})
	}
}

func auditResultV2(
	t *testing.T, where string,
	boxes []*boxpacker3.Box, items []*boxpacker3.Item,
	result *boxpacker3.Result, rules boxpacker3.Rules,
) {
	t.Helper()

	requireInvariants(t, items, result)
	auditStock(t, where, boxes, result)
	auditPlacements(t, where, result, rules)
	auditClasses(t, where, result)
	auditGroupsV2(t, where, result)
	auditReasons(t, where, boxes, result)
	auditReport(t, where, result)
}

func auditStock(t *testing.T, where string, boxes []*boxpacker3.Box, result *boxpacker3.Result) {
	t.Helper()

	stock := map[string]int{}
	for _, box := range boxes {
		stock[box.ID()] = box.Quantity()
	}

	used := map[string]int{}
	indexes := map[string]map[int]bool{}

	for _, box := range result.Boxes {
		used[box.ID()]++

		if indexes[box.ID()] == nil {
			indexes[box.ID()] = map[int]bool{}
		}

		require.False(t, indexes[box.ID()][box.Index],
			"%s: box %s index %d appears twice", where, box.ID(), box.Index)
		indexes[box.ID()][box.Index] = true

		require.Less(t, box.Index, stock[box.ID()],
			"%s: box %s index %d is beyond its stock of %d", where, box.ID(), box.Index, stock[box.ID()])
	}

	for id, count := range used {
		require.LessOrEqual(t, count, stock[id], "%s: %d of box %s, stock is %d", where, count, id, stock[id])
	}
}

func auditPlacements(t *testing.T, where string, result *boxpacker3.Result, rules boxpacker3.Rules) {
	t.Helper()

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			require.Contains(t, item.Item.Orientations(), item.Dimension,
				"%s: %s lies in an orientation it does not allow", where, item.ID())

			if rules.MinSupportRatio > 0 {
				require.GreaterOrEqual(t, supportRatioOf(box, item), rules.MinSupportRatio-1e-9,
					"%s: %s rests on too little", where, item.ID())
			}

			if item.Item.NothingOnTop() {
				require.Zero(t, loadOnTopOf(box, item),
					"%s: %s carries a load although it must stay clear", where, item.ID())
			}

			if limit := item.Item.MaxLoadOnTop(); limit > 0 {
				require.LessOrEqual(t, loadOnTopOf(box, item), limit+1e-9,
					"%s: %s carries more than its cap", where, item.ID())
			}
		}
	}
}

func auditClasses(t *testing.T, where string, result *boxpacker3.Result) {
	t.Helper()

	for _, box := range result.Boxes {
		accepts := box.Box.Accepts()

		for i, one := range box.Items {
			if len(accepts) > 0 {
				require.Contains(t, accepts, one.Item.Class(),
					"%s: box %s does not take %s", where, box.ID(), one.ID())
			}

			for _, other := range box.Items[i+1:] {
				require.NotContains(t, one.Item.SeparateFrom(), other.Item.Class(),
					"%s: %s shares a box with %s", where, one.ID(), other.ID())
				require.NotContains(t, other.Item.SeparateFrom(), one.Item.Class(),
					"%s: %s shares a box with %s", where, other.ID(), one.ID())
			}
		}
	}
}

func auditGroupsV2(t *testing.T, where string, result *boxpacker3.Result) {
	t.Helper()

	holders := map[string]map[string]bool{}
	unpacked := map[string]bool{}

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			if group := item.Item.Group(); group != "" {
				if holders[group] == nil {
					holders[group] = map[string]bool{}
				}

				holders[group][fmt.Sprintf("%s#%d", box.ID(), box.Index)] = true
			}
		}
	}

	for _, item := range result.Unpacked {
		if group := item.Item.Group(); group != "" {
			unpacked[group] = true
		}
	}

	for group, boxes := range holders {
		require.Len(t, boxes, 1, "%s: group %s is spread over %d boxes", where, group, len(boxes))
		require.False(t, unpacked[group], "%s: group %s is part packed and part left behind", where, group)
	}
}

func auditReasons(t *testing.T, where string, boxes []*boxpacker3.Box, result *boxpacker3.Result) {
	t.Helper()

	for _, left := range result.Unpacked {
		holdable := false
		for _, box := range boxes {
			holdable = holdable || couldHold(box, left.Item)
		}

		if left.Reason.Structural() {
			require.False(t, holdable, "%s: %s says %s but a box could hold it", where, left.ID(), left.Reason)
		}

		if left.Reason == boxpacker3.ReasonNoRoom {
			require.True(t, holdable, "%s: %s fits no box but says only no room", where, left.ID())
		}

		require.NotEmpty(t, left.Reason.String())
	}
}

func auditReport(t *testing.T, where string, result *boxpacker3.Result) {
	t.Helper()

	require.NotEmpty(t, result.Report.Algorithm, "%s: the report does not say what packed", where)
	require.Equal(t, max(len(result.Boxes)-result.Report.Bound.Boxes, 0), result.Report.GapBoxes, where)

	packed := 0.0
	for _, box := range result.Boxes {
		packed += box.Stats.ItemsVolume
	}

	require.LessOrEqual(t, packed, result.Report.Bound.Volume+1e-6,
		"%s: more volume packed than the goods come to", where)

	if len(result.Unpacked) == 0 && result.Report.Bound.Boxes > 0 {
		require.GreaterOrEqual(t, len(result.Boxes), result.Report.Bound.Boxes,
			"%s: fewer boxes than the bound allows", where)
	}

	for _, box := range result.Boxes {
		require.InDelta(t, box.Stats.GrossWeight, box.Box.EmptyWeight()+box.Stats.ItemsWeight, 1e-6, where)
		require.LessOrEqual(t, box.Stats.GrossWeight, box.Box.MaxWeight()+1e-6,
			"%s: box %s is over its gross limit", where, box.ID())
		require.InDelta(t, box.Stats.Fill, box.Stats.ItemsVolume/box.Volume(), 1e-9, where)
	}
}

func TestAuditV2_PackerIsStateless(t *testing.T) {
	t.Parallel()

	for name, algorithm := range everyWayToPack(t) {
		if name == nameSearchFiller {
			continue
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			packer := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(algorithm))

			first, firstItems := auditOrderV2(t, 3)
			other, otherItems := auditOrderV2(t, 4)

			before := fingerprint(packed(t, packer, first, firstItems))

			packed(t, packer, other, otherItems)

			after := fingerprint(packed(t, packer, first, firstItems))
			require.Equal(t, before, after, "the run in between changed the answer")

			fresh := fingerprint(packed(t, boxpacker3.NewPacker(boxpacker3.WithAlgorithm(algorithm)), first, firstItems))
			require.Equal(t, before, fresh, "a fresh packer answers differently")
		})
	}
}

func TestAuditV2_EveryAlgorithmHonoursABudget(t *testing.T) {
	t.Parallel()

	boxes, items := auditOrderV2(t, 9)

	for name, algorithm := range everyWayToPack(t) {
		if name == nameSearchFiller {
			continue
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			started := time.Now()

			_, err := boxpacker3.NewPacker(
				boxpacker3.WithAlgorithm(algorithm),
				boxpacker3.WithBudget(boxpacker3.Budget{Timeout: 50 * time.Millisecond}),
			).Pack(t.Context(), boxes, items)

			elapsed := time.Since(started)

			if err != nil {
				require.ErrorIs(t, err, context.DeadlineExceeded)
			}

			require.Less(t, elapsed, 10*time.Second, "%s ran far past its budget", name)
		})
	}
}

func TestAuditV2_DegenerateInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		boxes []*boxpacker3.Box
		items []*boxpacker3.Item
	}{
		{"nothing at all", nil, nil},
		{"boxes, no goods", []*boxpacker3.Box{boxpacker3.NewBox("a", 10, 10, 10, 100)}, nil},
		{"goods, no boxes", nil, []*boxpacker3.Item{boxpacker3.NewItem("a", 1, 1, 1, 1)}},
		{
			"an item filling the box exactly",
			[]*boxpacker3.Box{boxpacker3.NewBox("a", 10, 10, 10, 100)},
			[]*boxpacker3.Item{boxpacker3.NewItem("a", 10, 10, 10, 1)},
		},
		{
			"a box far longer than it is wide",
			[]*boxpacker3.Box{boxpacker3.NewBox("a", 10000, 10, 10, 1e9)},
			[]*boxpacker3.Item{boxpacker3.NewItem("a", 10, 10, 10, 1), boxpacker3.NewItem("b", 10, 10, 10, 1)},
		},
		{
			"weightless goods in a weightless box",
			[]*boxpacker3.Box{boxpacker3.NewBox("a", 10, 10, 10, 0)},
			[]*boxpacker3.Item{boxpacker3.NewItem("a", 1, 1, 1, 0)},
		},
	}

	for name, algorithm := range everyWayToPack(t) {
		if name == nameSearchFiller {
			continue
		}

		for _, test := range cases {
			result, err := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(algorithm)).
				Pack(t.Context(), test.boxes, test.items)
			require.NoError(t, err, "%s: %s", name, test.name)
			require.NotNil(t, result, "%s: %s", name, test.name)

			requireInvariants(t, test.items, result)
		}
	}
}

func TestAuditV2_ScaleInvariance(t *testing.T) {
	t.Parallel()

	build := func(scale float64) ([]*boxpacker3.Box, []*boxpacker3.Item) {
		boxes := []*boxpacker3.Box{boxpacker3.NewBox("crate", 12*scale, 8*scale, 6*scale, 1e9)}

		items := make([]*boxpacker3.Item, 0, 12)
		for i := range 12 {
			items = append(items, boxpacker3.NewItem(
				fmt.Sprintf("cube-%d", i), 4*scale, 4*scale, 3*scale, 1))
		}

		return boxes, items
	}

	for name, algorithm := range everyWayToPack(t) {
		if name == nameSearchFiller {
			continue
		}

		counts := map[float64]int{}

		for _, scale := range []float64{0.001, 1, 1000} {
			boxes, items := build(scale)

			result, err := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(algorithm)).
				Pack(t.Context(), boxes, items)
			require.NoError(t, err)

			counts[scale] = len(result.Boxes[0].Items)
		}

		require.Equal(t, counts[1], counts[0.001], "%s packs differently at a thousandth of the scale", name)
		require.Equal(t, counts[1], counts[1000], "%s packs differently at a thousand times the scale", name)
	}
}

func TestAuditV2_GeometryUnderEveryMerit(t *testing.T) {
	t.Parallel()

	for _, merit := range boxpacker3.Merits() {
		for _, corners := range []bool{false, true} {
			for seed := range 25 {
				boxes, items := auditOrderV2(t, seed+100)

				result, err := boxpacker3.NewPacker(
					boxpacker3.WithRules(boxpacker3.Rules{Merit: merit, FreeSpaceCorners: corners}),
				).Pack(t.Context(), boxes, items)
				require.NoError(t, err)

				where := fmt.Sprintf("merit %s corners %v seed %d", merit.Name(), corners, seed)

				requireInvariants(t, items, result)

				for _, box := range result.Boxes {
					volume := 0.0
					for _, item := range box.Items {
						volume += item.Volume()

						require.InEpsilon(t, item.Dimension[0]*item.Dimension[1]*item.Dimension[2], item.Volume(), 1e-9,
							"%s: %s reports a volume its dimensions do not give", where, item.ID())
					}

					require.InDelta(t, volume, box.Stats.ItemsVolume, 1e-6, where)
					require.InDelta(t, box.Volume()-volume, box.RemainingVolume(), 1e-6, where)
				}
			}
		}
	}
}

func TestAuditV2_ReportedNumbersAddUp(t *testing.T) {
	t.Parallel()

	for seed := range 50 {
		boxes, items := auditOrderV2(t, seed+200)

		result, err := boxpacker3.NewPacker(
			boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes)),
		).Pack(t.Context(), boxes, items)
		require.NoError(t, err)

		metrics := boxpacker3.MetricsOf(result)
		require.Equal(t, len(result.Boxes), metrics.Boxes)
		require.Equal(t, len(result.Unpacked), metrics.Unpacked)

		volume := 0.0
		for _, box := range result.Boxes {
			volume += box.Volume()
		}

		require.InDelta(t, volume, metrics.BoxVolume, 1e-6)
		require.False(t, math.IsNaN(metrics.Fill))
		require.GreaterOrEqual(t, metrics.Fill, 0.0)
		require.LessOrEqual(t, metrics.Fill, 1.0)
	}
}
