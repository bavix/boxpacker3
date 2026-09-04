package boxpacker3_test

import (
	"context"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/google/uuid"

	"github.com/bavix/boxpacker3/v2"
)

func BenchmarkPacker(b *testing.B) {
	items := make([]*boxpacker3.Item, 0, 100)

	for range cap(items) {
		w, _ := rand.Int(rand.Reader, big.NewInt(150))
		l, _ := rand.Int(rand.Reader, big.NewInt(150))
		h, _ := rand.Int(rand.Reader, big.NewInt(150))
		w2, _ := rand.Int(rand.Reader, big.NewInt(100))

		items = append(items, boxpacker3.NewItem(
			uuid.New().String(),
			float64(w.Int64()),
			float64(l.Int64()),
			float64(h.Int64()),
			float64(w2.Int64()),
		))
	}

	boxes := NewDefaultBoxList()

	packer := boxpacker3.NewPacker()

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, _ = packer.Pack(context.Background(), boxes, items)
	}
}

func generateItems(count int) []*boxpacker3.Item {
	items := make([]*boxpacker3.Item, 0, count)

	for range count {
		w, _ := rand.Int(rand.Reader, big.NewInt(150))
		l, _ := rand.Int(rand.Reader, big.NewInt(150))
		h, _ := rand.Int(rand.Reader, big.NewInt(150))
		w2, _ := rand.Int(rand.Reader, big.NewInt(100))

		items = append(items, boxpacker3.NewItem(
			uuid.New().String(),
			float64(w.Int64())+10,
			float64(l.Int64())+10,
			float64(h.Int64())+10,
			float64(w2.Int64())+10,
		))
	}

	return items
}

func BenchmarkPacker_FirstFitDecreasing(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectFirstFit)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, _ = packer.Pack(context.Background(), boxes, items)
	}
}

func BenchmarkPacker_FirstFitIncreasing(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectFirstFit)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, _ = packer.Pack(context.Background(), boxes, items)
	}
}

func BenchmarkPacker_BestFitIncreasing(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectBestFit)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, _ = packer.Pack(context.Background(), boxes, items)
	}
}

func BenchmarkPacker_BestFitDecreasing(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := rulePacker(boxpacker3.OrderDecreasing, boxpacker3.SelectBestFit)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, _ = packer.Pack(context.Background(), boxes, items)
	}
}

func BenchmarkPacker_NextFitIncreasing(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectNextFit)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, _ = packer.Pack(context.Background(), boxes, items)
	}
}

func BenchmarkPacker_WorstFitIncreasing(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectWorstFit)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, _ = packer.Pack(context.Background(), boxes, items)
	}
}

func BenchmarkPacker_AlmostWorstFitIncreasing(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := rulePacker(boxpacker3.OrderIncreasing, boxpacker3.SelectAlmostWorstFit)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_, _ = packer.Pack(context.Background(), boxes, items)
	}
}

func BenchmarkPacker_DifferentSizes(b *testing.B) {
	sizes := []struct {
		name  string
		count int
	}{
		{"10Items", 10},
		{"50Items", 50},
		{"100Items", 100},
		{"200Items", 200},
		{"500Items", 500},
	}

	boxes := NewDefaultBoxList()

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			items := generateItems(size.count)
			packer := boxpacker3.NewPacker()

			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				_, _ = packer.Pack(context.Background(), boxes, items)
			}
		})
	}
}

func BenchmarkPacker_StrategyComparison(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()

	strategies := []struct {
		name     string
		strategy boxpacker3.RuleSettings
	}{
		{nameFirstFitDecreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectFirstFit}},
		{nameFirstFitIncreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectFirstFit}},
		{nameBestFitIncreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectBestFit}},
		{nameBestFitDecreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectBestFit}},
		{nameNextFitIncreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectNextFit}},
		{nameWorstFitIncreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectWorstFit}},
		{nameAlmostWorstFitIncreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectAlmostWorstFit}},
	}

	for _, strategy := range strategies {
		b.Run(strategy.name, func(b *testing.B) {
			packer := rulePacker(strategy.strategy.Order, strategy.strategy.Selection)

			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				_, _ = packer.Pack(context.Background(), boxes, items)
			}
		})
	}
}
