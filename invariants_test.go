package boxpacker3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

const geometryTolerance = 1e-9

func itemsOverlap(a, b boxpacker3.PackedItem) bool {
	aPosition, aDimension := a.Position, a.Dimension
	bPosition, bDimension := b.Position, b.Dimension

	for axis := range 3 {
		low := max(aPosition[axis], bPosition[axis])
		high := min(aPosition[axis]+aDimension[axis], bPosition[axis]+bDimension[axis])
		scale := max(aDimension[axis], bDimension[axis])

		if high-low <= scale*geometryTolerance {
			return false
		}
	}

	return true
}

func requireNoOverlap(t *testing.T, box boxpacker3.PackedBox) {
	t.Helper()

	items := box.Items
	for i := range items {
		for j := i + 1; j < len(items); j++ {
			require.False(t, itemsOverlap(items[i], items[j]),
				"items %s %v%v and %s %v%v share space in box %s",
				items[i].ID(), items[i].Position, items[i].Dimension,
				items[j].ID(), items[j].Position, items[j].Dimension,
				box.ID())
		}
	}
}

func requireWithinBox(t *testing.T, box boxpacker3.PackedBox, item boxpacker3.PackedItem) {
	t.Helper()

	position, dimension := item.Position, item.Dimension
	limits := [3]float64{box.Box.Width(), box.Box.Height(), box.Box.Depth()}

	for axis := range 3 {
		require.GreaterOrEqual(t, position[axis], -geometryTolerance,
			"item %s starts outside box %s on axis %d", item.ID(), box.ID(), axis)
		require.LessOrEqual(t, position[axis]+dimension[axis],
			limits[axis]+max(limits[axis], 1)*geometryTolerance,
			"item %s sticks out of box %s on axis %d", item.ID(), box.ID(), axis)
	}
}

func requireWithinCapacity(t *testing.T, box boxpacker3.PackedBox) {
	t.Helper()

	var weight, volume float64

	for _, item := range box.Items {
		weight += item.Weight()
		volume += item.Volume()
	}

	require.LessOrEqual(t, weight+box.Box.EmptyWeight(), box.Box.MaxWeight()+geometryTolerance,
		"box %s is over its weight limit once its tare is counted", box.ID())
	require.LessOrEqual(t, volume, box.Volume()+max(box.Volume(), 1)*geometryTolerance,
		"box %s is over its volume", box.ID())
}

func requireInvariants(t *testing.T, offered []*boxpacker3.Item, result *boxpacker3.Result) {
	t.Helper()

	require.NotNil(t, result)

	type copyOf struct {
		id    string
		index int
	}

	seen := make(map[copyOf]int, len(offered))

	for _, box := range result.Boxes {
		require.NotNil(t, box.Box)
		requireNoOverlap(t, box)
		requireWithinCapacity(t, box)

		for _, item := range box.Items {
			require.NotNil(t, item.Item)
			requireWithinBox(t, box, item)

			seen[copyOf{id: item.ID(), index: item.Index}]++
		}
	}

	for _, item := range result.Unpacked {
		require.NotNil(t, item.Item)

		seen[copyOf{id: item.ID(), index: item.Index}]++
	}

	expected := make(map[copyOf]int, len(offered))

	for _, item := range offered {
		if item == nil {
			continue
		}

		for index := range item.Quantity() {
			expected[copyOf{id: item.ID(), index: index}]++
		}
	}

	require.Equal(t, expected, seen, "every copy offered must appear exactly once")
}

func volumeUtilisation(box boxpacker3.PackedBox) float64 {
	if box.Volume() <= 0 {
		return 0
	}

	return box.Stats.ItemsVolume / box.Volume()
}
