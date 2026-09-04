package boxpacker3_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

func checkPlacements(box boxpacker3.PackedBox, support float64) []violation {
	found := make([]violation, 0)
	items := box.Items
	span := [3]float64{box.Box.Width(), box.Box.Height(), box.Box.Depth()}

	for _, item := range items {
		position := item.Position
		dimension := item.Dimension

		for axis := range span {
			if position[axis] < -1e-9 || position[axis]+dimension[axis] > span[axis]+1e-9 {
				found = append(found, violation{
					kind: "out-of-bounds",
					note: fmt.Sprintf("%s in %s axis %d", item.ID(), box.ID(), axis),
				})
			}
		}

		if support > 0 && supportUnder(box, item) < support-1e-6 {
			found = append(found, violation{
				kind: "unsupported",
				note: fmt.Sprintf("%s in %s has %.2f of %.2f", item.ID(), box.ID(),
					supportUnder(box, item), support),
			})
		}
	}

	for i, a := range items {
		for _, b := range items[i+1:] {
			if overlapping(a, b) {
				found = append(found, violation{
					kind: "overlap",
					note: fmt.Sprintf("%s and %s in %s", a.ID(), b.ID(), box.ID()),
				})
			}
		}
	}

	return found
}

func TestAudit_PackingInvariants(t *testing.T) {
	t.Parallel()

	tally := map[string]int{}

	for seed := range 900 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 2+random.Intn(4), 90, 200, float64(600+random.Intn(4000)))
		items := groupedOrder(t, random, 5+random.Intn(18))
		support := []float64{0, 0.5, 1}[seed%3]

		for _, rule := range boxpacker3.BoxSelections() {
			options := []boxpacker3.Option{
				boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, rule)),
				boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: support}),
			}

			if seed%2 == 0 {
				options = append(options, boxpacker3.WithFinishers(boxpacker3.BalanceWeight{MaxBoxes: 5}))
			}

			result, err := boxpacker3.NewPacker(options...).Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			for _, bad := range auditResult(result, items, support) {
				tally[bad.kind]++

				if tally[bad.kind] <= 3 {
					t.Errorf("seed %d rule %s support %.1f: %s — %s",
						seed, rule, support, bad.kind, bad.note)
				}
			}
		}
	}

	t.Logf("violations: %v", tally)
}

func TestAudit_SearchStaysValid(t *testing.T) {
	t.Parallel()

	tally := map[string]int{}

	for seed := range 60 {
		boxes, items := randomOrder(int64(seed))

		search, err := boxpacker3.NewSearch(96, 2)
		if err != nil {
			t.Fatal(err)
		}

		result, err := boxpacker3.NewPacker(
			boxpacker3.WithAlgorithm(search),
			boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: 0.5}),
		).Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		for _, bad := range auditResult(result, items, 0.5) {
			tally[bad.kind]++

			if tally[bad.kind] <= 3 {
				t.Errorf("seed %d search: %s — %s", seed, bad.kind, bad.note)
			}
		}
	}

	t.Logf("search violations: %v", tally)
}

func TestAudit_SingleContainerStaysValid(t *testing.T) {
	t.Parallel()

	tally := map[string]int{}

	for seed := range 60 {
		boxes, items := randomOrder(int64(seed))

		result, err := singlePacker().Pack(t.Context(), []*boxpacker3.Box{boxes[0]}, items)
		if err != nil {
			t.Fatal(err)
		}

		for _, bad := range auditResult(result, items, 0) {
			tally[bad.kind]++

			if tally[bad.kind] <= 3 {
				t.Errorf("seed %d single: %s — %s", seed, bad.kind, bad.note)
			}
		}
	}

	t.Logf("single-container violations: %v", tally)
}

func TestAudit_TwoDimensionalStaysFlat(t *testing.T) {
	t.Parallel()

	tally := map[string]int{}

	for seed := range 120 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := make([]*boxpacker3.Box, 0, 3)

		for i := range 1 + random.Intn(3) {
			boxes = append(boxes, boxpacker3.NewBox2D(
				fmt.Sprintf("sheet-%d", i),
				float64(300+random.Intn(700)), float64(200+random.Intn(500)), 100000))
		}

		items := make([]*boxpacker3.Item, 0, 24)

		for i := range 6 + random.Intn(18) {
			items = append(items, boxpacker3.NewItem2D(
				fmt.Sprintf("cut-%d", i),
				float64(40+random.Intn(200)), float64(30+random.Intn(160)), 40))
		}

		for _, rule := range boxpacker3.BoxSelections() {
			result, err := rulePacker(boxpacker3.OrderDecreasing, rule).
				Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			for _, bad := range append(auditResult(result, items, 0), leftTheSheet(result)...) {
				tally[bad.kind]++

				if tally[bad.kind] <= 3 {
					t.Errorf("seed %d rule %s: %s — %s", seed, rule, bad.kind, bad.note)
				}
			}
		}
	}

	t.Logf("2D violations: %v", tally)
}

func leftTheSheet(result *boxpacker3.Result) []violation {
	found := make([]violation, 0)

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			if item.Position[2] != 0 || item.Dimension[2] != 1 {
				found = append(found, violation{
					kind: "left-the-sheet",
					note: fmt.Sprintf("%s at z=%.2f depth %.2f",
						item.ID(), item.Position[2], item.Dimension[2]),
				})
			}
		}
	}

	return found
}

func TestAudit_ScaleRobustness(t *testing.T) {
	t.Parallel()

	tally := map[string]int{}

	for _, scale := range []float64{1e-3, 1, 1e3, 1e6} {
		for seed := range 60 {
			random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

			boxes := []*boxpacker3.Box{
				boxpacker3.NewBox("box", 200*scale, 160*scale, 140*scale, 1e12),
			}

			items := make([]*boxpacker3.Item, 0, 16)

			for i := range 6 + random.Intn(10) {
				items = append(items, boxpacker3.NewItem(
					fmt.Sprintf("item-%d", i),
					float64(20+random.Intn(50))*scale,
					float64(20+random.Intn(50))*scale,
					float64(20+random.Intn(50))*scale,
					100))
			}

			result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			for _, bad := range auditResult(result, items, 0) {
				key := fmt.Sprintf("%s@%g", bad.kind, scale)
				tally[key]++

				if tally[key] <= 2 {
					t.Errorf("scale %g seed %d: %s — %s", scale, seed, bad.kind, bad.note)
				}
			}
		}
	}

	t.Logf("scale violations: %v", tally)
}

func TestAudit_ExactFits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		box  [3]float64
		item [3]float64
		grid [3]int
	}{
		{"two by two by two", [3]float64{100, 100, 100}, [3]float64{50, 50, 50}, [3]int{2, 2, 2}},
		{"a single snug item", [3]float64{80, 60, 40}, [3]float64{80, 60, 40}, [3]int{1, 1, 1}},
		{"a column", [3]float64{30, 30, 300}, [3]float64{30, 30, 30}, [3]int{1, 1, 10}},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			box := boxpacker3.NewBox("box", test.box[0], test.box[1], test.box[2], 1e9)
			count := test.grid[0] * test.grid[1] * test.grid[2]
			items := make([]*boxpacker3.Item, 0, count)

			for i := range count {
				items = append(items, boxpacker3.NewItem(
					fmt.Sprintf("item-%d", i), test.item[0], test.item[1], test.item[2], 1))
			}

			for _, rule := range boxpacker3.BoxSelections() {
				result, err := rulePacker(boxpacker3.OrderDecreasing, rule).
					Pack(t.Context(), []*boxpacker3.Box{box}, items)
				if err != nil {
					t.Fatal(err)
				}

				if len(result.Unpacked) > 0 {
					t.Errorf("rule %s left %d of %d behind in an exact fit",
						rule, len(result.Unpacked), count)
				}

				for _, bad := range auditResult(result, items, 0) {
					t.Errorf("rule %s: %s — %s", rule, bad.kind, bad.note)
				}
			}
		})
	}
}

func TestAudit_CancelledContext(t *testing.T) {
	t.Parallel()

	boxes, items := randomOrder(11)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	result, err := boxpacker3.NewPacker().Pack(ctx, boxes, items)
	if err == nil {
		t.Fatalf("a cancelled context produced a result: %v", result != nil)
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("want context.Canceled, got %v", err)
	}

	if result != nil {
		t.Errorf("a cancelled pack still handed back a result")
	}
}

func TestAudit_InputsAreNotMutated(t *testing.T) {
	t.Parallel()

	boxes, items := randomOrder(23)

	before := make([][3]float64, len(items))
	for i, item := range items {
		before[i] = [3]float64{item.Width(), item.Height(), item.Depth()}
	}

	boxesBefore := make([][4]float64, len(boxes))
	for i, box := range boxes {
		boxesBefore[i] = [4]float64{box.Width(), box.Height(), box.Depth(), box.Volume()}
	}

	_, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	for i, item := range items {
		if [3]float64{item.Width(), item.Height(), item.Depth()} != before[i] {
			t.Errorf("%s was turned in the caller's own slice", item.ID())
		}
	}

	for i, box := range boxes {
		if ([4]float64{box.Width(), box.Height(), box.Depth(), box.Volume()}) != boxesBefore[i] {
			t.Errorf("%s came back changed in the caller's own slice", box.ID())
		}
	}
}

func TestAudit_PackerCarriesNoState(t *testing.T) {
	t.Parallel()

	first, firstItems := randomOrder(3)
	other, otherItems := randomOrder(9)

	for _, rule := range boxpacker3.BoxSelections() {
		packer := rulePacker(boxpacker3.OrderDecreasing, rule)

		before := fingerprint(packed(t, packer, first, firstItems))
		packed(t, packer, other, otherItems)
		after := fingerprint(packed(t, packer, first, firstItems))

		if before != after {
			t.Errorf("%s: the run in between changed the answer", rule)
		}
	}
}

func TestAudit_ResultsDoNotAlias(t *testing.T) {
	t.Parallel()

	boxes, items := randomOrder(5)
	packer := boxpacker3.NewPacker()

	kept, _ := packer.Pack(t.Context(), boxes, items)
	keptMark := fingerprint(kept)

	result, err := packer.Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	for i := range result.Boxes {
		result.Boxes[i].Items = nil
	}

	if fingerprint(kept) != keptMark {
		t.Error("the first result changed when the second was packed")
	}
}

func TestAudit_OrderAsGivenIsRespected(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("crate", 300, 300, 300, 1e9)}

	items := make([]*boxpacker3.Item, 0, 9)

	for i := range 9 {
		side := float64(40 + i*10)
		items = append(items, boxpacker3.NewItem(fmt.Sprintf("item-%d", i), side, side, side, 10))
	}

	asGiven := func(order []*boxpacker3.Item) *boxpacker3.Result {
		result, err := boxpacker3.NewPacker(
			boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderAsGiven, boxpacker3.SelectFirstFit)),
		).Pack(t.Context(), boxes, order)
		if err != nil {
			t.Fatal(err)
		}

		return result
	}

	forwards := asGiven(items)

	reversed := make([]*boxpacker3.Item, len(items))
	for i, item := range items {
		reversed[len(items)-1-i] = item
	}

	if fingerprint(forwards) == fingerprint(asGiven(reversed)) {
		t.Error("reversing the caller's order changed nothing, so it was not honoured")
	}
}

func TestAudit_RepeatedPointers(t *testing.T) {
	t.Parallel()

	packedAndLeft := func(result *boxpacker3.Result) (int, int) {
		packed := 0

		for _, box := range result.Boxes {
			packed += len(box.Items)
		}

		return packed, len(result.Unpacked)
	}

	t.Run("the same box twice", func(t *testing.T) {
		t.Parallel()

		box := boxpacker3.NewBox("crate", 200, 200, 200, 1e9)
		items := []*boxpacker3.Item{
			boxpacker3.NewItem("one", 190, 190, 190, 1),
			boxpacker3.NewItem("two", 190, 190, 190, 1),
		}

		result, err := boxpacker3.NewPacker().
			Pack(t.Context(), []*boxpacker3.Box{box, box}, items)
		if err != nil {
			t.Fatal(err)
		}

		packed, left := packedAndLeft(result)
		if packed+left != len(items) {
			t.Errorf("sent %d, got %d packed and %d unfit", len(items), packed, left)
		}

		if box.Volume() != 200*200*200.0 {
			t.Errorf("the caller's box came back changed: volume %.0f", box.Volume())
		}
	})

	t.Run("the same item twice", func(t *testing.T) {
		t.Parallel()

		item := boxpacker3.NewItem("twin", 90, 90, 90, 1)

		result, err := boxpacker3.NewPacker().Pack(t.Context(),
			[]*boxpacker3.Box{boxpacker3.NewBox("crate", 200, 200, 200, 1e9)},
			[]*boxpacker3.Item{item, item})
		if err != nil {
			t.Fatal(err)
		}

		packed, left := packedAndLeft(result)
		if packed+left != 2 {
			t.Errorf("sent the same item twice, got %d packed and %d unfit", packed, left)
		}

		for _, box := range result.Boxes {
			for _, bad := range checkPlacements(box, 0) {
				t.Errorf("%s — %s", bad.kind, bad.note)
			}
		}
	})
}

func TestAudit_DuplicateIdsAreNotLost(t *testing.T) {
	t.Parallel()

	build := func(count int) []*boxpacker3.Item {
		items := make([]*boxpacker3.Item, 0, count)
		for range count {
			items = append(items, boxpacker3.NewItem("book", 90, 90, 90, 100))
		}

		return items
	}

	countBack := func(result *boxpacker3.Result) int {
		packed := 0

		for _, box := range result.Boxes {
			packed += len(box.Items)
		}

		return packed + len(result.Unpacked)
	}

	t.Run("single container", func(t *testing.T) {
		t.Parallel()

		items := build(10)

		result, err := singlePacker().Pack(t.Context(),
			[]*boxpacker3.Box{boxpacker3.NewBox("crate", 200, 200, 100, 1e9)}, items)
		if err != nil {
			t.Fatal(err)
		}

		if countBack(result) != len(items) {
			t.Errorf("sent %d, got %d back", len(items), countBack(result))
		}
	})

	t.Run("several boxes", func(t *testing.T) {
		t.Parallel()

		for _, rule := range boxpacker3.BoxSelections() {
			items := build(10)

			result, err := rulePacker(boxpacker3.OrderDecreasing, rule).
				Pack(t.Context(), []*boxpacker3.Box{
					boxpacker3.NewBox("crate", 200, 200, 100, 1e9),
				}, items)
			if err != nil {
				t.Fatal(err)
			}

			if countBack(result) != len(items) {
				t.Errorf("rule %s: sent %d, got %d back", rule, len(items), countBack(result))
			}
		}
	})
}

func TestAudit_SameIdDifferentSize(t *testing.T) {
	t.Parallel()

	for _, rule := range boxpacker3.BoxSelections() {
		boxes := []*boxpacker3.Box{
			boxpacker3.NewBox("crate", 100, 100, 100, 1e9),
			boxpacker3.NewBox("crate", 400, 400, 400, 1e9),
		}

		items := []*boxpacker3.Item{
			boxpacker3.NewItem("small", 90, 90, 90, 1),
			boxpacker3.NewItem("big", 380, 380, 380, 1),
		}

		result, err := rulePacker(boxpacker3.OrderDecreasing, rule).
			Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		packed := 0
		for _, box := range result.Boxes {
			packed += len(box.Items)
		}

		if packed+len(result.Unpacked) != len(items) {
			t.Errorf("%s lost items: %d packed, %d unfit", rule, packed, len(result.Unpacked))
		}

		if rule != boxpacker3.SelectNextFit && packed != len(items) {
			t.Errorf("%s packed only %d of %d with a big enough box available",
				rule, packed, len(items))
		}
	}
}

func TestAudit_NoItemLeftWhileAnEmptyBoxWouldHoldIt(t *testing.T) {
	t.Parallel()

	rules := []boxpacker3.BoxSelection{
		boxpacker3.SelectFirstFit, boxpacker3.SelectBestFit, boxpacker3.SelectWorstFit,
		boxpacker3.SelectAlmostWorstFit, boxpacker3.SelectFullestBox,
	}

	misses := map[string]int{}

	for seed := range 600 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 2+random.Intn(4), 90, 200, float64(2000+random.Intn(6000)))
		items := cargoOf(random, 6+random.Intn(16), 20, 90, 100, 500)

		for _, rule := range rules {
			result, err := rulePacker(boxpacker3.OrderDecreasing, rule).
				Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			countMisses(t, boxes, result, rule.String(), misses)
		}
	}

	t.Logf("missed placements: %v", misses)
}

func countMisses(t *testing.T, shelf []*boxpacker3.Box, result *boxpacker3.Result, rule string, misses map[string]int) {
	t.Helper()

	used := map[*boxpacker3.Box]int{}
	for _, box := range result.Boxes {
		used[box.Box]++
	}

	spare := make([]*boxpacker3.Box, 0, len(shelf))

	for _, box := range shelf {
		if used[box] < box.Quantity() {
			spare = append(spare, box)
		}
	}

	for _, item := range result.Unpacked {
		for _, box := range spare {
			if !packsAlone(box, item.Item) {
				continue
			}

			misses[rule]++

			if misses[rule] <= 3 {
				t.Errorf("rule %s: %s was left behind while %s stood empty",
					rule, item.ID(), box.ID())
			}

			break
		}
	}
}

func TestAudit_EverythingAtOnce(t *testing.T) {
	t.Parallel()

	tally := map[string]int{}

	for seed := range 120 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 2+random.Intn(3), 120, 160, float64(2000+random.Intn(4000)))
		items := groupedOrder(t, random, 5+random.Intn(12))
		support := []float64{0, 0.5, 1}[seed%3]

		for name, algorithm := range everyAlgorithm(t) {
			result, err := boxpacker3.NewPacker(
				boxpacker3.WithAlgorithm(algorithm),
				boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: support}),
				boxpacker3.WithRules(boxpacker3.Rules{Placement: []boxpacker3.PlacementRule{noTopShelf{}}}),
				boxpacker3.WithFinishers(boxpacker3.BalanceWeight{MaxBoxes: 5}),
			).Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			bad := auditResult(result, items, support)
			bad = append(bad, aboveTheShelf(result)...)

			for _, broken := range groupSplits(t, result) {
				bad = append(bad, violation{"group-split", broken})
			}

			for _, one := range bad {
				tally[one.kind]++

				if tally[one.kind] <= 3 {
					t.Errorf("seed %d %s support %.1f: %s — %s",
						seed, name, support, one.kind, one.note)
				}
			}
		}
	}

	t.Logf("violations: %v", tally)
}

func everyAlgorithm(t *testing.T) map[string]boxpacker3.Algorithm {
	t.Helper()

	search, err := boxpacker3.NewSearch(96, 2)
	if err != nil {
		t.Fatal(err)
	}

	rules := make([]boxpacker3.Algorithm, 0, 6)
	for _, rule := range boxpacker3.BoxSelections() {
		rules = append(rules, boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, rule))
	}

	goal, _ := boxpacker3.GoalByName("BalancedWeight")

	return map[string]boxpacker3.Algorithm{
		"search":   search,
		"parallel": boxpacker3.NewPortfolio(goal, rules...),
	}
}

func aboveTheShelf(result *boxpacker3.Result) []violation {
	found := make([]violation, 0)

	for _, box := range result.Boxes {
		allowed := box.Box.Depth() * 0.9

		for _, item := range box.Items {
			if top := item.Position[2] + item.Dimension[2]; top > allowed+1e-6 {
				found = append(found, violation{
					kind: "constraint",
					note: fmt.Sprintf("%s reaches %.1f of the allowed %.1f", item.ID(), top, allowed),
				})
			}
		}
	}

	return found
}
