package boxpacker3

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestEnforceGroups_StrandsNothingABoxWouldTake(t *testing.T) {
	t.Parallel()

	stranded := 0

	for seed := range 200 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec
		result := packedWithGroups(t, random)

		before := len(result.unfit)
		enforceGroups(result)

		for _, item := range result.unfit[before:] {
			if item.group == "" && anyBoxTakes(result, item) {
				stranded++
			}
		}
	}

	if stranded > 0 {
		t.Errorf("%d ungrouped goods were stranded although a box would take them", stranded)
	}
}

func packedWithGroups(t *testing.T, random *rand.Rand) *Packing {
	t.Helper()

	boxes := make([]*Box, 0, 3)

	for i := range 3 {
		box, err := NewBoxFromSpec(BoxSpec{
			ID: fmt.Sprintf("box-%d", i), OuterWidth: 100, OuterHeight: 100, OuterDepth: 100,
			MaxWeight: 1e9, Quantity: 4,
		})
		if err != nil {
			t.Fatal(err)
		}

		boxes = append(boxes, box)
	}

	items := make([]*Item, 0, 50)

	for i := range 20 + random.Intn(30) {
		group := ""
		if random.Intn(3) == 0 {
			group = fmt.Sprintf("group-%d", random.Intn(4))
		}

		item, err := NewItemFromSpec(ItemSpec{
			ID:     fmt.Sprintf("item-%d", i),
			Width:  10 + float64(random.Intn(50)),
			Height: 10 + float64(random.Intn(50)),
			Depth:  10 + float64(random.Intn(50)),
			Weight: 1,
			Group:  group,
		})
		if err != nil {
			t.Fatal(err)
		}

		items = append(items, item)
	}

	result, err := NewGreedy(OrderDecreasing, SelectFullestBox).Pack(t.Context(), NewProblem(boxes, items))
	if err != nil {
		t.Fatal(err)
	}

	return result
}
