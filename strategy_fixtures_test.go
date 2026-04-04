package boxpacker3_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

func shelfOf(random *rand.Rand, count, minSide, spread int, maxWeight float64) []*boxpacker3.Box {
	boxes := make([]*boxpacker3.Box, 0, count)

	for i := range count {
		side := float64(minSide + random.Intn(spread))
		boxes = append(boxes, boxpacker3.NewBox(
			fmt.Sprintf("box-%d", i), side, side*0.8, side*0.7, maxWeight))
	}

	return boxes
}

func identicalShelf(count int, width, height, depth, maxWeight float64) []*boxpacker3.Box {
	boxes := make([]*boxpacker3.Box, 0, count)

	for i := range count {
		boxes = append(boxes, boxpacker3.NewBox(
			fmt.Sprintf("same-%d", i), width, height, depth, maxWeight))
	}

	return boxes
}

func cargoOf(random *rand.Rand, count, minSide, spread, minWeight, weightSpread int) []*boxpacker3.Item {
	items := make([]*boxpacker3.Item, 0, count)

	for i := range count {
		items = append(items, boxpacker3.NewItem(
			fmt.Sprintf("item-%d", i),
			float64(minSide+random.Intn(spread)),
			float64(minSide+random.Intn(spread)),
			float64(minSide+random.Intn(spread)),
			float64(minWeight+random.Intn(max(weightSpread, 1)))))
	}

	return items
}

func usedBoxCountOf(result *boxpacker3.Result) int {
	used := 0

	for _, box := range result.Boxes {
		if len(box.Items) > 0 {
			used++
		}
	}

	return used
}

func packsAlone(box *boxpacker3.Box, item *boxpacker3.Item) bool {
	alone, err := boxpacker3.NewPacker().Pack(
		context.Background(), []*boxpacker3.Box{box}, []*boxpacker3.Item{item})

	return err == nil && len(alone.Unpacked) == 0
}

func packed(tb testing.TB, packer *boxpacker3.Packer, boxes []*boxpacker3.Box, items []*boxpacker3.Item) *boxpacker3.Result {
	tb.Helper()

	result, err := packer.Pack(tb.Context(), boxes, items)
	if err != nil {
		tb.Fatal(err)
	}

	return result
}

func rulePacker(order boxpacker3.ItemSorter, selection boxpacker3.BoxSelection) *boxpacker3.Packer {
	return boxpacker3.NewPacker(boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(order, selection)))
}

func singlePacker() *boxpacker3.Packer {
	return boxpacker3.NewPacker(boxpacker3.WithAlgorithm(boxpacker3.SingleContainer(nil)))
}
