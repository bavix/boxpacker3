package boxpacker3_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

const (
	fuzzMaxBoxes = 3
	fuzzMaxItems = 12
)

func decodeDimension(b byte, limit float64) float64 {
	return 1 + float64(b%64)*limit/64
}

func decodeInput(data []byte) ([]*boxpacker3.Box, []*boxpacker3.Item) {
	if len(data) < 8 {
		data = append(data, make([]byte, 8-len(data))...)
	}

	boxCount := 1 + int(data[0])%fuzzMaxBoxes
	itemCount := 1 + int(data[1])%fuzzMaxItems

	boxes := make([]*boxpacker3.Box, 0, boxCount)
	items := make([]*boxpacker3.Item, 0, itemCount)

	at := func(i int) byte { return data[i%len(data)] }

	cursor := 2
	for i := range boxCount {
		width := 20 + float64(at(cursor)%100)
		height := 20 + float64(at(cursor+1)%100)
		depth := 20 + float64(at(cursor+2)%100)
		cursor += 3

		boxes = append(boxes, boxpacker3.NewBox(
			fmt.Sprintf("box-%d", i), width, height, depth, 10000))
	}

	for i := range itemCount {
		width := decodeDimension(at(cursor), 40)
		height := decodeDimension(at(cursor+1), 40)
		depth := decodeDimension(at(cursor+2), 40)
		weight := float64(at(cursor+3) % 50)
		cursor += 4

		items = append(items, boxpacker3.NewItem(
			fmt.Sprintf("item-%d", i), width, height, depth, weight))
	}

	return boxes, items
}

func FuzzPacking(f *testing.F) {
	f.Add([]byte{1, 5, 100, 100, 100, 40, 40, 40, 10})
	f.Add([]byte{1, 3, 60, 60, 60, 60, 60, 60, 0})
	f.Add([]byte{1, 8, 30, 30, 30, 10, 10, 10, 1})
	f.Add([]byte{2, 12, 50, 50, 50, 25, 25, 25, 5, 25, 25, 25, 5})
	f.Add([]byte{1, 4, 120, 20, 20, 40, 20, 20, 3})
	f.Add([]byte{3, 12, 40, 40, 40, 1, 1, 1, 0, 39, 39, 39, 0})

	f.Fuzz(func(t *testing.T, data []byte) {
		boxes, items := decodeInput(data)

		for _, strategy := range allStrategies() {
			packer := rulePacker(strategy.Order, strategy.Selection)

			result, err := packer.Pack(t.Context(), boxes, items)
			require.NoError(t, err)

			requireInvariants(t, items, result)
		}
	})
}

func FuzzPackingWithSupport(f *testing.F) {
	f.Add([]byte{1, 6, 80, 80, 80, 30, 30, 20, 4})
	f.Add([]byte{1, 10, 100, 100, 40, 20, 20, 10, 2})

	f.Fuzz(func(t *testing.T, data []byte) {
		boxes, items := decodeInput(data)

		packer := boxpacker3.NewPacker(boxpacker3.WithRules(boxpacker3.Rules{MinSupportRatio: 1}))

		result, err := packer.Pack(t.Context(), boxes, items)
		require.NoError(t, err)

		requireInvariants(t, items, result)

		for _, box := range result.Boxes {
			for _, item := range box.Items {
				require.GreaterOrEqual(t, supportRatioOf(box, item), 1-1e-6,
					"item %s rests on %.3f of its base", item.ID(), supportRatioOf(box, item))
			}
		}
	})
}

func cappedItems(t *testing.T, items []*boxpacker3.Item, data []byte) ([]*boxpacker3.Item, map[string]float64) {
	t.Helper()

	if len(data) == 0 {
		data = []byte{0}
	}

	capped := make([]*boxpacker3.Item, 0, len(items))
	caps := make(map[string]float64, len(items))

	for i, source := range items {
		limit := 0.0
		keepClear := false

		switch data[i%len(data)] % 3 {
		case 0:
			limit = float64(data[(i+1)%len(data)] % 40)
		case 1:
			keepClear = true
		}

		built, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
			ID:           source.ID(),
			Width:        source.Width(),
			Height:       source.Height(),
			Depth:        source.Depth(),
			Weight:       source.Weight(),
			MaxLoadOnTop: limit,
			NothingOnTop: keepClear,
		})
		require.NoError(t, err)

		if keepClear {
			limit = 0
		}

		caps[source.ID()] = limit

		capped = append(capped, built)
	}

	return capped, caps
}

func separated(t *testing.T, items []*boxpacker3.Item, data []byte) []*boxpacker3.Item {
	t.Helper()

	if len(data) == 0 {
		data = []byte{0}
	}

	classes := []string{classFood, classChemicals, "steel"}
	labelled := make([]*boxpacker3.Item, 0, len(items))

	for i, source := range items {
		class := classes[int(data[i%len(data)])%len(classes)]

		var away []string
		if class == classFood {
			away = []string{classChemicals}
		}

		built, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
			ID:           source.ID(),
			Width:        source.Width(),
			Height:       source.Height(),
			Depth:        source.Depth(),
			Weight:       source.Weight(),
			Class:        class,
			SeparateFrom: away,
		})
		require.NoError(t, err)

		labelled = append(labelled, built)
	}

	return labelled
}

func FuzzPackingWithSeparation(f *testing.F) {
	f.Add([]byte{1, 8, 90, 90, 90, 30, 30, 30, 9})
	f.Add([]byte{2, 12, 60, 60, 60, 20, 20, 20, 40, 20, 20, 20, 1})

	f.Fuzz(func(t *testing.T, data []byte) {
		boxes, plain := decodeInput(data)
		items := separated(t, plain, data)

		result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
		require.NoError(t, err)

		requireInvariants(t, items, result)

		for _, box := range result.Boxes {
			food, chemicals := false, false

			for _, packed := range box.Items {
				food = food || packed.Item.Class() == classFood
				chemicals = chemicals || packed.Item.Class() == classChemicals
			}

			require.False(t, food && chemicals,
				"box %s holds food with chemicals", box.ID())
		}
	})
}

func FuzzPackingWithLoadLimits(f *testing.F) {
	f.Add([]byte{1, 6, 90, 90, 90, 30, 30, 30, 9})
	f.Add([]byte{2, 12, 60, 60, 60, 20, 20, 20, 40, 20, 20, 20, 1})
	f.Add([]byte{1, 4, 40, 40, 40, 39, 39, 39, 0})

	f.Fuzz(func(t *testing.T, data []byte) {
		boxes, plain := decodeInput(data)
		items, caps := cappedItems(t, plain, data)

		result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
		require.NoError(t, err)

		requireInvariants(t, items, result)

		for _, box := range result.Boxes {
			for _, item := range box.Items {
				load := loadOnTopOf(box, item)

				if item.Item.NothingOnTop() {
					require.Zero(t, load, "%s must stay clear", item.ID())

					continue
				}

				if limit := caps[item.ID()]; limit > 0 {
					require.LessOrEqual(t, load, limit+1e-9,
						"%s carries %.0f of %.0f", item.ID(), load, limit)
				}
			}
		}
	})
}

func FuzzSearch(f *testing.F) {
	f.Add([]byte{1, 8, 90, 90, 90, 30, 30, 30, 9})
	f.Add([]byte{2, 14, 70, 70, 70, 20, 20, 20, 30, 35, 35, 35, 4})
	f.Add([]byte{1, 20, 100, 60, 60, 25, 25, 25, 40})

	f.Fuzz(func(t *testing.T, data []byte) {
		boxes, items := decodeInput(data)
		if len(boxes) == 0 || len(items) == 0 {
			return
		}

		search, err := boxpacker3.NewSearch(32, 3)
		require.NoError(t, err)

		result, err := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(search)).
			Pack(t.Context(), boxes, items)
		require.NoError(t, err)

		requireInvariants(t, items, result)
	})
}

func FuzzQuantitiesAndReasons(f *testing.F) {
	f.Add([]byte{1, 6, 80, 80, 80, 25, 25, 25, 5}, uint8(3))
	f.Add([]byte{2, 10, 60, 60, 60, 20, 20, 20, 8, 30, 30, 30, 2}, uint8(7))
	f.Add([]byte{1, 4, 40, 40, 40, 39, 39, 39, 0}, uint8(1))

	f.Fuzz(func(t *testing.T, data []byte, quantity uint8) {
		boxes, plain := decodeInput(data)
		if len(boxes) == 0 || len(plain) == 0 {
			return
		}

		items := make([]*boxpacker3.Item, 0, len(plain))

		for i, source := range plain {
			built, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
				ID:       source.ID(),
				Width:    source.Width(),
				Height:   source.Height(),
				Depth:    source.Depth(),
				Weight:   source.Weight(),
				Quantity: 1 + int(quantity)%4 + i%2,
			})
			require.NoError(t, err)

			items = append(items, built)
		}

		result, err := boxpacker3.NewPacker().Pack(t.Context(), boxes, items)
		require.NoError(t, err)

		requireInvariants(t, items, result)

		for _, left := range result.Unpacked {
			holdable := false
			for _, box := range boxes {
				holdable = holdable || couldHold(box, left.Item)
			}

			if left.Reason.Structural() {
				require.False(t, holdable, "%s says %s but a box could hold it", left.ID(), left.Reason)
			}
		}
	})
}
