package boxpacker3_test

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

func groupSplits(t *testing.T, result *boxpacker3.Result) []string {
	t.Helper()

	where := map[string]map[string]int{}

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			group := item.Item.Group()
			if group == "" {
				continue
			}

			if where[group] == nil {
				where[group] = map[string]int{}
			}

			where[group][box.ID()]++
		}
	}

	for _, item := range result.Unpacked {
		group := item.Item.Group()
		if group == "" {
			continue
		}

		if where[group] == nil {
			where[group] = map[string]int{}
		}

		where[group]["<unfit>"]++
	}

	broken := make([]string, 0)

	for group, spread := range where {
		if len(spread) > 1 {
			broken = append(broken, fmt.Sprintf("%s spread over %v", group, spread))
		}
	}

	return broken
}

func TestAudit_GroupsShipTogether(t *testing.T) {
	t.Parallel()

	failures := 0

	for seed := range 400 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 2+random.Intn(3), 100, 160, 100000)
		items := groupedOrder(t, random, 6+random.Intn(14))

		for _, rule := range boxpacker3.BoxSelections() {
			for _, balance := range []int{0, 4} {
				failures += checkGroupsHeld(t, boxes, items, rule, balance, seed, failures)
			}
		}
	}

	t.Logf("group splits: %d", failures)
}

func checkGroupsHeld(t *testing.T,
	boxes []*boxpacker3.Box, items []*boxpacker3.Item,
	rule boxpacker3.BoxSelection, balance, seed, sofar int,
) int {
	t.Helper()

	options := []boxpacker3.Option{boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, rule))}
	if balance > 0 {
		options = append(options, boxpacker3.WithFinishers(boxpacker3.BalanceWeight{MaxBoxes: balance}))
	}

	result, err := boxpacker3.NewPacker(options...).Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	broken := groupSplits(t, result)
	if len(broken) > 0 && sofar < 5 {
		t.Errorf("seed %d rule %s balance %d: %v", seed, rule, balance, broken)
	}

	return len(broken)
}

type violation struct {
	kind string
	note string
}

func auditResult(result *boxpacker3.Result, sent []*boxpacker3.Item, support float64) []violation {
	found := make([]violation, 0)

	for _, box := range result.Boxes {
		found = append(found, checkPlacements(box, support)...)
		found = append(found, checkWeight(box)...)
	}

	return append(found, checkEveryItemCameBack(result, sent)...)
}

func checkWeight(box boxpacker3.PackedBox) []violation {
	weight := 0.0

	for _, item := range box.Items {
		weight += item.Weight()
	}

	if weight <= box.Box.MaxWeight()+1e-9 {
		return nil
	}

	return []violation{{
		kind: "overweight",
		note: fmt.Sprintf("%s carries %.0f of %.0f", box.ID(), weight, box.Box.MaxWeight()),
	}}
}

func checkEveryItemCameBack(result *boxpacker3.Result, sent []*boxpacker3.Item) []violation {
	found := make([]violation, 0)
	seen := map[string]int{}

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			seen[item.ID()]++
		}
	}

	for _, item := range result.Unpacked {
		seen[item.ID()]++
	}

	for _, item := range sent {
		switch seen[item.ID()] {
		case 1:
		case 0:
			found = append(found, violation{kind: "lost", note: item.ID()})
		default:
			found = append(found, violation{kind: "duplicated", note: item.ID()})
		}
	}

	return found
}

func overlapping(a, b boxpacker3.PackedItem) bool {
	pa, da := a.Position, a.Dimension
	pb, db := b.Position, b.Dimension

	for axis := range 3 {
		if pa[axis]+da[axis] <= pb[axis]+1e-9 || pb[axis]+db[axis] <= pa[axis]+1e-9 {
			return false
		}
	}

	return true
}

func fingerprint(result *boxpacker3.Result) string {
	var out strings.Builder

	for _, box := range result.Boxes {
		out.WriteString(box.ID())
		out.WriteString("{")

		for _, item := range box.Items {
			position := item.Position
			dimension := item.Dimension
			fmt.Fprintf(&out, "%s@%.0f,%.0f,%.0f:%.0fx%.0fx%.0f;",
				item.ID(), position[0], position[1], position[2],
				dimension[0], dimension[1], dimension[2])
		}

		out.WriteString("}")
	}

	out.WriteString("|unfit:")

	for _, item := range result.Unpacked {
		out.WriteString(item.ID())
		out.WriteString(",")
	}

	return out.String()
}

func randomOrder(seed int64) ([]*boxpacker3.Box, []*boxpacker3.Item) {
	random := rand.New(rand.NewSource(seed)) //nolint:gosec

	return shelfOf(random, 2+random.Intn(3), 100, 170, 100000),
		cargoOf(random, 6+random.Intn(16), 20, 70, 100, 400)
}

func supportUnder(box boxpacker3.PackedBox, item boxpacker3.PackedItem) float64 {
	position, dimension := item.Position, item.Dimension
	base := dimension[0] * dimension[1]

	if base <= 0 || position[2] <= 1e-9 {
		return 1
	}

	carried := 0.0

	for _, other := range box.Items {
		if other.Item == item.Item && other.Index == item.Index {
			continue
		}

		op, od := other.Position, other.Dimension
		if math.Abs(op[2]+od[2]-position[2]) > 1e-6 {
			continue
		}

		carried += overlapSpan(position[0], dimension[0], op[0], od[0]) *
			overlapSpan(position[1], dimension[1], op[1], od[1])
	}

	return math.Min(carried/base, 1)
}

func overlapSpan(start, span, otherStart, otherSpan float64) float64 {
	return math.Max(0, math.Min(start+span, otherStart+otherSpan)-math.Max(start, otherStart))
}

type floorOnly struct{}

func (floorOnly) Allow(_ *boxpacker3.Container, _ *boxpacker3.Item,
	position boxpacker3.Pivot, _ boxpacker3.Dimension,
) bool {
	return position[2] == 0
}

func taredShelf(t *testing.T, random *rand.Rand, count int) ([]*boxpacker3.Box, map[string]int) {
	t.Helper()

	stock := make(map[string]int, count)
	boxes := make([]*boxpacker3.Box, 0, count)

	for i := range count {
		id := fmt.Sprintf("box-%d", i)
		side := 110 + float64(random.Intn(140))
		quantity := 1 + random.Intn(3)
		stock[id] = quantity

		built, err := boxpacker3.NewBoxFromSpec(boxpacker3.BoxSpec{
			ID:          id,
			OuterWidth:  side + 8,
			OuterHeight: side*0.8 + 8,
			OuterDepth:  side*0.7 + 8,
			InnerWidth:  side,
			InnerHeight: side * 0.8,
			InnerDepth:  side * 0.7,
			EmptyWeight: float64(100 + random.Intn(400)),
			MaxWeight:   float64(1500 + random.Intn(3000)),
			Quantity:    quantity,
		})
		if err != nil {
			t.Fatal(err)
		}

		boxes = append(boxes, built)
	}

	return boxes, stock
}

func groupedOrder(t *testing.T, random *rand.Rand, count int) []*boxpacker3.Item {
	t.Helper()

	items := make([]*boxpacker3.Item, 0, count)

	for i := range count {
		group := ""
		if random.Intn(3) > 0 {
			group = fmt.Sprintf("g%d", random.Intn(3))
		}

		built, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
			ID:     fmt.Sprintf("item-%d", i),
			Width:  20 + float64(random.Intn(70)),
			Height: 20 + float64(random.Intn(70)),
			Depth:  20 + float64(random.Intn(70)),
			Weight: float64(120 + random.Intn(600)),
			Group:  group,
		})
		if err != nil {
			t.Fatal(err)
		}

		items = append(items, built)
	}

	return items
}

func checkStockTareAndFloor(
	result *boxpacker3.Result,
	stock map[string]int,
	floorOnlyRun bool,
	report func(kind, note string),
) {
	used := map[string]int{}

	for _, box := range result.Boxes {
		if len(box.Items) == 0 {
			continue
		}

		kind, _, _ := strings.Cut(box.ID(), "#")
		used[kind]++

		if gross := box.Stats.GrossWeight; gross > box.Box.MaxWeight()+1e-9 {
			report("overweight", fmt.Sprintf("%s gross %.0f of %.0f",
				box.ID(), gross, box.Box.MaxWeight()))
		}

		if !floorOnlyRun {
			continue
		}

		for _, item := range box.Items {
			if item.Position[2] != 0 {
				report("constraint", fmt.Sprintf("%s stacked at %.0f",
					item.ID(), item.Position[2]))
			}
		}
	}

	for kind, count := range used {
		if count > stock[kind] {
			report("overstock", fmt.Sprintf("%s used %d of %d", kind, count, stock[kind]))
		}
	}
}

func TestAudit_StockTareAndConstraints(t *testing.T) {
	t.Parallel()

	tally := map[string]int{}

	report := func(kind, note string) {
		tally[kind]++

		if tally[kind] <= 3 {
			t.Errorf("%s — %s", kind, note)
		}
	}

	for seed := range 500 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes, stock := taredShelf(t, random, 2+random.Intn(3))
		items := groupedOrder(t, random, 5+random.Intn(14))
		floorOnlyRun := seed%2 == 0

		for _, rule := range boxpacker3.BoxSelections() {
			options := []boxpacker3.Option{
				boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, rule)),
				boxpacker3.WithFinishers(boxpacker3.BalanceWeight{MaxBoxes: 5}),
			}

			if floorOnlyRun {
				options = append(options, boxpacker3.WithRules(boxpacker3.Rules{Placement: []boxpacker3.PlacementRule{floorOnly{}}}))
			}

			result, err := boxpacker3.NewPacker(options...).Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			checkStockTareAndFloor(result, stock, floorOnlyRun, report)
		}
	}

	t.Logf("violations: %v", tally)
}

func TestAudit_Determinism(t *testing.T) {
	t.Parallel()

	rules := boxpacker3.BoxSelections()
	drifted := 0

	for seed := range 120 {
		boxes, items := randomOrder(int64(seed))

		for _, rule := range rules {
			packer := rulePacker(boxpacker3.OrderDecreasing, rule)

			first := ""

			for repeat := range 6 {
				result, err := packer.Pack(t.Context(), boxes, items)
				if err != nil {
					t.Fatal(err)
				}

				mark := fingerprint(result)
				if repeat == 0 {
					first = mark

					continue
				}

				if mark != first {
					drifted++
					if drifted <= 3 {
						t.Errorf("seed %d rule %s drifted on repeat %d", seed, rule, repeat)
					}
				}
			}
		}
	}

	t.Logf("drifts: %d", drifted)
}

func TestAudit_SharedPackerUnderLoad(t *testing.T) {
	t.Parallel()

	boxes, items := randomOrder(7)
	packer := rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectFullestBox)

	expected := fingerprint(packed(t, packer, boxes, items))

	marks := make([]string, 16)

	var group sync.WaitGroup

	for i := range marks {
		group.Go(func() {
			marks[i] = fingerprint(packed(t, packer, boxes, items))
		})
	}

	group.Wait()

	for i, mark := range marks {
		if mark != expected {
			t.Errorf("worker %d produced a different packing", i)
		}
	}
}

func TestAudit_UnpackableIsToldFromNoRoom(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("small", 100, 100, 100, 5000)}

	fits := boxpacker3.NewItem("fits", 50, 50, 50, 100)
	tooBig := boxpacker3.NewItem("too-big", 500, 50, 50, 100)
	tooHeavy := boxpacker3.NewItem("too-heavy", 50, 50, 50, 90000)

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, []*boxpacker3.Item{fits, tooBig, tooHeavy})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Unpackable()) != 2 {
		t.Fatalf("two items nothing can hold, reported %v", result.Unpacked)
	}

	result, err = boxpacker3.NewPacker().Pack(t.Context(), boxes, []*boxpacker3.Item{fits})
	if err != nil || len(result.Unpackable()) != 0 {
		t.Errorf("an order that fits was rejected: %v %v", result.Unpacked, err)
	}
}

func TestAudit_PerfectTilingShortfall(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("tray", 300, 200, 20, 1e9)

	items := make([]*boxpacker3.Item, 0, 20)

	for i := range 20 {
		items = append(items, boxpacker3.NewItem(fmt.Sprintf("piece-%d", i), 60, 50, 20, 1))
	}

	for _, rule := range boxpacker3.BoxSelections() {
		result, err := rulePacker(boxpacker3.OrderDecreasing, rule).
			Pack(t.Context(), []*boxpacker3.Box{box}, items)
		if err != nil {
			t.Fatal(err)
		}

		for _, bad := range auditResult(result, items, 0) {
			t.Errorf("rule %s: %s — %s", rule, bad.kind, bad.note)
		}

		t.Logf("%s left %d of 20 on a tray that tiles exactly", rule, len(result.Unpacked))
	}
}

type noTopShelf struct{}

func (noTopShelf) Allow(box *boxpacker3.Container, _ *boxpacker3.Item,
	position boxpacker3.Pivot, dimension boxpacker3.Dimension,
) bool {
	return position[2]+dimension[2] <= box.Depth()*0.9+1e-9
}
