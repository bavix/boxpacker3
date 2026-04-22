package boxpacker3_test

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

func packedWith(t *testing.T, strategy boxpacker3.RuleSettings,
	boxes []*boxpacker3.Box, items []*boxpacker3.Item,
) (int, int) {
	t.Helper()

	result, err := rulePacker(strategy.Order, strategy.Selection).
		Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	return usedBoxCountOf(result), len(result.Unpacked)
}

func TestMinimizeBoxes_NoRuleDoesBetter(t *testing.T) {
	t.Parallel()

	wins := 0

	for seed := range 200 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 4+random.Intn(4), 220, 150, 1e9)
		items := cargoOf(random, 8+random.Intn(14), 40, 100, 100, 900)

		best, err := boxpacker3.NewPacker(
			boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes))).
			Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		ours, left := usedBoxCountOf(best), len(best.Unpacked)

		for _, strategy := range boxpacker3.EveryRuleSettings() {
			theirs, theirLeft := packedWith(t, strategy, boxes, items)

			if theirLeft < left {
				t.Fatalf("seed %d: %s left %d goods behind where minimizing boxes left %d",
					seed, strategy, theirLeft, left)
			}

			if theirLeft == left && theirs < ours {
				t.Fatalf("seed %d: %s used %d boxes where minimizing boxes used %d",
					seed, strategy, theirs, ours)
			}

			if theirs > ours {
				wins++
			}
		}
	}

	t.Logf("minimizing boxes beat a named rule %d times", wins)
}

func TestMinimizeBoxes_ReachesTheVolumeBound(t *testing.T) {
	t.Parallel()

	atBound := 0

	for seed := range 200 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 4+random.Intn(4), 220, 150, 1e9)
		items := cargoOf(random, 8+random.Intn(14), 40, 100, 100, 900)

		best, err := boxpacker3.NewPacker(
			boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes))).
			Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		if len(best.Unpacked) > 0 {
			continue
		}

		volume := 0.0
		for _, item := range items {
			volume += item.Volume()
		}

		roomiest := 0.0
		for _, box := range boxes {
			roomiest = max(roomiest, box.Width()*box.Height()*box.Depth())
		}

		if float64(usedBoxCountOf(best)) <= math.Ceil(volume/roomiest) {
			atBound++
		}
	}

	t.Logf("at the volume bound on %d of 200 orders", atBound)

	if atBound < 150 {
		t.Errorf("only %d of 200 orders reached the volume bound", atBound)
	}
}

func TestMinimizeBoxes_KeepsTheConstraints(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("hold-1", 400, 400, 400, 1e9),
		boxpacker3.NewBox("hold-2", 400, 400, 400, 1e9),
	}

	items := []*boxpacker3.Item{
		classedItem(t, "flour", 100, 1000, classFood, classChemicals),
		classedItem(t, "bleach", 100, 1000, classChemicals),
	}

	result, err := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes))).
		Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	if used := usedBoxCountOf(result); used != 2 {
		t.Errorf("goods that must stay apart went into %d boxes", used)
	}
}

func TestGoals_EachOneAnswersItsOwnQuestion(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewSource(11)) //nolint:gosec
	boxes := shelfOf(random, 6, 220, 150, 1e9)
	items := cargoOf(random, 18, 40, 100, 100, 900)

	fewest, err := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes))).
		Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	tightest, err := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.LeastVolume))).
		Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	most, err := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.MostItems))).
		Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	if usedBoxCountOf(fewest) > usedBoxCountOf(tightest) {
		t.Errorf("minimizing boxes used %d, tightest packing %d",
			usedBoxCountOf(fewest), usedBoxCountOf(tightest))
	}

	if boxVolumeOf(tightest) > boxVolumeOf(fewest) {
		t.Errorf("tightest packing came to %.0f mm³ of boxes, minimizing boxes %.0f",
			boxVolumeOf(tightest), boxVolumeOf(fewest))
	}

	if len(most.Unpacked) > len(fewest.Unpacked) {
		t.Errorf("maximizing items left %d behind, minimizing boxes %d",
			len(most.Unpacked), len(fewest.Unpacked))
	}
}

func boxVolumeOf(result *boxpacker3.Result) float64 {
	volume := 0.0

	for _, box := range result.Boxes {
		if len(box.Items) == 0 {
			continue
		}

		volume += box.Box.Width() * box.Box.Height() * box.Box.Depth()
	}

	return volume
}

func spread(result *boxpacker3.Result) float64 {
	weights := make([]float64, 0, len(result.Boxes))
	total := 0.0

	for _, box := range result.Boxes {
		if len(box.Items) == 0 {
			continue
		}

		weights = append(weights, box.Stats.ItemsWeight)
		total += box.Stats.ItemsWeight
	}

	if len(weights) < 2 {
		return 0
	}

	mean := total / float64(len(weights))
	variance := 0.0

	for _, weight := range weights {
		variance += (weight - mean) * (weight - mean)
	}

	return math.Sqrt(variance / float64(len(weights)))
}

func bestSpreadByAnyRule(t *testing.T, boxes []*boxpacker3.Box, items []*boxpacker3.Item) float64 {
	t.Helper()

	best := math.MaxFloat64

	for _, strategy := range boxpacker3.EveryRuleSettings() {
		result, err := rulePacker(strategy.Order, strategy.Selection).
			Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		if len(result.Unpacked) == 0 {
			best = min(best, spread(result))
		}
	}

	return best
}

func TestBestFor_NamesItsGoal(t *testing.T) {
	t.Parallel()

	for _, goal := range boxpacker3.Goals() {
		if name := boxpacker3.BestFor(goal).Name(); name != goal.Name() {
			t.Errorf("the answer for %s calls itself %q", goal.Name(), name)
		}
	}
}

func ExampleBestFor() {
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small", 200, 200, 200, 10000),
		boxpacker3.NewBox("large", 400, 400, 400, 40000),
	}

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("a", 190, 190, 190, 1000),
		boxpacker3.NewItem("b", 190, 190, 190, 1000),
	}

	result, err := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes))).Pack(context.Background(), boxes, items)
	if err != nil {
		panic(err)
	}

	used := 0

	for _, box := range result.Boxes {
		if len(box.Items) > 0 {
			used++
		}
	}

	fmt.Println(used)
	// Output: 1
}

func TestMinimizeBoxes_IsDeterministic(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewSource(7)) //nolint:gosec
	boxes := shelfOf(random, 5, 220, 150, 1e9)
	items := cargoOf(random, 16, 40, 100, 100, 900)

	first := packed(t, boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes))), boxes, items)

	for range 20 {
		again := packed(t, boxpacker3.NewPacker(
			boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes))), boxes, items)

		if fingerprint(first) != fingerprint(again) {
			t.Fatal("the same order packed two different ways")
		}
	}
}

func TestMinimizeBoxes_CancelledContextReportsTheError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	random := rand.New(rand.NewSource(3)) //nolint:gosec
	boxes := shelfOf(random, 4, 220, 150, 1e9)
	items := cargoOf(random, 12, 40, 100, 100, 900)

	_, err := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes))).
		Pack(ctx, boxes, items)
	if err == nil {
		t.Error("a cancelled context produced an answer")
	}
}

func TestMinimizeBoxes_RespectsTheBoxSupply(t *testing.T) {
	t.Parallel()

	supply, err := boxpacker3.NewBoxFromSpec(boxpacker3.BoxSpec{
		ID: "crate", OuterWidth: 300, OuterHeight: 300, OuterDepth: 300,
		MaxWeight: 1e9, Quantity: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	items := make([]*boxpacker3.Item, 0, 12)
	for i := range 12 {
		items = append(items, boxpacker3.NewItem(fmt.Sprintf("cube-%d", i), 140, 140, 140, 10))
	}

	result, err := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes))).
		Pack(t.Context(), []*boxpacker3.Box{supply}, items)
	if err != nil {
		t.Fatal(err)
	}

	if used := usedBoxCountOf(result); used > 2 {
		t.Errorf("two crates were offered and %d were used", used)
	}
}

func TestMinimizeBoxes_KeepsTheSupportRule(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewSource(19)) //nolint:gosec
	boxes := shelfOf(random, 4, 220, 150, 1e9)
	items := cargoOf(random, 14, 40, 100, 100, 900)

	result, err := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes)),
		boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: 1})).
		Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			if ratio := supportRatioOf(box, item); ratio < 1-1e-6 {
				t.Errorf("%s rests on %.3f of its base", item.ID(), ratio)
			}
		}
	}
}

func TestMinimizeBoxes_KeepsTheLoadLimits(t *testing.T) {
	t.Parallel()

	for seed := range 100 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 4, 220, 150, 1e9)
		items, caps := randomCargo(t, random)

		result, err := boxpacker3.NewPacker(
			boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes))).
			Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		for _, breach := range overloaded(result, caps) {
			t.Fatalf("seed %d: %s", seed, breach)
		}
	}
}

func BenchmarkMinimizeBoxes(b *testing.B) {
	random := rand.New(rand.NewSource(5)) //nolint:gosec
	boxes := shelfOf(random, 8, 220, 150, 1e9)
	items := cargoOf(random, 30, 40, 100, 100, 900)

	b.Run("default", func(b *testing.B) {
		for range b.N {
			_, _ = boxpacker3.NewPacker().Pack(b.Context(), boxes, items)
		}
	})

	b.Run("best of every rule", func(b *testing.B) {
		for range b.N {
			_, _ = boxpacker3.NewPacker(boxpacker3.WithAlgorithm(
				boxpacker3.EveryRule(boxpacker3.FewestBoxes))).Pack(b.Context(), boxes, items)
		}
	})

	b.Run("fewest boxes", func(b *testing.B) {
		for range b.N {
			_, _ = boxpacker3.NewPacker(
				boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes))).Pack(b.Context(), boxes, items)
		}
	})
}

func TestMinimizeBoxes_LeavesLessBehindOnCrowdedOrders(t *testing.T) {
	t.Parallel()

	better, worse, ours, theirs := 0, 0, 0, 0

	for seed := range 200 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := make([]*boxpacker3.Box, 0, 3)
		for i := range 3 {
			boxes = append(boxes, boxpacker3.NewBox(
				fmt.Sprintf("crate-%d", i), 300, 300, 300, 1e9))
		}

		items := make([]*boxpacker3.Item, 0, 24)

		for i := range 18 + random.Intn(6) {
			side := float64(120 + random.Intn(60))
			items = append(items, boxpacker3.NewItem(
				fmt.Sprintf("item-%d", i), side, side, side, 10))
		}

		best, err := boxpacker3.NewPacker(
			boxpacker3.WithAlgorithm(boxpacker3.BestFor(boxpacker3.FewestBoxes))).
			Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		plain, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		ours += len(best.Unpacked)
		theirs += len(plain.Unpacked)

		switch {
		case len(best.Unpacked) < len(plain.Unpacked):
			better++
		case len(best.Unpacked) > len(plain.Unpacked):
			worse++
		}
	}

	t.Logf("left behind: %d against the default rule's %d — better on %d orders, worse on %d",
		ours, theirs, better, worse)

	if ours > theirs {
		t.Errorf("minimizing boxes left %d goods behind where the default left %d", ours, theirs)
	}
}

func TestGoals_AreAllRepaired(t *testing.T) {
	t.Parallel()

	answers := map[string]boxpacker3.Goal{
		nameFewestBoxes: boxpacker3.FewestBoxes,
		nameLeastVolume: boxpacker3.LeastVolume,
		nameMostItems:   boxpacker3.MostItems,
	}

	for name, goal := range answers {
		build := func() boxpacker3.Algorithm { return boxpacker3.BestFor(goal) }

		if got := build().Name(); got != name {
			t.Errorf("want name %q, got %q", name, got)
		}

		left, plainLeft := 0, 0

		for seed := range 100 {
			random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

			boxes := make([]*boxpacker3.Box, 0, 3)
			for i := range 3 {
				boxes = append(boxes, boxpacker3.NewBox(
					fmt.Sprintf("crate-%d", i), 300, 300, 300, 1e9))
			}

			items := make([]*boxpacker3.Item, 0, 24)

			for i := range 18 + random.Intn(6) {
				side := float64(120 + random.Intn(60))
				items = append(items, boxpacker3.NewItem(
					fmt.Sprintf("item-%d", i), side, side, side, 10))
			}

			answer, err := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(build())).
				Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			plain, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			left += len(answer.Unpacked)
			plainLeft += len(plain.Unpacked)
		}

		t.Logf("%-16s left %d behind, the default rule %d", name, left, plainLeft)

		if left > plainLeft {
			t.Errorf("%s left %d goods behind where the default left %d", name, left, plainLeft)
		}
	}
}

func TestMinimizeBoxes_WhereTheGainComesFrom(t *testing.T) {
	t.Parallel()

	crowded := crowdedOrderFor

	leftBy := func(build func() boxpacker3.Algorithm) int {
		left := 0

		for seed := range 200 {
			boxes, items := crowded(seed)

			options := []boxpacker3.Option{}
			if build != nil {
				options = append(options, boxpacker3.WithAlgorithm(build()))
			}

			result, err := boxpacker3.NewPacker(options...).Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			left += len(result.Unpacked)
		}

		return left
	}

	plain := leftBy(nil)
	searched := leftBy(func() boxpacker3.Algorithm {
		return boxpacker3.EveryRule(boxpacker3.FewestBoxes)
	})
	repaired := leftBy(func() boxpacker3.Algorithm {
		return boxpacker3.BestFor(boxpacker3.FewestBoxes)
	})

	t.Logf("left behind: %d by the default rule, %d by trying every rule, %d after the repairs",
		plain, searched, repaired)

	if searched >= plain {
		t.Errorf("trying every rule left %d, no better than the default's %d", searched, plain)
	}

	if repaired > searched {
		t.Errorf("the repairs left %d, worse than the %d they started from", repaired, searched)
	}
}

func TestGoals_EachIsAtLeastAsGoodAsTheRules(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewSource(23)) //nolint:gosec

	boxes := identicalShelf(6, 220, 180, 160, 1e9)
	items := cargoOf(random, 26, 40, 70, 100, 900)

	answer := func(build func() boxpacker3.Algorithm) *boxpacker3.Result {
		result, err := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(build())).
			Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		return result
	}

	balanced := answer(func() boxpacker3.Algorithm { return boxpacker3.BestFor(boxpacker3.BalancedWeight) })
	fewest := answer(func() boxpacker3.Algorithm { return boxpacker3.BestFor(boxpacker3.FewestBoxes) })
	fullest := answer(func() boxpacker3.Algorithm { return boxpacker3.BestFor(boxpacker3.HighestFill) })

	t.Logf("weight spread: balanced %.0f, fewest boxes %.0f", spread(balanced), spread(fewest))

	best := bestSpreadByAnyRule(t, boxes, items)

	if len(balanced.Unpacked) == 0 && spread(balanced) > best+1e-6 {
		t.Errorf("the balanced goal spread the weight by %.2f where a rule managed %.2f",
			spread(balanced), best)
	}

	if usedBoxCountOf(fullest) > usedBoxCountOf(fewest) {
		t.Logf("fullest boxes used %d, fewest boxes %d",
			usedBoxCountOf(fullest), usedBoxCountOf(fewest))
	}
}

func constrainedOrder(t *testing.T, random *rand.Rand) ([]*boxpacker3.Item, map[string]float64) {
	t.Helper()

	items := make([]*boxpacker3.Item, 0, 16)
	caps := map[string]float64{}

	for i := range 10 + random.Intn(6) {
		id := fmt.Sprintf("item-%d", i)
		side := float64(50 + random.Intn(50))
		weight := float64(100 + random.Intn(600))
		limit := 0.0

		if i%4 == 0 {
			limit = float64(200 + random.Intn(600))
		}

		caps[id] = limit

		class := classFood
		away := []string{classChemicals}

		if i%3 == 0 {
			class, away = classChemicals, nil
		}

		built, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
			ID: id, Width: side, Height: side, Depth: side, Weight: weight,
			MaxLoadOnTop: limit, Class: class, SeparateFrom: away,
		})
		if err != nil {
			t.Fatal(err)
		}

		items = append(items, built)
	}

	return items, caps
}

func TestGoals_AllHonourTheConstraints(t *testing.T) {
	t.Parallel()

	goals := make([]func() boxpacker3.Algorithm, 0, len(boxpacker3.Goals()))

	for _, goal := range boxpacker3.Goals() {
		goals = append(goals, func() boxpacker3.Algorithm { return boxpacker3.BestFor(goal) })
	}

	apart := map[string]map[string]bool{
		classFood:      {classChemicals: true},
		classChemicals: {classFood: true},
	}

	for _, build := range goals {
		for seed := range 60 {
			random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

			boxes := shelfOf(random, 4, 200, 140, 1e9)
			items, caps := constrainedOrder(t, random)

			result, err := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(build())).
				Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			for _, mixed := range mixedBoxes(result, apart) {
				t.Fatalf("%s seed %d: %s", build().Name(), seed, mixed)
			}

			for _, breach := range overloaded(result, caps) {
				t.Fatalf("%s seed %d: %s", build().Name(), seed, breach)
			}
		}
	}
}

func TestGoals_EveryNamedGoalHasAnAlgorithm(t *testing.T) {
	t.Parallel()

	for _, name := range boxpacker3.GoalNames() {
		algorithm, ok := boxpacker3.BestForName(name)
		if !ok {
			t.Errorf("%s is offered by name but has no algorithm", name)

			continue
		}

		if algorithm.Name() != name {
			t.Errorf("the algorithm for %s calls itself %q", name, algorithm.Name())
		}

		if _, ok := boxpacker3.GoalByName(name); !ok {
			t.Errorf("%s is offered by name but has no goal", name)
		}
	}

	if _, ok := boxpacker3.BestForName("NoSuchGoal"); ok {
		t.Error("a goal nobody named came back with an algorithm")
	}
}

func crowdedOrderFor(seed int) ([]*boxpacker3.Box, []*boxpacker3.Item) {
	random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

	boxes := make([]*boxpacker3.Box, 0, 3)
	for i := range 3 {
		boxes = append(boxes, boxpacker3.NewBox(
			fmt.Sprintf("crate-%d", i), 300, 300, 300, 1e9))
	}

	items := make([]*boxpacker3.Item, 0, 24)

	for i := range 18 + random.Intn(6) {
		side := float64(120 + random.Intn(60))
		items = append(items, boxpacker3.NewItem(
			fmt.Sprintf("item-%d", i), side, side, side, 10))
	}

	return boxes, items
}
