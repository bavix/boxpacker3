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
	for i := 0; i < b.N; i++ {
		_ = packer.Pack(boxes, items)
	}
}
