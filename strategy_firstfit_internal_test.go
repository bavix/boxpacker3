package boxpacker3

import (
	"fmt"
	"math/rand"
	"testing"
)

func itemMajorFirstFit(t *testing.T, boxes []*Container, items []*piece) []string {
	t.Helper()

	packed := make([]string, 0, len(items))

	for _, item := range items {
		placed := false

		for _, box := range boxes {
			if packItem(box, item, true) {
				packed = append(packed, box.id)
				placed = true

				break
			}
		}

		if !placed {
			packed = append(packed, "-")
		}
	}

	return packed
}

type walkCost struct {
	containers int
	left       int
}

func costOfDefinition(t *testing.T, boxes []*Container, ordered []*piece) (walkCost, []string) {
	t.Helper()

	placed := itemMajorFirstFit(t, boxes, ordered)
	used := map[string]bool{}
	cost := walkCost{containers: 0, left: 0}

	for _, id := range placed {
		if id == "-" {
			cost.left++

			continue
		}

		used[id] = true
	}

	cost.containers = len(used)

	return cost, placed
}

func TestFirstFit_IsNeverWorseThanTheDefinition(t *testing.T) {
	t.Parallel()

	differences := 0
	shipped := walkCost{containers: 0, left: 0}
	definition := walkCost{containers: 0, left: 0}

	for seed := range 400 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		inputBoxes, inputItems := plainOrder(random)

		byDefinition, _ := prepareData(inputBoxes, inputItems, problemOf(inputBoxes, inputItems))
		ordered := sortItems(compact(cloneSlice(inputItems)), OrderDecreasing)

		cost, want := costOfDefinition(t, byDefinition, ordered)
		definition.containers += cost.containers
		definition.left += cost.left

		run, err := runFirstFit(t.Context(), inputBoxes,
			sortItems(compact(cloneSlice(inputItems)), OrderDecreasing), problemOf(inputBoxes, inputItems))
		if err != nil {
			t.Fatal(err)
		}

		if differsFrom(run, ordered, want) {
			differences++
		}

		shipped.containers += filledBoxes(run)
		shipped.left += len(run.unfit)
	}

	t.Logf("orders where the walk and the definition disagree: %d of 400", differences)
	t.Logf("containers: shipped %d, definition %d", shipped.containers, definition.containers)
	t.Logf("left behind: shipped %d, definition %d", shipped.left, definition.left)

	if shipped.containers > definition.containers {
		t.Errorf("the shipped rule used %d containers where the plain definition used %d",
			shipped.containers, definition.containers)
	}

	if shipped.left > definition.left {
		t.Errorf("the shipped rule left %d behind where the plain definition left %d",
			shipped.left, definition.left)
	}
}

func plainOrder(random *rand.Rand) ([]*Box, []*piece) {
	boxes := make([]*Box, 0, 5)

	for i := range 2 + random.Intn(4) {
		side := 100 + float64(random.Intn(180))
		boxes = append(boxes, NewBox(
			fmt.Sprintf("box-%d", i), side, side*0.8, side*0.7, 1e9))
	}

	items := make([]*piece, 0, 20)

	for i := range 6 + random.Intn(14) {
		items = append(items, testPiece(
			fmt.Sprintf("item-%d", i),
			20+float64(random.Intn(90)),
			20+float64(random.Intn(90)),
			20+float64(random.Intn(90)), 1))
	}

	return boxes, items
}

func differsFrom(run *Packing, ordered []*piece, want []string) bool {
	where := map[string]string{}

	for _, box := range run.boxes {
		for _, item := range box.items {
			where[item.ID()] = box.ID()
		}
	}

	for _, item := range run.unfit {
		where[item.ID()] = "-"
	}

	for i, item := range ordered {
		if where[item.ID()] != want[i] {
			return true
		}
	}

	return false
}

func filledBoxes(run *Packing) int {
	used := 0

	for _, box := range run.boxes {
		if len(box.items) > 0 {
			used++
		}
	}

	return used
}
