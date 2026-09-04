package boxpacker3_test

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

type turnable struct {
	dims [3]float64
	mode boxpacker3.Rotation
}

func turnableOrder(t *testing.T, random *rand.Rand, count int) ([]*boxpacker3.Item, map[string]turnable) {
	t.Helper()

	declared := make(map[string]turnable, count)
	items := make([]*boxpacker3.Item, 0, count)

	for i := range count {
		id := fmt.Sprintf("item-%d", i)
		dims := [3]float64{
			float64(20 + random.Intn(70)),
			float64(20 + random.Intn(70)),
			float64(20 + random.Intn(70)),
		}
		mode := boxpacker3.Rotations()[random.Intn(3)]

		built, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
			ID: id, Width: dims[0], Height: dims[1], Depth: dims[2],
			Weight: 100, Rotation: mode,
		})
		if err != nil {
			t.Fatal(err)
		}

		declared[id] = turnable{dims: dims, mode: mode}

		items = append(items, built)
	}

	return items, declared
}

func checkRotation(t *testing.T, tally map[string]int, where string,
	item boxpacker3.PackedItem, was turnable,
) {
	t.Helper()

	got := item.Dimension

	note := func(kind, detail string) {
		tally[kind]++

		if tally[kind] <= 3 {
			t.Errorf("%s: %s %s", where, item.ID(), detail)
		}
	}

	switch was.mode {
	case boxpacker3.RotationNever:
		if got != was.dims {
			note("never-turned", fmt.Sprintf("declared %v shipped %v", was.dims, got))
		}
	case boxpacker3.RotationKeepFlat:
		if math.Abs(got[2]-was.dims[2]) > 1e-9 {
			note("flat-tipped",
				fmt.Sprintf("kept flat but depth %.0f became %.0f", was.dims[2], got[2]))
		}
	case boxpacker3.RotationBestFit:
	}

	if math.Abs(got[0]*got[1]*got[2]-was.dims[0]*was.dims[1]*was.dims[2]) > 1e-6 {
		note("resized", "changed volume")
	}
}

func TestAudit_RotationModesHold(t *testing.T) {
	t.Parallel()

	tally := map[string]int{}

	for seed := range 300 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 2+random.Intn(2), 120, 160, 100000)
		items, declared := turnableOrder(t, random, 6+random.Intn(14))

		for _, rule := range boxpacker3.BoxSelections() {
			result, err := rulePacker(boxpacker3.OrderDecreasing, rule).
				Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			where := fmt.Sprintf("seed %d rule %s", seed, rule)

			for _, box := range result.Boxes {
				for _, item := range box.Items {
					checkRotation(t, tally, where, item, declared[item.ID()])
				}
			}
		}
	}

	t.Logf("rotation violations: %v", tally)
}

func axisOrder(t *testing.T, random *rand.Rand, count int) ([]*boxpacker3.Item, map[string][]float64) {
	t.Helper()

	allowed := make(map[string][]float64, count)
	items := make([]*boxpacker3.Item, 0, count)

	for i := range count {
		id := fmt.Sprintf("item-%d", i)
		whd := [3]float64{
			float64(20 + random.Intn(60)),
			float64(20 + random.Intn(60)),
			float64(20 + random.Intn(60)),
		}

		axes := []boxpacker3.Axis{boxpacker3.WidthAxis, boxpacker3.HeightAxis, boxpacker3.DepthAxis}

		random.Shuffle(3, func(a, b int) { axes[a], axes[b] = axes[b], axes[a] })
		axes = axes[:1+random.Intn(2)]

		sizes := make([]float64, 0, len(axes))
		for _, axis := range axes {
			sizes = append(sizes, whd[axis])
		}

		built, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
			ID: id, Width: whd[0], Height: whd[1], Depth: whd[2],
			Weight: 100, VerticalAxes: axes,
		})
		if err != nil {
			t.Fatal(err)
		}

		allowed[id] = sizes

		items = append(items, built)
	}

	return items, allowed
}

func TestAudit_VerticalAxesHold(t *testing.T) {
	t.Parallel()

	breaches := 0

	for seed := range 300 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 2+random.Intn(2), 130, 150, 100000)
		items, allowed := axisOrder(t, random, 5+random.Intn(12))

		for _, rule := range boxpacker3.BoxSelections() {
			result, err := rulePacker(boxpacker3.OrderDecreasing, rule).
				Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			for _, box := range result.Boxes {
				for _, item := range box.Items {
					if standsOnAnAllowedEdge(item, allowed[item.ID()]) {
						continue
					}

					breaches++

					if breaches <= 3 {
						t.Errorf("seed %d rule %s: %s stands %.0f high, allowed %v",
							seed, rule, item.ID(), item.Dimension[2],
							allowed[item.ID()])
					}
				}
			}
		}
	}

	t.Logf("vertical-axis violations: %d", breaches)
}

func standsOnAnAllowedEdge(item boxpacker3.PackedItem, allowed []float64) bool {
	shipped := item.Dimension[2]

	for _, size := range allowed {
		if math.Abs(size-shipped) < 1e-9 {
			return true
		}
	}

	return false
}

func TestAudit_UnusableVerticalAxesRejected(t *testing.T) {
	t.Parallel()

	item, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
		ID: "odd", Width: 50, Height: 50, Depth: 50, Weight: 10,
		VerticalAxes: []boxpacker3.Axis{boxpacker3.Axis(9)},
	})
	if err == nil {
		t.Errorf("an axis outside width/height/depth was accepted, leaving %q unpackable",
			item.ID())
	}

	if !errors.Is(err, boxpacker3.ErrInvalidAxis) {
		t.Errorf("want ErrInvalidAxis, got %v", err)
	}
}

func TestAudit_BoxSpecValidation(t *testing.T) {
	t.Parallel()

	base := boxpacker3.BoxSpec{
		ID: "b", OuterWidth: 100, OuterHeight: 100, OuterDepth: 100,
		InnerWidth: 90, InnerHeight: 90, InnerDepth: 90,
		EmptyWeight: 10, MaxWeight: 1000, Quantity: 1,
	}

	cases := map[string]struct {
		spoil func(boxpacker3.BoxSpec) boxpacker3.BoxSpec
		want  error
	}{
		"inner wider than outer": {func(s boxpacker3.BoxSpec) boxpacker3.BoxSpec {
			s.InnerWidth = 200

			return s
		}, boxpacker3.ErrInnerExceedsOuter},
		"zero inner depth": {func(s boxpacker3.BoxSpec) boxpacker3.BoxSpec {
			s.InnerDepth, s.OuterDepth = 0, 0

			return s
		}, boxpacker3.ErrInvalidDimension},
		"negative tare": {func(s boxpacker3.BoxSpec) boxpacker3.BoxSpec {
			s.EmptyWeight = -1

			return s
		}, boxpacker3.ErrInvalidWeight},
		"tare over the limit": {func(s boxpacker3.BoxSpec) boxpacker3.BoxSpec {
			s.EmptyWeight = 2000

			return s
		}, boxpacker3.ErrInvalidWeight},
		"negative quantity": {func(s boxpacker3.BoxSpec) boxpacker3.BoxSpec {
			s.Quantity = -3

			return s
		}, boxpacker3.ErrInvalidQuantity},
		"negative max weight": {func(s boxpacker3.BoxSpec) boxpacker3.BoxSpec {
			s.MaxWeight = -5

			return s
		}, boxpacker3.ErrInvalidWeight},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := boxpacker3.NewBoxFromSpec(test.spoil(base))
			if !errors.Is(err, test.want) {
				t.Errorf("want %v, got %v", test.want, err)
			}
		})
	}
}

func TestAudit_SearchValidation(t *testing.T) {
	t.Parallel()

	for _, pair := range [][2]int{{0, 1}, {1, 0}, {-1, 2}, {2, -1}} {
		_, err := boxpacker3.NewSearch(pair[0], pair[1])
		if !errors.Is(err, boxpacker3.ErrInvalidSetting) {
			t.Errorf("search(%d,%d) gave %v", pair[0], pair[1], err)
		}
	}

	search, err := boxpacker3.NewSearch(64, 2)
	if err != nil {
		t.Fatalf("a sane search was rejected: %v", err)
	}

	if search.Name() == "" {
		t.Error("the search strategy has no name")
	}
}

func TestAudit_EmptyAndDegenerateOrders(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("b", 100, 100, 100, 1000)
	item := boxpacker3.NewItem("i", 50, 50, 50, 100)

	cases := map[string]struct {
		boxes []*boxpacker3.Box
		items []*boxpacker3.Item
	}{
		"no boxes and no items": {nil, nil},
		"boxes but no items":    {[]*boxpacker3.Box{box}, nil},
		"items but no boxes":    {nil, []*boxpacker3.Item{item}},
		"a nil box in the list": {[]*boxpacker3.Box{nil, box}, []*boxpacker3.Item{item}},
		"a nil item in the list": {
			[]*boxpacker3.Box{box}, []*boxpacker3.Item{nil, item},
		},
	}

	for name, test := range cases {
		for _, rule := range boxpacker3.BoxSelections() {
			t.Run(fmt.Sprintf("%s/%s", name, rule), func(t *testing.T) {
				t.Parallel()

				result, err := rulePacker(boxpacker3.OrderDecreasing, rule).
					Pack(t.Context(), test.boxes, test.items)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if result == nil {
					t.Fatal("no result")
				}
			})
		}
	}
}

func TestAudit_SingleContainerReportsUnpackable(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("crate", 100, 100, 100, 5000)
	items := []*boxpacker3.Item{
		boxpacker3.NewItem("fits", 50, 50, 50, 100),
		boxpacker3.NewItem("too-big", 500, 50, 50, 100),
	}

	packer := boxpacker3.NewPacker()

	viaBoxes, err := packer.Pack(t.Context(), []*boxpacker3.Box{box}, items)
	if err != nil || len(viaBoxes.Unpackable()) != 1 {
		t.Fatalf("Pack did not report the unpackable item: %v %v", viaBoxes.Unpacked, err)
	}

	single := singlePacker()

	viaSingle, err := single.Pack(t.Context(), []*boxpacker3.Box{box}, items)
	if err != nil || len(viaSingle.Unpackable()) != 1 {
		t.Errorf("the single-container search swallowed the unpackable item: %v %v", viaSingle.Unpacked, err)
	}
}

func TestAudit_GoalNamesResolve(t *testing.T) {
	t.Parallel()

	for _, name := range boxpacker3.GoalNames() {
		if _, ok := boxpacker3.GoalByName(name); !ok {
			t.Errorf("%s is listed but does not resolve", name)
		}
	}

	if _, ok := boxpacker3.GoalByName("no-such-goal"); ok {
		t.Error("an unknown goal resolved")
	}
}
