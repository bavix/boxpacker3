//go:build measure

package boxpacker3_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

type fixtureBox struct {
	ID       string  `json:"id"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Depth    float64 `json:"depth"`
	Weight   float64 `json:"weight"`
	Quantity int     `json:"quantity"`
}

type fixtureItem struct {
	ID       string  `json:"id"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Depth    float64 `json:"depth"`
	Weight   float64 `json:"weight"`
	Quantity int     `json:"quantity"`
}

type fixtureOrder struct {
	Name  string        `json:"name"`
	Items []fixtureItem `json:"items"`
}

type fixture struct {
	Name   string         `json:"name"`
	Boxes  []fixtureBox   `json:"boxes"`
	Orders []fixtureOrder `json:"orders"`
}

const referenceDir = "testdata/reference"

func identicalCartons(count int) []fixtureBox {
	boxes := make([]fixtureBox, 0, count)
	for i := range count {
		boxes = append(boxes, fixtureBox{
			ID: fmt.Sprintf("same-%d", i), Width: 425, Height: 265, Depth: 380,
			Weight: 20000, Quantity: 1,
		})
	}

	return boxes
}

func sevenSizes() []fixtureBox {
	sizes := [][3]float64{
		{220, 185, 50},
		{165, 215, 100},
		{265, 165, 190},
		{425, 165, 190},
		{425, 265, 190},
		{425, 265, 380},
		{530, 380, 265},
	}

	boxes := make([]fixtureBox, 0, len(sizes))
	for i, size := range sizes {
		boxes = append(boxes, fixtureBox{
			ID: fmt.Sprintf("kind-%d", i), Width: size[0], Height: size[1], Depth: size[2],
			Weight: 20000, Quantity: 30,
		})
	}

	return boxes
}

func randomOrders(count int, seedFrom int64) []fixtureOrder {
	orders := make([]fixtureOrder, 0, count)

	for n := range count {
		random := rand.New(rand.NewSource(seedFrom + int64(n))) //nolint:gosec

		items := make([]fixtureItem, 0, 40)
		for i := range 12 + random.Intn(26) {
			items = append(items, fixtureItem{
				ID:       fmt.Sprintf("item-%d", i),
				Width:    float64(30 + random.Intn(140)),
				Height:   float64(30 + random.Intn(140)),
				Depth:    float64(20 + random.Intn(110)),
				Weight:   float64(100 + random.Intn(500)),
				Quantity: 1,
			})
		}

		orders = append(orders, fixtureOrder{Name: fmt.Sprintf("order-%d", n), Items: items})
	}

	return orders
}

func bookshop(count int) []fixtureOrder {
	orders := make([]fixtureOrder, 0, count)

	for n := range count {
		random := rand.New(rand.NewSource(9000 + int64(n))) //nolint:gosec

		items := make([]fixtureItem, 0, 4)
		for i := range 1 + random.Intn(3) {
			items = append(items, fixtureItem{
				ID:       fmt.Sprintf("book-%d", i),
				Width:    float64(120 + random.Intn(80)),
				Height:   float64(180 + random.Intn(70)),
				Depth:    float64(15 + random.Intn(35)),
				Weight:   float64(200 + random.Intn(500)),
				Quantity: 1 + random.Intn(12),
			})
		}

		orders = append(orders, fixtureOrder{Name: fmt.Sprintf("basket-%d", n), Items: items})
	}

	return orders
}

func fixtures() []fixture {
	return []fixture{
		{Name: "identical-cartons", Boxes: identicalCartons(30), Orders: randomOrders(40, 1)},
		{Name: "seven-sizes", Boxes: sevenSizes(), Orders: randomOrders(40, 500)},
		{Name: "bookshop", Boxes: sevenSizes(), Orders: bookshop(120)},
	}
}

func TestWriteReferenceFixtures(t *testing.T) {
	t.Parallel()

	err := os.MkdirAll(referenceDir, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	for _, set := range fixtures() {
		encoded, err := json.MarshalIndent(set, "", "  ")
		if err != nil {
			t.Fatal(err)
		}

		path := filepath.Join(referenceDir, set.Name+".json")

		err = os.WriteFile(path, encoded, 0o600)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("wrote %s: %d boxes, %d orders", path, len(set.Boxes), len(set.Orders))
	}
}

func boxesOf(set fixture) []*boxpacker3.Box {
	boxes := make([]*boxpacker3.Box, 0, len(set.Boxes))

	for _, box := range set.Boxes {
		built, err := boxpacker3.NewBoxFromSpec(boxpacker3.BoxSpec{
			ID:          box.ID,
			OuterWidth:  box.Width,
			OuterHeight: box.Height,
			OuterDepth:  box.Depth,
			InnerWidth:  box.Width,
			InnerHeight: box.Height,
			InnerDepth:  box.Depth,
			MaxWeight:   box.Weight,
			Quantity:    max(box.Quantity, 1),
		})
		if err != nil {
			panic(err)
		}

		boxes = append(boxes, built)
	}

	return boxes
}

func itemsOf(order fixtureOrder) []*boxpacker3.Item {
	items := make([]*boxpacker3.Item, 0, len(order.Items))

	for _, item := range order.Items {
		for copyAt := range max(item.Quantity, 1) {
			items = append(items, boxpacker3.NewItem(
				fmt.Sprintf("%s#%d", item.ID, copyAt+1),
				item.Width, item.Height, item.Depth, item.Weight))
		}
	}

	return items
}

func TestReference_Before(t *testing.T) {
	t.Parallel()

	for _, set := range fixtures() {
		t.Log("=== " + set.Name + " ===")

		for _, rule := range boxpacker3.BoxSelections() {
			containers, volume, unfit := 0, 0.0, 0

			for _, order := range set.Orders {
				result, err := boxpacker3.NewPacker(
					boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(boxpacker3.OrderDecreasing, rule)),
				).Pack(t.Context(), boxesOf(set), itemsOf(order))
				if err != nil {
					t.Fatal(err)
				}

				for _, box := range result.Boxes {
					if len(box.Items) == 0 {
						continue
					}

					containers++
					volume += box.Box.Width() * box.Box.Height() * box.Box.Depth()
				}

				unfit += len(result.Unpacked)
			}

			runs := float64(len(set.Orders))
			t.Logf("%-18s %7.2f containers %10.1f L %6.2f left",
				rule, float64(containers)/runs, volume/runs/1e6, float64(unfit)/runs)
		}
	}
}
