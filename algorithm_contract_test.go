package boxpacker3_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

const (
	classGlass = "glass"
	boxSmall   = "small"
	groupPair  = "pair"
)

type stubAlgorithm struct {
	seenRules boxpacker3.Rules
}

func (*stubAlgorithm) Name() string { return "Stub" }

func (a *stubAlgorithm) Pack(_ context.Context, problem *boxpacker3.Problem) (*boxpacker3.Packing, error) {
	a.seenRules = problem.Rules()

	box := problem.Open(problem.Boxes()[0], 0)
	leftover := make([]boxpacker3.Instance, 0)

	for _, instance := range problem.Instances() {
		if !box.Fit(instance) {
			leftover = append(leftover, instance)
		}
	}

	return problem.Packing([]*boxpacker3.Container{box}, leftover), nil
}

type recordingFinisher struct {
	ran *int
}

func (recordingFinisher) Name() string { return "Recording" }

func (f recordingFinisher) Finish(_ context.Context, _ *boxpacker3.Problem, _ *boxpacker3.Packing) error {
	*f.ran++

	return nil
}

func TestAlgorithmContract_RulesAndFinishersReachACallersAlgorithm(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{boxpacker3.NewBox("crate", 100, 100, 100, 1e9)}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("a", 50, 50, 50, 1),
		boxpacker3.NewItem("b", 50, 50, 50, 1),
		boxpacker3.NewItem("c", 50, 50, 50, 1),
	}

	floorOnly := boxpacker3.PlacementRuleFunc(
		func(_ *boxpacker3.Container, _ *boxpacker3.Item, position boxpacker3.Pivot, _ boxpacker3.Dimension) bool {
			return position[2] == 0
		})

	ran := 0
	algorithm := &stubAlgorithm{seenRules: boxpacker3.Rules{}}

	packer := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(algorithm),
		boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: 0.5, Placement: []boxpacker3.PlacementRule{floorOnly}}),
		boxpacker3.WithFinishers(recordingFinisher{ran: &ran}),
	)

	result, err := packer.Pack(t.Context(), boxes, items)
	require.NoError(t, err)

	require.InDelta(t, 0.5, algorithm.seenRules.MinSupportRatio, 1e-9, "the algorithm sees the packer's rules")
	require.Equal(t, 1, ran, "the finisher ran on the algorithm's answer")

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			require.Zero(t, item.Position[2], "the placement rule held for %s", item.ID())
		}
	}

	require.Len(t, result.Boxes, 1)
	require.Len(t, result.Boxes[0].Items, 3, "three cubes fit on the floor of a 100 box")
	require.Equal(t, "Stub", result.Report.Algorithm)
}

func TestRules_AdmissionKeepsAPieceOutOfABox(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("sack", 100, 100, 100, 1e9),
		boxpacker3.NewBox("crate", 100, 100, 100, 1e9),
	}
	items := []*boxpacker3.Item{
		mustItem(t, boxpacker3.ItemSpec{ID: "vase", Width: 40, Height: 40, Depth: 40, Weight: 1, Class: classGlass}),
		boxpacker3.NewItem("towel", 40, 40, 40, 1),
	}

	noGlassInSacks := boxpacker3.AdmissionRuleFunc(func(box *boxpacker3.Container, item *boxpacker3.Item) bool {
		return box.ID() != "sack" || item.Class() != classGlass
	})

	result, err := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit)),
		boxpacker3.WithRules(boxpacker3.Rules{Admission: []boxpacker3.AdmissionRule{noGlassInSacks}}),
		boxpacker3.WithFinishers(),
	).Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Empty(t, result.Unpacked)

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			if item.ID() == "vase" {
				require.Equal(t, "crate", box.ID(), "the vase may not go in the sack")
			}
		}
	}
}

func TestRules_PairKeepsTwoPiecesApart(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("a", 100, 100, 100, 1e9),
		boxpacker3.NewBox("b", 100, 100, 100, 1e9),
	}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("red", 40, 40, 40, 1),
		boxpacker3.NewItem("blue", 40, 40, 40, 1),
	}

	apart := boxpacker3.PairRuleFunc(func(a, b *boxpacker3.Item) bool {
		return a.ID() != "red" || b.ID() != "blue"
	})

	result, err := boxpacker3.NewPacker(
		boxpacker3.WithRules(boxpacker3.Rules{Pair: []boxpacker3.PairRule{apart}}),
		boxpacker3.WithFinishers(),
	).Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Empty(t, result.Unpacked)
	require.Len(t, result.Boxes, 2, "the pair rule holds both ways round, so they part")
}

func TestFinishers_RehomeTakesTheSmallerBox(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox(boxSmall, 50, 50, 50, 1e9),
		boxpacker3.NewBox("large", 100, 100, 100, 1e9),
	}
	items := []*boxpacker3.Item{boxpacker3.NewItem("cube", 40, 40, 40, 1)}

	worst := boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectWorstFit)

	asIs, err := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(worst), boxpacker3.WithFinishers()).
		Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Equal(t, "large", asIs.Boxes[0].ID())

	rehomed, err := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(worst), boxpacker3.WithFinishers(boxpacker3.Rehome)).
		Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Len(t, rehomed.Boxes, 1)
	require.Equal(t, boxSmall, rehomed.Boxes[0].ID(), "the cube fits the small box entire")
	require.Len(t, rehomed.Boxes[0].Items, 1)
}

func TestFinishers_RehomeRespectsTheStock(t *testing.T) {
	t.Parallel()

	small, err := boxpacker3.NewBoxFromSpec(boxpacker3.BoxSpec{
		ID: boxSmall, OuterWidth: 50, OuterHeight: 50, OuterDepth: 50, MaxWeight: 1e9, Quantity: 1,
	})
	require.NoError(t, err)

	large, err := boxpacker3.NewBoxFromSpec(boxpacker3.BoxSpec{
		ID: "large", OuterWidth: 100, OuterHeight: 100, OuterDepth: 100, MaxWeight: 1e9, Quantity: 2,
	})
	require.NoError(t, err)

	items := []*boxpacker3.Item{
		boxpacker3.NewItem("red", 40, 40, 40, 1),
		boxpacker3.NewItem("blue", 40, 40, 40, 1),
	}

	apart := boxpacker3.PairRuleFunc(func(a, b *boxpacker3.Item) bool {
		return a.ID() == b.ID()
	})

	result, err := boxpacker3.NewPacker(
		boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit)),
		boxpacker3.WithRules(boxpacker3.Rules{Pair: []boxpacker3.PairRule{apart}}),
		boxpacker3.WithFinishers(boxpacker3.Rehome),
	).Pack(t.Context(), []*boxpacker3.Box{small, large}, items)
	require.NoError(t, err)
	require.Len(t, result.Boxes, 2)

	kinds := map[string]int{}
	for _, box := range result.Boxes {
		kinds[box.ID()]++
	}

	require.Equal(t, map[string]int{boxSmall: 1, "large": 1}, kinds)
}

func TestFinishers_NoneStillKeepsGroupsWhole(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("a", 60, 60, 60, 1e9),
		boxpacker3.NewBox("b", 60, 60, 60, 1e9),
	}
	items := []*boxpacker3.Item{
		mustItem(t, boxpacker3.ItemSpec{ID: "one", Width: 50, Height: 50, Depth: 50, Weight: 1, Group: groupPair}),
		mustItem(t, boxpacker3.ItemSpec{ID: "two", Width: 50, Height: 50, Depth: 50, Weight: 1, Group: groupPair}),
	}

	result, err := boxpacker3.NewPacker(boxpacker3.WithFinishers()).Pack(t.Context(), boxes, items)
	require.NoError(t, err)

	boxesHolding := map[string]bool{}

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			boxesHolding[box.ID()] = true
			_ = item
		}
	}

	require.LessOrEqual(t, len(boxesHolding), 1, "a group is in one box or none")
}

func TestGoals_ShippingCostPrefersTheCheaperConsignment(t *testing.T) {
	t.Parallel()

	tariff := boxpacker3.Tariff{
		PerBox:     map[string]float64{"crate": 5},
		PerKg:      2,
		DimDivisor: 5000,
		Minimum:    10,
	}

	goal := boxpacker3.ShippingCost("Carrier", tariff)
	require.Equal(t, "Carrier", goal.Name())

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("crate", 100, 100, 100, 1e9),
		boxpacker3.NewBox("pallet", 400, 400, 400, 1e9),
	}
	items := []*boxpacker3.Item{boxpacker3.NewItem("cube", 90, 90, 90, 1000)}

	inCrate, err := rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit).
		Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Equal(t, "crate", inCrate.Boxes[0].ID())

	inPallet, err := rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectWorstFit).
		Pack(t.Context(), boxes, items)
	require.NoError(t, err)
	require.Equal(t, "pallet", inPallet.Boxes[0].ID())

	require.Negative(t, goal.Compare(inCrate, inPallet), "the crate is cheaper by volumetric weight")
	require.Positive(t, goal.Compare(inPallet, inCrate))

	require.InDelta(t, 5+2*1000, tariff.Total(inCrate), 1e-9, "gross weight beats the volumetric weight here")
	require.InDelta(t, 2*(400*400*400/5000), tariff.Total(inPallet), 1e-9, "the pallet is charged on its volume")
}

func TestGoals_MinimumIsAFloorOnTheBill(t *testing.T) {
	t.Parallel()

	tariff := boxpacker3.Tariff{PerBox: nil, PerKg: 1, DimDivisor: 0, Minimum: 25}
	empty := &boxpacker3.Result{Boxes: nil, Unpacked: nil, Report: boxpacker3.Report{}}

	require.InDelta(t, 25.0, tariff.Total(empty), 1e-9)
}

func TestGoals_WeightedAddsItsTermsUp(t *testing.T) {
	t.Parallel()

	boxesCriterion := boxpacker3.Criterion{
		Name:    "boxes",
		Measure: func(result *boxpacker3.Result) float64 { return float64(len(result.Boxes)) },
		Higher:  false,
	}
	fill := boxpacker3.Criterion{
		Name:    "fill",
		Measure: func(result *boxpacker3.Result) float64 { return boxpacker3.MetricsOf(result).Fill },
		Higher:  true,
	}

	goal := boxpacker3.Weighted("FillOverBoxes",
		boxpacker3.Term{Criterion: boxesCriterion, Weight: 1},
		boxpacker3.Term{Criterion: fill, Weight: 100},
	)

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("small-a", 60, 60, 60, 1e9),
		boxpacker3.NewBox("small-b", 60, 60, 60, 1e9),
		boxpacker3.NewBox("large", 200, 200, 200, 1e9),
	}
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("one", 50, 50, 50, 1),
		boxpacker3.NewItem("two", 50, 50, 50, 1),
	}

	spread, err := rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectWorstFit).
		Pack(t.Context(), boxes, items)
	require.NoError(t, err)

	together, err := rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit).
		Pack(t.Context(), boxes, items)
	require.NoError(t, err)

	require.Equal(t, goal.Compare(spread, together) < 0,
		boxpacker3.MetricsOf(spread).Fill*100-float64(len(spread.Boxes)) >
			boxpacker3.MetricsOf(together).Fill*100-float64(len(together.Boxes)),
		"the weighted goal ranks on the sum it states")
}

func TestGoals_NamesAreUniqueAndResolve(t *testing.T) {
	t.Parallel()

	seen := map[string]bool{}

	for _, goal := range boxpacker3.Goals() {
		require.NotEmpty(t, goal.Name())
		require.False(t, seen[goal.Name()], "two goals answer to %q", goal.Name())

		seen[goal.Name()] = true

		found, ok := boxpacker3.GoalByName(goal.Name())
		require.True(t, ok)
		require.Equal(t, goal, found)

		answer, ok := boxpacker3.BestForName(goal.Name())
		require.True(t, ok)
		require.Equal(t, goal.Name(), answer.Name())
	}

	require.Len(t, seen, len(boxpacker3.GoalNames()))

	_, ok := boxpacker3.GoalByName("NoSuchGoal")
	require.False(t, ok)

	_, ok = boxpacker3.BestForName("NoSuchGoal")
	require.False(t, ok)
}

func TestGoals_AreNeverBeatenByARuleOnTheirOwnMeasure(t *testing.T) {
	t.Parallel()

	measures := map[string]func(*boxpacker3.Result) float64{
		nameFewestBoxes:    func(r *boxpacker3.Result) float64 { return float64(len(r.Boxes)) },
		nameLeastVolume:    func(r *boxpacker3.Result) float64 { return boxpacker3.MetricsOf(r).BoxVolume },
		nameMostItems:      func(r *boxpacker3.Result) float64 { return float64(len(r.Unpacked)) },
		nameHighestFill:    func(r *boxpacker3.Result) float64 { return -boxpacker3.MetricsOf(r).Fill },
		nameBalancedWeight: func(r *boxpacker3.Result) float64 { return boxpacker3.MetricsOf(r).WeightSpread },
	}

	random := rand.New(rand.NewSource(11)) //nolint:gosec

	for seed := range 25 {
		boxes := shelfOf(random, 4+seed%3, 110, 160, 1e9)
		items := cargoOf(random, 10+seed%10, 30, 80, 100, 500)

		for _, goal := range boxpacker3.Goals() {
			answer, ok := boxpacker3.BestForName(goal.Name())
			require.True(t, ok)

			chosen := packed(t, boxpacker3.NewPacker(boxpacker3.WithAlgorithm(answer)), boxes, items)
			measure := measures[goal.Name()]

			for _, rule := range boxpacker3.EveryRuleSettings() {
				byRule := packed(t, rulePacker(rule.Order, rule.Selection), boxes, items)

				if len(byRule.Unpacked) != len(chosen.Unpacked) {
					continue
				}

				require.LessOrEqual(t, measure(chosen), measure(byRule)+1e-9,
					"seed %d: %s was beaten by %s on its own measure", seed, goal.Name(), rule.Name())
			}
		}
	}
}
