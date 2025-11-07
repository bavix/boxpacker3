package boxpacker3_test

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/google/uuid"

	"github.com/bavix/boxpacker3"
)

// BenchmarkPacker benchmarks the Pack method of the Packer type.
//
// This benchmark uses a list of 100 random items and the default box list.
// The benchmark reports the number of allocations and resets the timer before
// the loop.
//
// The loop iterates over the items and packs them into boxes.
func BenchmarkPacker(b *testing.B) {
	// Create a slice of 100 random items
	items := make([]*boxpacker3.Item, 0, 100)

	for range cap(items) {
		// Generate random dimensions for the item
		w, _ := rand.Int(rand.Reader, big.NewInt(150))
		l, _ := rand.Int(rand.Reader, big.NewInt(150))
		h, _ := rand.Int(rand.Reader, big.NewInt(150))
		w2, _ := rand.Int(rand.Reader, big.NewInt(100))

		// Create a new item with the random dimensions
		items = append(items, boxpacker3.NewItem(
			uuid.New().String(),
			float64(w.Int64()),
			float64(l.Int64()),
			float64(h.Int64()),
			float64(w2.Int64()),
		))
	}

	// Create a new box list with the default boxes
	boxes := NewDefaultBoxList()
	// Create a new packer
	packer := boxpacker3.NewPacker()

	// Report allocations and reset the timer
	b.ReportAllocs()
	b.ResetTimer()

	// Iterate over the items and pack them into boxes
	for range b.N {
		_ = packer.Pack(boxes, items)
	}
}

// generateItems creates a slice of random items with the specified count.
func generateItems(count int) []*boxpacker3.Item {
	items := make([]*boxpacker3.Item, 0, count)

	for range count {
		// Generate random dimensions for the item
		w, _ := rand.Int(rand.Reader, big.NewInt(150))
		l, _ := rand.Int(rand.Reader, big.NewInt(150))
		h, _ := rand.Int(rand.Reader, big.NewInt(150))
		w2, _ := rand.Int(rand.Reader, big.NewInt(100))

		// Create a new item with the random dimensions
		items = append(items, boxpacker3.NewItem(
			uuid.New().String(),
			float64(w.Int64())+10, // Ensure minimum size
			float64(l.Int64())+10,
			float64(h.Int64())+10,
			float64(w2.Int64())+10,
		))
	}

	return items
}

// BenchmarkPacker_StrategyMinimizeBoxes benchmarks the MinimizeBoxes strategy.
func BenchmarkPacker_StrategyMinimizeBoxes(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyMinimizeBoxes))

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_ = packer.Pack(boxes, items)
	}
}

// BenchmarkPacker_StrategyGreedy benchmarks the Greedy strategy.
func BenchmarkPacker_StrategyGreedy(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyGreedy))

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_ = packer.Pack(boxes, items)
	}
}

// BenchmarkPacker_StrategyBestFit benchmarks the BestFit strategy.
func BenchmarkPacker_StrategyBestFit(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFit))

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_ = packer.Pack(boxes, items)
	}
}

// BenchmarkPacker_StrategyBestFitDecreasing benchmarks the BestFitDecreasing strategy.
func BenchmarkPacker_StrategyBestFitDecreasing(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyBestFitDecreasing))

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_ = packer.Pack(boxes, items)
	}
}

// BenchmarkPacker_StrategyNextFit benchmarks the NextFit strategy.
func BenchmarkPacker_StrategyNextFit(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyNextFit))

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_ = packer.Pack(boxes, items)
	}
}

// BenchmarkPacker_StrategyWorstFit benchmarks the WorstFit strategy.
func BenchmarkPacker_StrategyWorstFit(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyWorstFit))

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_ = packer.Pack(boxes, items)
	}
}

// BenchmarkPacker_StrategyAlmostWorstFit benchmarks the AlmostWorstFit strategy.
func BenchmarkPacker_StrategyAlmostWorstFit(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()
	packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(boxpacker3.StrategyAlmostWorstFit))

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		_ = packer.Pack(boxes, items)
	}
}

// BenchmarkPacker_DifferentSizes benchmarks different item counts.
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
				_ = packer.Pack(boxes, items)
			}
		})
	}
}

// BenchmarkPacker_StrategyComparison compares all strategies on the same dataset.
func BenchmarkPacker_StrategyComparison(b *testing.B) {
	items := generateItems(100)
	boxes := NewDefaultBoxList()

	strategies := []struct {
		name     string
		strategy boxpacker3.PackingStrategy
	}{
		{"MinimizeBoxes", boxpacker3.StrategyMinimizeBoxes},
		{"Greedy", boxpacker3.StrategyGreedy},
		{"BestFit", boxpacker3.StrategyBestFit},
		{"BestFitDecreasing", boxpacker3.StrategyBestFitDecreasing},
		{"NextFit", boxpacker3.StrategyNextFit},
		{"WorstFit", boxpacker3.StrategyWorstFit},
		{"AlmostWorstFit", boxpacker3.StrategyAlmostWorstFit},
	}

	for _, strategy := range strategies {
		b.Run(strategy.name, func(b *testing.B) {
			packer := boxpacker3.NewPacker(boxpacker3.WithStrategy(strategy.strategy))

			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				_ = packer.Pack(boxes, items)
			}
		})
	}
}
