package boxpacker3_test

import (
	rand2 "crypto/rand"
	"math/big"
	"testing"

	"github.com/google/uuid"

	"github.com/bavix/boxpacker3"
)

func BenchmarkPacker(b *testing.B) {
	items := make(boxpacker3.ItemSlice, 0, 100)

	for x := 0; x < 100; x++ {
		w, _ := rand2.Int(rand2.Reader, big.NewInt(150))
		l, _ := rand2.Int(rand2.Reader, big.NewInt(150))
		h, _ := rand2.Int(rand2.Reader, big.NewInt(150))
		w2, _ := rand2.Int(rand2.Reader, big.NewInt(100))

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

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = packer.Pack(boxes, items)
	}
}
