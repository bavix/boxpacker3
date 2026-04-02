package boxpacker3

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
)

func slackOf(box *Container) float64 {
	return box.remainingVolume()
}

func randomShelf(random *rand.Rand, count int) []*Container {
	boxes := make([]*Container, 0, count)

	for i := range count {
		box := openBox(fmt.Sprintf("box-%d", i), 100, 100, 100, 1e9)

		filled := 10 + random.Intn(80)
		if !box.putItem(testPiece(fmt.Sprintf("fill-%d", i), float64(filled), 100, 100, 1), Pivot{}) {
			panic("the filler did not go in")
		}

		boxes = append(boxes, box)
	}

	return boxes
}

func slackList(boxes []*Container) []float64 {
	slack := make([]float64, len(boxes))

	for i, box := range boxes {
		slack[i] = slackOf(box)
	}

	return slack
}

func ranked(slack []float64) []float64 {
	sorted := append([]float64(nil), slack...)

	for i := range sorted {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] > sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

func wantedSlack(rule BoxSelection, slack []float64) float64 {
	order := ranked(slack)

	switch rule {
	case SelectBestFit:
		return order[len(order)-1]
	case SelectAlmostWorstFit:
		if len(order) > 1 {
			return order[1]
		}

		return order[0]
	case SelectWorstFit, SelectFirstFit, SelectNextFit, SelectFullestBox:
	}

	return order[0]
}

func TestSelectBox_MatchesTheDefinitionOnRandomShelves(t *testing.T) {
	t.Parallel()

	rules := []BoxSelection{SelectBestFit, SelectWorstFit, SelectAlmostWorstFit}
	deviations := map[string]int{}

	for seed := range 600 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec
		count := 2 + random.Intn(5)

		boxes := randomShelf(random, count)

		choice, found := selectBox(boxes, testPiece("probe", 10, 10, 10, 1), SelectFirstFit)
		if !found {
			t.Fatal("nothing took the probe")
		}

		if choice.index != 0 {
			deviations["first-fit"]++
		}

		for _, rule := range rules {
			shelf := randomShelf(random, count)
			slack := slackList(shelf)

			picked, _ := selectBox(shelf, testPiece("probe", 10, 10, 10, 1), rule)
			if slack[picked.index] != wantedSlack(rule, slack) {
				deviations[rule.String()]++
			}
		}
	}

	if len(deviations) > 0 {
		t.Errorf("deviations from the definitions: %v", deviations)
	}
}

func shelfWithOpenAndFresh(t *testing.T, random *rand.Rand, count int) []*Container {
	t.Helper()

	boxes := make([]*Container, 0, count)

	for i := range count {
		box := openBox(fmt.Sprintf("box-%d", i), 100, 100, 100, 1e9)

		if random.Intn(2) == 0 {
			filled := 10 + random.Intn(70)
			if !box.putItem(testPiece(fmt.Sprintf("fill-%d", i),
				float64(filled), 100, 100, 1), Pivot{}) {
				t.Fatal("the filler did not go in")
			}
		}

		boxes = append(boxes, box)
	}

	return boxes
}

func openSlack(boxes []*Container) []float64 {
	slack := make([]float64, 0, len(boxes))

	for _, box := range boxes {
		if len(box.items) > 0 {
			slack = append(slack, box.remainingVolume())
		}
	}

	return slack
}

func checkOpenFirst(boxes []*Container, rule BoxSelection, breaches map[string]int) {
	choice, found := selectBox(boxes, testPiece("probe", 10, 10, 10, 1), rule)
	if !found {
		return
	}

	open := openSlack(boxes)
	chosen := boxes[choice.index]

	if len(chosen.items) == 0 {
		if len(open) > 0 {
			breaches[rule.String()+": opened a box while one was already open"]++
		}

		return
	}

	if chosen.remainingVolume() != wantedSlack(rule, open) {
		breaches[rule.String()+": wrong box among the open ones"]++
	}
}

func TestSelectBox_ChoosesAmongOpenBoxesFirst(t *testing.T) {
	t.Parallel()

	rules := []BoxSelection{SelectBestFit, SelectWorstFit, SelectAlmostWorstFit}
	breaches := map[string]int{}

	for seed := range 800 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec
		count := 2 + random.Intn(5)

		for _, rule := range rules {
			checkOpenFirst(shelfWithOpenAndFresh(t, random, count), rule, breaches)
		}
	}

	if len(breaches) > 0 {
		t.Errorf("breaches of the open-first definition: %v", breaches)
	}
}

func exhaustiveFullest(ctx context.Context, boxes []*Container, remaining []*piece) int {
	best, bestVolume := -1, 0.0
	seen := map[string]bool{}

	for i, box := range boxes {
		if box == nil || len(box.items) > 0 || seen[box.id] {
			continue
		}

		seen[box.id] = true

		trial := clonePtr(box)
		trial.reset()

		placements := snapshot(remaining)
		_, _ = packToBox(ctx, trial, remaining)
		restore(remaining, placements)

		measured := trial.itemsVolume
		if measured <= 0 {
			continue
		}

		if best < 0 || measured > bestVolume ||
			(measured == bestVolume && box.volume < boxes[best].volume) {
			best, bestVolume = i, measured
		}
	}

	return best
}

func TestFullestBoxSearchMatchesTheDefinition(t *testing.T) {
	t.Parallel()

	deviations := 0

	for seed := range 400 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

		inputBoxes := make([]*Box, 0, 6)

		for i := range 2 + random.Intn(5) {
			side := 90 + float64(random.Intn(220))
			inputBoxes = append(inputBoxes, NewBox(
				fmt.Sprintf("box-%d", i), side, side*0.8, side*0.7, 1e9))
		}

		items := make([]*piece, 0, 20)
		for i := range 5 + random.Intn(14) {
			items = append(items, testPiece(
				fmt.Sprintf("item-%d", i),
				20+float64(random.Intn(90)),
				20+float64(random.Intn(90)),
				20+float64(random.Intn(90)), 1))
		}

		sorted, _ := prepareData(inputBoxes, items, problemOf(inputBoxes, items))
		deviations += walkTheGreedyLoop(t, sorted, sortItems(compact(items), OrderDecreasing))
	}

	t.Logf("deviations from the definition: %d", deviations)
}

func walkTheGreedyLoop(t *testing.T, sorted []*Container, remaining []*piece) int {
	t.Helper()

	search := newFullestSearch(sorted)
	deviations := 0

	for round := 0; len(remaining) > 0 && round < 6; round++ {
		lazy, err := search.pick(t.Context(), sorted, remaining)
		if err != nil {
			t.Fatal(err)
		}

		want := exhaustiveFullest(t.Context(), sorted, remaining)

		if want < 0 {
			break
		}

		if lazy < 0 {
			deviations++

			break
		}

		if sorted[lazy].volume != sorted[want].volume {
			deviations++
		}

		before := len(remaining)

		remaining, err = packToBox(t.Context(), sorted[lazy], remaining)
		if err != nil {
			t.Fatal(err)
		}

		if len(remaining) == before {
			break
		}
	}

	return deviations
}

func TestFullestBoxSearchDoesNotTrustAStaleMeasurement(t *testing.T) {
	t.Parallel()

	shelf := func(t *testing.T, id string, width, height, depth float64, quantity int) *Box {
		t.Helper()

		box, err := NewBoxFromSpec(BoxSpec{
			ID: id, OuterWidth: width, OuterHeight: height, OuterDepth: depth,
			MaxWeight: 1e9, Quantity: quantity,
		})
		if err != nil {
			t.Fatal(err)
		}

		return box
	}

	boxes := []*Box{
		shelf(t, "cube", 10, 10, 10, 1),
		shelf(t, "tray", 10, 10, 5, 2),
		shelf(t, "wide", 11, 11, 6, 1),
	}

	items := make([]*piece, 0, 10)
	items = append(items, testPiece("block", 6, 6, 6, 1))

	for i := range 9 {
		items = append(items, testPiece(fmt.Sprintf("cube-%d", i), 5, 5, 5, 1))
	}

	sorted, result := prepareData(boxes, items, problemOf(boxes, items))
	remaining := sortItems(compact(items), OrderDecreasing)
	search := newFullestSearch(sorted)

	for len(remaining) > 0 {
		lazy, err := search.pick(t.Context(), sorted, remaining)
		if err != nil {
			t.Fatal(err)
		}

		want := exhaustiveFullest(t.Context(), sorted, remaining)
		if want < 0 || lazy < 0 {
			break
		}

		if sorted[lazy].id != sorted[want].id {
			t.Fatalf("search picked %s where the definition picks %s", sorted[lazy].id, sorted[want].id)
		}

		remaining, err = packToBox(t.Context(), sorted[lazy], remaining)
		if err != nil {
			t.Fatal(err)
		}
	}

	result.unfit = append(result.unfit, remaining...)

	if used := countUsedBoxes(result.boxes); used != 2 {
		t.Errorf("wide takes the block and three cubes, cube takes the other six: 2 boxes, got %d", used)
	}
}

func TestFullestBoxTieIsJudgedToWithinRounding(t *testing.T) {
	t.Parallel()

	boxes := []*Container{openBox("small", 1, 1, 1, 1), openBox("large", 2, 2, 2, 1)}
	search := newFullestSearch(boxes)

	parts := []float64{0.1, 0.2, 0.3}
	forward, backward := 0.0, 0.0

	for i := range parts {
		forward += parts[i]
		backward += parts[len(parts)-1-i]
	}

	if forward == backward {
		t.Fatal("the sums must differ for the test to mean anything")
	}

	if search.prefer(boxes, 0, 1, forward, backward) {
		t.Error("the larger box must not win a tie")
	}

	if !search.prefer(boxes, 1, 0, backward, forward) {
		t.Error("the smaller box must win a tie")
	}
}
