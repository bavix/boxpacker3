package boxpacker3_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

const (
	classFood      = "food"
	classChemicals = "chemicals"
	oddID          = "odd"
)

func classedItem(t *testing.T, id string, side, weight float64, class string, away ...string) *boxpacker3.Item {
	t.Helper()

	built, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
		ID: id, Width: side, Height: side, Depth: side, Weight: weight,
		Class: class, SeparateFrom: away,
	})
	if err != nil {
		t.Fatal(err)
	}

	return built
}

func mixedBoxes(result *boxpacker3.Result, apart map[string]map[string]bool) []string {
	mixed := make([]string, 0)

	for _, box := range result.Boxes {
		for _, one := range box.Items {
			for _, other := range box.Items {
				if one.ID() == other.ID() {
					continue
				}

				if apart[one.Item.Class()][other.Item.Class()] {
					mixed = append(mixed, fmt.Sprintf("%s holds %s with %s",
						box.ID(), one.ID(), other.ID()))
				}
			}
		}
	}

	return mixed
}

func TestSeparation_GoodsKeptApartNeverShareABox(t *testing.T) {
	t.Parallel()

	apart := map[string]map[string]bool{
		classFood:      {classChemicals: true},
		classChemicals: {classFood: true},
	}

	for seed := range 200 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 2+random.Intn(3), 200, 120, 1e9)
		items := make([]*boxpacker3.Item, 0, 16)

		for i := range 6 + random.Intn(10) {
			side := float64(40 + random.Intn(40))
			weight := float64(100 + random.Intn(500))

			if i%2 == 0 {
				items = append(items, classedItem(t,
					fmt.Sprintf("food-%d", i), side, weight, classFood, classChemicals))

				continue
			}

			items = append(items, classedItem(t,
				fmt.Sprintf("drum-%d", i), side, weight, classChemicals))
		}

		for _, rule := range boxpacker3.BoxSelections() {
			result, err := rulePacker(boxpacker3.OrderDecreasing, rule).
				Pack(t.Context(), boxes, items)
			if err != nil {
				t.Fatal(err)
			}

			for _, mixed := range mixedBoxes(result, apart) {
				t.Fatalf("seed %d rule %s: %s", seed, rule, mixed)
			}
		}
	}
}

func TestSeparation_ForcesASecondBox(t *testing.T) {
	t.Parallel()

	roomy := []*boxpacker3.Box{
		boxpacker3.NewBox("hold-1", 400, 400, 400, 1e9),
		boxpacker3.NewBox("hold-2", 400, 400, 400, 1e9),
	}

	items := []*boxpacker3.Item{
		classedItem(t, "flour", 100, 1000, classFood, classChemicals),
		classedItem(t, "bleach", 100, 1000, classChemicals),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), roomy, items)
	if err != nil {
		t.Fatal(err)
	}

	if used := usedBoxCountOf(result); used != 2 {
		t.Fatalf("goods that must stay apart went into %d boxes", used)
	}
}

func TestSeparation_UnlabelledGoodsAreUnaffected(t *testing.T) {
	t.Parallel()

	for seed := range 40 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 3, 200, 120, 1e9)
		plain := make([]*boxpacker3.Item, 0, 12)
		labelled := make([]*boxpacker3.Item, 0, 12)

		for i := range 12 {
			id := fmt.Sprintf("item-%d", i)
			side := float64(40 + random.Intn(50))
			weight := float64(100 + random.Intn(700))

			plain = append(plain, boxpacker3.NewItem(id, side, side, side, weight))
			labelled = append(labelled, classedItem(t, id, side, weight, "general"))
		}

		packer := boxpacker3.NewPacker()

		if fingerprint(packed(t, packer, boxes, plain)) != fingerprint(packed(t, packer, boxes, labelled)) {
			t.Fatalf("seed %d: a label nobody avoids changed the packing", seed)
		}
	}
}

func TestSeparation_OnePerBox(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("hold-1", 400, 400, 400, 1e9),
		boxpacker3.NewBox("hold-2", 400, 400, 400, 1e9),
		boxpacker3.NewBox("hold-3", 400, 400, 400, 1e9),
	}

	items := []*boxpacker3.Item{
		classedItem(t, "engine-1", 100, 1000, "engine", "engine"),
		classedItem(t, "engine-2", 100, 1000, "engine", "engine"),
		classedItem(t, "engine-3", 100, 1000, "engine", "engine"),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	for _, box := range result.Boxes {
		if len(box.Items) > 1 {
			t.Errorf("%s holds %d engines", box.ID(), len(box.Items))
		}
	}
}

func TestSeparation_EmptyClassIsRejected(t *testing.T) {
	t.Parallel()

	_, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
		ID: oddID, Width: 10, Height: 10, Depth: 10, Weight: 1,
		SeparateFrom: []string{""},
	})
	if err == nil {
		t.Error("an empty class was accepted")
	}
}

func TestSeparation_ContradictsAGroup(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("hold-1", 400, 400, 400, 1e9),
		boxpacker3.NewBox("hold-2", 400, 400, 400, 1e9),
	}

	together := func(id, class string, away ...string) *boxpacker3.Item {
		built, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
			ID: id, Width: 100, Height: 100, Depth: 100, Weight: 1000,
			Group: "order-1", Class: class, SeparateFrom: away,
		})
		if err != nil {
			t.Fatal(err)
		}

		return built
	}

	items := []*boxpacker3.Item{
		together("flour", classFood, classChemicals),
		together("bleach", classChemicals),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Unpacked) != 2 {
		t.Errorf("the contradiction left %d items unpacked, not both", len(result.Unpacked))
	}

	if used := usedBoxCountOf(result); used != 0 {
		t.Errorf("%d boxes hold part of a group that cannot ship", used)
	}
}

func dedicatedBox(t *testing.T, id string, accepts ...string) *boxpacker3.Box {
	t.Helper()

	const side = 400

	built, err := boxpacker3.NewBoxFromSpec(boxpacker3.BoxSpec{
		ID: id, OuterWidth: side, OuterHeight: side, OuterDepth: side,
		MaxWeight: 1e9, Accepts: accepts,
	})
	if err != nil {
		t.Fatal(err)
	}

	return built
}

func TestDedicatedBox_TakesOnlyWhatItNames(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		dedicatedBox(t, "reefer", classFood),
		boxpacker3.NewBox("general", 400, 400, 400, 1e9),
	}

	items := []*boxpacker3.Item{
		classedItem(t, "flour", 100, 1000, classFood),
		classedItem(t, "bleach", 100, 1000, classChemicals),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	for _, box := range result.Boxes {
		for _, packed := range box.Items {
			if box.ID() == "reefer" && packed.Item.Class() != classFood {
				t.Errorf("the reefer took %s, which is %q", packed.ID(), packed.Item.Class())
			}
		}
	}
}

func TestDedicatedBox_RefusesUnlabelledGoods(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{dedicatedBox(t, "reefer", classFood)}
	items := []*boxpacker3.Item{boxpacker3.NewItem("crate", 100, 100, 100, 1000)}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	if err != nil || len(result.Unpackable()) != 1 {
		t.Fatal("goods no box will take were not reported as unpackable")
	}

	if len(result.Unpacked) != 1 {
		t.Errorf("%d items were left behind, not one", len(result.Unpacked))
	}

	for _, box := range result.Boxes {
		if len(box.Items) != 0 {
			t.Errorf("%s took goods it does not name", box.ID())
		}
	}
}

func TestDedicatedBox_NamingNothingTakesAnything(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{dedicatedBox(t, "general")}
	items := []*boxpacker3.Item{
		classedItem(t, "flour", 100, 1000, classFood),
		boxpacker3.NewItem("crate", 100, 100, 100, 1000),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Boxes[0].Items) != 2 {
		t.Errorf("a box that names nothing took %d of two items",
			len(result.Boxes[0].Items))
	}
}

func TestDedicatedBox_EmptyClassIsRejected(t *testing.T) {
	t.Parallel()

	_, err := boxpacker3.NewBoxFromSpec(boxpacker3.BoxSpec{
		ID: oddID, OuterWidth: 10, OuterHeight: 10, OuterDepth: 10, MaxWeight: 10,
		Accepts: []string{""},
	})
	if err == nil {
		t.Error("a box accepting an empty class was allowed")
	}
}

func TestDedicatedBox_DoesNotAttractTheGoodsItTakes(t *testing.T) {
	t.Parallel()

	boxes := []*boxpacker3.Box{
		dedicatedBox(t, "reefer", classFood),
		boxpacker3.NewBox("general", 400, 400, 400, 1e9),
	}

	items := []*boxpacker3.Item{
		classedItem(t, "flour", 100, 1000, classFood),
		classedItem(t, "bleach", 100, 1000, classChemicals, classFood),
		classedItem(t, "crate", 100, 1000, "hardware"),
	}

	result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
	if err != nil {
		t.Fatal(err)
	}

	for _, box := range result.Boxes {
		food, chemicals := false, false

		for _, packed := range box.Items {
			food = food || packed.Item.Class() == classFood
			chemicals = chemicals || packed.Item.Class() == classChemicals

			if box.ID() == "reefer" && packed.Item.Class() != classFood {
				t.Errorf("the reefer took %s", packed.ID())
			}
		}

		if food && chemicals {
			t.Errorf("%s holds food with chemicals", box.ID())
		}
	}

	t.Logf("goods left behind: %d", len(result.Unpacked))
}

func TestSeparation_SurvivesARepack(t *testing.T) {
	t.Parallel()

	apart := map[string]map[string]bool{
		classFood:      {classChemicals: true},
		classChemicals: {classFood: true},
	}

	for seed := range 150 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := []*boxpacker3.Box{boxpacker3.NewBox("hold", 260, 260, 260, 1e9)}
		items := make([]*boxpacker3.Item, 0, 14)

		for i := range 10 + random.Intn(4) {
			side := float64(60 + random.Intn(60))
			weight := float64(100 + random.Intn(400))
			id := fmt.Sprintf("item-%d", i)

			if i%3 == 0 {
				items = append(items, classedItem(t, id, side, weight, classChemicals))

				continue
			}

			items = append(items, classedItem(t, id, side, weight, classFood, classChemicals))
		}

		result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		for _, mixed := range mixedBoxes(result, apart) {
			t.Fatalf("seed %d: %s", seed, mixed)
		}
	}
}

func TestSeparation_SurvivesWeightBalancing(t *testing.T) {
	t.Parallel()

	apart := map[string]map[string]bool{
		classFood:      {classChemicals: true},
		classChemicals: {classFood: true},
	}

	for seed := range 150 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		boxes := shelfOf(random, 4, 200, 140, 1e9)
		items := make([]*boxpacker3.Item, 0, 16)

		for i := range 10 + random.Intn(6) {
			id := fmt.Sprintf("item-%d", i)
			side := float64(50 + random.Intn(50))

			weight := float64(100 + random.Intn(4000))

			if i%3 == 0 {
				items = append(items, classedItem(t, id, side, weight, classChemicals))

				continue
			}

			items = append(items, classedItem(t, id, side, weight, classFood, classChemicals))
		}

		result, err := boxpacker3.NewPacker(
			boxpacker3.WithFinishers(boxpacker3.BalanceWeight{MaxBoxes: 12})).
			Pack(t.Context(), boxes, items)
		if err != nil {
			t.Fatal(err)
		}

		for _, mixed := range mixedBoxes(result, apart) {
			t.Fatalf("seed %d: %s", seed, mixed)
		}
	}
}

func TestSeparation_HoldsInASingleContainer(t *testing.T) {
	t.Parallel()

	apart := map[string]map[string]bool{
		classFood:      {classChemicals: true},
		classChemicals: {classFood: true},
	}

	for seed := range 60 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		box := boxpacker3.NewBox("hold", 300, 300, 300, 1e9)
		items := make([]*boxpacker3.Item, 0, 12)

		for i := range 8 + random.Intn(4) {
			id := fmt.Sprintf("item-%d", i)
			side := float64(60 + random.Intn(60))
			weight := float64(100 + random.Intn(400))

			if i%3 == 0 {
				items = append(items, classedItem(t, id, side, weight, classChemicals))

				continue
			}

			items = append(items, classedItem(t, id, side, weight, classFood, classChemicals))
		}

		result, err := singlePacker().Pack(t.Context(), []*boxpacker3.Box{box}, items)
		if err != nil {
			t.Fatal(err)
		}

		for _, mixed := range mixedBoxes(result, apart) {
			t.Fatalf("seed %d: %s", seed, mixed)
		}
	}
}
