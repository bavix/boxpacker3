package boxpacker3_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

func fragileItem(t *testing.T, id string, side, weight, carries float64, keepClear bool) *boxpacker3.Item {
	t.Helper()

	built, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
		ID: id, Width: side, Height: side, Depth: side, Weight: weight,
		MaxLoadOnTop: carries, NothingOnTop: keepClear,
	})
	if err != nil {
		t.Fatal(err)
	}

	return built
}

func loadOnTopOf(box boxpacker3.PackedBox, below boxpacker3.PackedItem) float64 {
	load := 0.0
	belowAt, belowSize := below.Position, below.Dimension
	top := belowAt[2] + belowSize[2]

	for _, above := range box.Items {
		if above.Item == below.Item && above.Index == below.Index {
			continue
		}

		at, size := above.Position, above.Dimension
		if at[2]+1e-9 < top {
			continue
		}

		if overlapSpan(at[0], size[0], belowAt[0], belowSize[0]) <= 0 ||
			overlapSpan(at[1], size[1], belowAt[1], belowSize[1]) <= 0 {
			continue
		}

		load += above.Weight()
	}

	return load
}

func TestLoadBearing_NothingRestsOnAClearItem(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("crate", 300, 300, 300, 1e9)

	items := []*boxpacker3.Item{
		fragileItem(t, "glass", 100, 500, 0, true),
		fragileItem(t, "brick-1", 100, 4000, 0, false),
		fragileItem(t, "brick-2", 100, 4000, 0, false),
		fragileItem(t, "brick-3", 100, 4000, 0, false),
	}

	for _, rule := range boxpacker3.BoxSelections() {
		result, err := rulePacker(boxpacker3.OrderDecreasing, rule).
			Pack(t.Context(), []*boxpacker3.Box{box}, items)
		if err != nil {
			t.Fatal(err)
		}

		for _, packed := range result.Boxes {
			for _, item := range packed.Items {
				if item.ID() != "glass" {
					continue
				}

				if load := loadOnTopOf(packed, item); load > 0 {
					t.Errorf("rule %s: %.0f rests on the item that must stay clear", rule, load)
				}
			}
		}
	}
}

func randomCargo(t *testing.T, random *rand.Rand) ([]*boxpacker3.Item, map[string]float64) {
	t.Helper()

	items := make([]*boxpacker3.Item, 0, 20)
	caps := map[string]float64{}

	for i := range 8 + random.Intn(12) {
		id := fmt.Sprintf("item-%d", i)
		carries := 0.0

		if random.Intn(3) == 0 {
			carries = float64(200 + random.Intn(1500))
		}

		caps[id] = carries
		items = append(items, fragileItem(t, id,
			float64(40+random.Intn(50)), float64(100+random.Intn(700)), carries, false))
	}

	return items, caps
}

func overloaded(result *boxpacker3.Result, caps map[string]float64) []string {
	broken := make([]string, 0)

	for _, box := range result.Boxes {
		for _, item := range box.Items {
			limit := caps[item.ID()]
			if limit <= 0 {
				continue
			}

			if load := loadOnTopOf(box, item); load > limit+1e-9 {
				broken = append(broken,
					fmt.Sprintf("%s carries %.0f of %.0f", item.ID(), load, limit))
			}
		}
	}

	return broken
}

func TestLoadBearing_CapIsNeverExceeded(t *testing.T) {
	t.Parallel()

	for seed := range 300 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 2+random.Intn(2), 200, 120, 1e9)
		items, caps := randomCargo(t, random)

		for _, rule := range boxpacker3.BoxSelections() {
			result, err := rulePacker(boxpacker3.OrderDecreasing, rule).
				Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			for _, breach := range overloaded(result, caps) {
				t.Fatalf("seed %d rule %s: %s", seed, rule, breach)
			}
		}
	}
}

func TestLoadBearing_CapIsValidated(t *testing.T) {
	t.Parallel()

	_, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
		ID: oddID, Width: 10, Height: 10, Depth: 10, Weight: 1, MaxLoadOnTop: -5,
	})
	if err == nil {
		t.Error("a negative load cap was accepted")
	}
}

func TestLoadBearing_AGenerousCapChangesNothing(t *testing.T) {
	t.Parallel()

	for seed := range 40 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 3, 200, 120, 1e9)
		plain := make([]*boxpacker3.Item, 0, 12)
		generous := make([]*boxpacker3.Item, 0, 12)

		for i := range 12 {
			id := fmt.Sprintf("item-%d", i)
			side := float64(40 + random.Intn(50))
			weight := float64(100 + random.Intn(700))

			plain = append(plain, fragileItem(t, id, side, weight, 0, false))
			generous = append(generous, fragileItem(t, id, side, weight, 1e9, false))
		}

		packer := boxpacker3.NewPacker()

		if fingerprint(packed(t, packer, boxes, plain)) != fingerprint(packed(t, packer, boxes, generous)) {
			t.Fatalf("seed %d: an unreachable cap changed the packing", seed)
		}
	}
}

func TestLoadBearing_ACapStopsTheStack(t *testing.T) {
	t.Parallel()

	column := boxpacker3.NewBox("column", 100, 100, 400, 1e9)

	items := make([]*boxpacker3.Item, 0, 4)
	for i := range 4 {
		items = append(items, fragileItem(t, fmt.Sprintf("crate-%d", i), 100, 1000, 1500, false))
	}

	result, err := boxpacker3.NewPacker().
		Pack(t.Context(), []*boxpacker3.Box{column}, items)
	if err != nil {
		t.Fatal(err)
	}

	packed := 0
	for _, box := range result.Boxes {
		packed += len(box.Items)
	}

	if packed != 2 {
		t.Errorf("the column holds %d crates where the load caps allow two", packed)
	}

	free, err := boxpacker3.NewPacker().Pack(t.Context(),
		[]*boxpacker3.Box{boxpacker3.NewBox("column", 100, 100, 400, 1e9)},
		[]*boxpacker3.Item{
			boxpacker3.NewItem("a", 100, 100, 100, 1000),
			boxpacker3.NewItem("b", 100, 100, 100, 1000),
			boxpacker3.NewItem("c", 100, 100, 100, 1000),
			boxpacker3.NewItem("d", 100, 100, 100, 1000),
		})
	if err != nil {
		t.Fatalf("without caps the same column should take all four: %v", err)
	}

	if len(free.Boxes[0].Items) != 4 {
		t.Errorf("without caps the column took %d crates, not four", len(free.Boxes[0].Items))
	}
}

func BenchmarkLoadBearing_Capped(b *testing.B) {
	for _, capped := range []bool{false, true} {
		name := "plain"
		if capped {
			name = "capped"
		}

		b.Run(name, func(b *testing.B) {
			box := boxpacker3.NewBox("crate", 600, 600, 600, 1e9)
			items := make([]*boxpacker3.Item, 0, 60)

			for i := range 60 {
				limit := 0.0
				if capped {
					limit = 1e6
				}

				built, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
					ID: fmt.Sprintf("item-%d", i), Width: 100, Height: 100, Depth: 100,
					Weight: 100, MaxLoadOnTop: limit,
				})
				if err != nil {
					b.Fatal(err)
				}

				items = append(items, built)
			}

			b.ResetTimer()

			for range b.N {
				_, _ = boxpacker3.NewPacker().Pack(b.Context(), []*boxpacker3.Box{box}, items)
			}
		})
	}
}

func TestLoadBearing_SurvivesARepack(t *testing.T) {
	t.Parallel()

	for seed := range 150 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := []*boxpacker3.Box{boxpacker3.NewBox("hold", 260, 260, 260, 1e9)}
		items := make([]*boxpacker3.Item, 0, 14)
		caps := map[string]float64{}

		for i := range 10 + random.Intn(4) {
			id := fmt.Sprintf("item-%d", i)
			limit := 0.0

			if i%2 == 0 {
				limit = float64(150 + random.Intn(400))
			}

			caps[id] = limit
			items = append(items, fragileItem(t, id,
				float64(60+random.Intn(60)), float64(100+random.Intn(400)), limit, false))
		}

		result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		for _, breach := range overloaded(result, caps) {
			t.Fatalf("seed %d: %s", seed, breach)
		}
	}
}
