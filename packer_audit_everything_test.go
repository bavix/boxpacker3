package boxpacker3_test

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

const (
	auditSeeds        = 24
	auditSupportRatio = 0.6
	auditStableTilt   = 0.261
	auditFloorSuffix  = "-floor"
	classCold         = "cold"
)

type auditFacts struct {
	caps     map[string]float64
	clear    map[string]bool
	apart    map[string][]string
	quantity map[string]int
	shelf    []*boxpacker3.Box
}

func auditShelf(t *testing.T, random *rand.Rand, facts *auditFacts) []*boxpacker3.Box {
	t.Helper()

	boxes := make([]*boxpacker3.Box, 0, 6)

	for i := range 3 + random.Intn(3) {
		spec := boxpacker3.BoxSpec{
			ID:          fmt.Sprintf("box-%d", i),
			OuterWidth:  float64(220 + random.Intn(400)),
			OuterHeight: float64(220 + random.Intn(400)),
			OuterDepth:  float64(220 + random.Intn(400)),
			EmptyWeight: float64(random.Intn(600)),
			MaxWeight:   float64(4000 + random.Intn(16000)),
			Quantity:    1 + random.Intn(3),
		}
		spec.InnerWidth = spec.OuterWidth - float64(random.Intn(20))
		spec.InnerHeight = spec.OuterHeight - float64(random.Intn(20))
		spec.InnerDepth = spec.OuterDepth - float64(random.Intn(20))

		if random.Intn(4) == 0 {
			spec.Accepts = []string{classCold}
		}

		box, err := boxpacker3.NewBoxFromSpec(spec)
		require.NoError(t, err)

		facts.quantity[spec.ID] = spec.Quantity

		boxes = append(boxes, box)
	}

	return boxes
}

func auditItemSpec(random *rand.Rand, index int) boxpacker3.ItemSpec {
	classes := []string{"", classCold, "dry", "chem"}
	spec := boxpacker3.ItemSpec{
		ID:     fmt.Sprintf("item-%d", index),
		Width:  float64(30 + random.Intn(220)),
		Height: float64(30 + random.Intn(220)),
		Depth:  float64(30 + random.Intn(220)),
		Weight: float64(50 + random.Intn(1950)),
		Class:  classes[random.Intn(len(classes))],
	}

	spec = auditRotation(random, spec)

	if random.Intn(5) == 0 {
		spec.Group = fmt.Sprintf("group-%d", random.Intn(2))
	}

	if spec.Class == "chem" && random.Intn(2) == 0 {
		spec.SeparateFrom = []string{classCold}
	}

	switch random.Intn(6) {
	case 0:
		spec.MaxLoadOnTop = float64(300 + random.Intn(3000))
	case 1:
		spec.NothingOnTop = true
	}

	if random.Intn(8) == 0 {
		spec.ID += auditFloorSuffix
	}

	return spec
}

func auditRotation(random *rand.Rand, spec boxpacker3.ItemSpec) boxpacker3.ItemSpec {
	switch random.Intn(4) {
	case 0:
		spec.Rotation = boxpacker3.RotationKeepFlat
	case 1:
		spec.Rotation = boxpacker3.RotationNever
	case 2:
		spec.VerticalAxes = []boxpacker3.Axis{boxpacker3.Axis(random.Intn(3)), boxpacker3.Axis(random.Intn(3))}
	default:
		spec.Rotation = boxpacker3.RotationBestFit
	}

	return spec
}

func auditOrder(t *testing.T, random *rand.Rand, facts *auditFacts) []*boxpacker3.Item {
	t.Helper()

	items := make([]*boxpacker3.Item, 0, 30)

	for i := range 10 + random.Intn(16) {
		spec := auditItemSpec(random, i)

		item, err := boxpacker3.NewItemFromSpec(spec)
		require.NoError(t, err)

		facts.caps[spec.ID] = spec.MaxLoadOnTop
		facts.clear[spec.ID] = spec.NothingOnTop

		facts.apart[spec.ID] = spec.SeparateFrom

		items = append(items, item)
	}

	return items
}

func floorRule(_ *boxpacker3.Container, item *boxpacker3.Item, position boxpacker3.Pivot, _ boxpacker3.Dimension) bool {
	return staysOnTheFloor(item, position)
}

func staysOnTheFloor(item *boxpacker3.Item, position boxpacker3.Pivot) bool {
	return !strings.HasSuffix(item.ID(), auditFloorSuffix) || position[2] <= 1e-9
}

func isStable(box *boxpacker3.Box, dimension boxpacker3.Dimension) bool {
	depth := dimension[2]
	if depth <= 0 || math.Abs(depth-box.Depth()) <= box.Depth()*1e-9 {
		return true
	}

	return math.Atan(min(dimension[0], dimension[1])/depth) > auditStableTilt
}

func fitsDimension(box *boxpacker3.Box, dimension boxpacker3.Dimension) bool {
	limits := [3]float64{box.Width(), box.Height(), box.Depth()}

	for axis := range 3 {
		if dimension[axis] > limits[axis]+limits[axis]*1e-9 {
			return false
		}
	}

	return true
}

func couldHold(box *boxpacker3.Box, item *boxpacker3.Item) bool {
	if item.Weight()+box.EmptyWeight() > box.MaxWeight()+1e-9 {
		return false
	}

	if accepts := box.Accepts(); len(accepts) > 0 && !slices.Contains(accepts, item.Class()) {
		return false
	}

	for _, dimension := range item.Orientations() {
		if fitsDimension(box, dimension) {
			return true
		}
	}

	return false
}

func auditPlacement(t *testing.T, label string, packed boxpacker3.PackedBox, item boxpacker3.PackedItem, facts *auditFacts) {
	t.Helper()

	box := packed.Box
	id := item.ID()
	dimension := item.Dimension

	require.Contains(t, item.Item.Orientations(), dimension, "%s: %s lies in an orientation it does not allow", label, id)

	stableExists := false

	for _, orientation := range item.Item.Orientations() {
		if fitsDimension(box, orientation) && isStable(box, orientation) {
			stableExists = true
		}
	}

	if stableExists {
		require.True(t, isStable(box, dimension), "%s: %s stands unstably although it could lie flat", label, id)
	}

	require.GreaterOrEqual(t, supportRatioOf(packed, item), auditSupportRatio-1e-9,
		"%s: %s is under-supported", label, id)

	if facts.clear[id] {
		require.Zero(t, loadOnTopOf(packed, item), "%s: %s carries a load although it was marked clear", label, id)
	}

	if limit := facts.caps[id]; limit > 0 {
		require.LessOrEqual(t, loadOnTopOf(packed, item), limit+1e-9, "%s: %s carries more than its cap", label, id)
	}

	if accepts := box.Accepts(); len(accepts) > 0 {
		require.Contains(t, accepts, item.Item.Class(), "%s: %s is in a box that does not take it", label, id)
	}

	require.True(t, staysOnTheFloor(item.Item, item.Position), "%s: %s breaks the caller's rule", label, id)
}

func auditGroups(t *testing.T, label string, result *boxpacker3.Result) {
	t.Helper()

	holder := map[string]string{}
	unfit := map[string]bool{}

	for _, item := range result.Unpacked {
		if item.Item.Group() != "" {
			unfit[item.Item.Group()] = true
		}
	}

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			group := item.Item.Group()
			if group == "" {
				continue
			}

			require.False(t, unfit[group], "%s: group %s is split between a box and the leftovers", label, group)

			if previous, seen := holder[group]; seen {
				require.Equal(t, previous, box.ID(), "%s: group %s is split across boxes", label, group)
			}

			holder[group] = box.ID()
		}
	}
}

func judgeResult(
	t *testing.T, label string, offered []*boxpacker3.Item, result *boxpacker3.Result, err error, facts *auditFacts,
) {
	t.Helper()

	require.NoError(t, err, "%s: unexpected error", label)
	require.NotNil(t, result, "%s: no result", label)
	requireInvariants(t, offered, result)

	used := map[string]int{}

	for _, box := range result.Boxes {
		if len(box.Items) > 0 {
			used[box.ID()]++
		}

		for _, item := range box.Items {
			auditPlacement(t, label, box, item, facts)
		}
	}

	for id, count := range used {
		require.LessOrEqual(t, count, facts.quantity[id], "%s: more boxes %s than were in stock", label, id)
	}

	require.Empty(t, keptApartTogether(result, facts.apart), "%s: goods kept apart share a box", label)
	auditGroups(t, label, result)

	judgeUnpackable(t, label, result, facts)
}

func judgeUnpackable(t *testing.T, label string, result *boxpacker3.Result, facts *auditFacts) {
	t.Helper()

	for _, item := range result.Unpacked {
		holdable := false
		for _, box := range facts.shelf {
			holdable = holdable || couldHold(box, item.Item)
		}

		if item.Reason.Structural() {
			require.False(t, holdable, "%s: %s is reported %s but a box could hold it", label, item.ID(), item.Reason)
		} else if item.Reason == boxpacker3.ReasonNoRoom {
			require.True(t, holdable, "%s: %s fits no box yet is reported as merely out of room", label, item.ID())
		}
	}
}

func keptApartTogether(result *boxpacker3.Result, apart map[string][]string) []string {
	mixed := make([]string, 0)

	for _, box := range result.Boxes {
		items := box.Items

		for i, one := range items {
			for _, other := range items[i+1:] {
				if slices.Contains(apart[one.ID()], other.Item.Class()) ||
					slices.Contains(apart[other.ID()], one.Item.Class()) {
					mixed = append(mixed, fmt.Sprintf("%s holds %s with %s", box.ID(), one.ID(), other.ID()))
				}
			}
		}
	}

	return mixed
}

func auditAlgorithms(t *testing.T) map[string]boxpacker3.Algorithm {
	t.Helper()

	algorithms := map[string]boxpacker3.Algorithm{}

	for _, order := range boxpacker3.ItemOrders() {
		for _, selection := range boxpacker3.BoxSelections() {
			strategy := boxpacker3.NewGreedy(order, selection)
			algorithms[strategy.Name()] = strategy
		}
	}

	for _, name := range boxpacker3.GoalNames() {
		algorithm, ok := boxpacker3.BestForName(name)
		require.True(t, ok)

		algorithms[name] = algorithm
	}

	search, err := boxpacker3.NewSearch(96, 3)
	require.NoError(t, err)

	algorithms[search.Name()] = search

	return algorithms
}

func TestAudit_EveryAlgorithmUnderEveryRule(t *testing.T) {
	t.Parallel()

	for seed := range auditSeeds {
		t.Run(fmt.Sprintf("seed-%d", seed), func(t *testing.T) {
			t.Parallel()

			random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec
			facts := &auditFacts{
				caps: map[string]float64{}, clear: map[string]bool{},
				apart: map[string][]string{}, quantity: map[string]int{}, shelf: nil,
			}
			boxes := auditShelf(t, random, facts)
			items := auditOrder(t, random, facts)
			facts.shelf = boxes

			for name, algorithm := range auditAlgorithms(t) {
				rules := boxpacker3.WithRules(boxpacker3.Rules{
					MinSupportRatio: auditSupportRatio,
					Placement:       []boxpacker3.PlacementRule{boxpacker3.PlacementRuleFunc(floorRule)},
				})
				packer := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(algorithm), rules)

				result, err := packer.Pack(context.Background(), boxes, items)
				judgeResult(t, fmt.Sprintf("seed %d %s", seed, name), items, result, err, facts)

				alone := *facts
				alone.shelf = boxes[:1]

				singlePacker := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(boxpacker3.SingleContainer(algorithm)), rules)

				single, err := singlePacker.Pack(context.Background(), boxes[:1], items)
				judgeResult(t, fmt.Sprintf("seed %d %s single", seed, name), items, single, err, &alone)
			}
		})
	}
}
