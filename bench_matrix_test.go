package boxpacker3_test

import (
	"fmt"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

func matrixFixture() ([]*boxpacker3.Box, []*boxpacker3.Item) {
	boxes := []*boxpacker3.Box{
		boxpacker3.NewBox("s1", 60, 60, 60, 50000),
		boxpacker3.NewBox("s2", 60, 60, 60, 50000),
		boxpacker3.NewBox("m1", 90, 90, 90, 50000),
		boxpacker3.NewBox("m2", 90, 90, 90, 50000),
		boxpacker3.NewBox("l1", 120, 120, 120, 50000),
	}

	items := make([]*boxpacker3.Item, 0, 60)

	for i := range 60 {
		width := float64(10 + i%9*5)
		height := float64(10 + (i*7)%11*4)
		depth := float64(10 + (i*3)%7*6)
		items = append(items, boxpacker3.NewItem(
			fmt.Sprintf("item-%d", i), width, height, depth, float64(10+i%13*20)))
	}

	return boxes, items
}

func BenchmarkStrategyMatrix(b *testing.B) {
	for _, order := range boxpacker3.ItemOrders() {
		for _, selection := range boxpacker3.BoxSelections() {
			b.Run(selection.String()+"/"+order.Name(), func(b *testing.B) {
				packer := boxpacker3.NewPacker(
					boxpacker3.WithAlgorithm(boxpacker3.NewGreedy(order, selection)))

				var (
					used       int
					leftBehind int
					fill       float64
				)

				b.ReportAllocs()

				for b.Loop() {
					boxes, items := matrixFixture()

					result, err := packer.Pack(b.Context(), boxes, items)
					if err != nil {
						b.Fatal(err)
					}

					used, leftBehind, fill = 0, len(result.Unpacked), 0

					for _, box := range result.Boxes {
						if len(box.Items) == 0 {
							continue
						}

						used++
						fill += volumeUtilisation(box)
					}

					if used > 0 {
						fill /= float64(used)
					}
				}

				b.StopTimer()
				b.ReportMetric(float64(used), "boxes")
				b.ReportMetric(fill*100, "fill%")
				b.ReportMetric(float64(leftBehind), "left-behind")
			})
		}
	}
}
