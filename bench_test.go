package boxpacker3_test

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/google/uuid"

	"github.com/bavix/boxpacker3"
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

	for i := 0; i < b.N; i++ {
		_ = packer.Pack(boxes, items)
	}
}
