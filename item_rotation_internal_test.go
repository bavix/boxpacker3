package boxpacker3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestItem_DimensionMatchesTheGetters(t *testing.T) {
	t.Parallel()

	item := testPiece("test", 10, 20, 30, 1)
	item.setOrientation(OrientationWHD)

	dim := item.dimension()

	require.InDelta(t, item.Width(), dim[0], 0.0001, "width should match")
	require.InDelta(t, item.Height(), dim[1], 0.0001, "height should match")
	require.InDelta(t, item.Depth(), dim[2], 0.0001, "depth should match")
}

func TestItem_DimensionForEveryOrientation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		width       float64
		height      float64
		depth       float64
		orientation Orientation
		expected    [3]float64
	}{
		{"Small_Whd", 1, 2, 3, OrientationWHD, [3]float64{1, 2, 3}},
		{"Small_Hwd", 1, 2, 3, OrientationHWD, [3]float64{2, 1, 3}},
		{"Small_Hdw", 1, 2, 3, OrientationHDW, [3]float64{2, 3, 1}},
		{"Small_Dhw", 1, 2, 3, OrientationDHW, [3]float64{3, 2, 1}},
		{"Small_Dwh", 1, 2, 3, OrientationDWH, [3]float64{3, 1, 2}},
		{"Small_Wdh", 1, 2, 3, OrientationWDH, [3]float64{1, 3, 2}},

		{"Large_Whd", 100, 200, 300, OrientationWHD, [3]float64{100, 200, 300}},
		{"Large_Hwd", 100, 200, 300, OrientationHWD, [3]float64{200, 100, 300}},
		{"Large_Hdw", 100, 200, 300, OrientationHDW, [3]float64{200, 300, 100}},
		{"Large_Dhw", 100, 200, 300, OrientationDHW, [3]float64{300, 200, 100}},
		{"Large_Dwh", 100, 200, 300, OrientationDWH, [3]float64{300, 100, 200}},
		{"Large_Wdh", 100, 200, 300, OrientationWDH, [3]float64{100, 300, 200}},

		{"Equal_Whd", 5, 5, 5, OrientationWHD, [3]float64{5, 5, 5}},
		{"Equal_Hwd", 5, 5, 5, OrientationHWD, [3]float64{5, 5, 5}},
		{"Equal_Hdw", 5, 5, 5, OrientationHDW, [3]float64{5, 5, 5}},

		{"Tiny_Whd", 0.1, 0.2, 0.3, OrientationWHD, [3]float64{0.1, 0.2, 0.3}},
		{"Tiny_Hwd", 0.1, 0.2, 0.3, OrientationHWD, [3]float64{0.2, 0.1, 0.3}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			item := testPiece("test", tc.width, tc.height, tc.depth, 1)
			item.setOrientation(tc.orientation)

			dim := item.dimension()

			require.InDelta(t, tc.expected[0], dim[0], 0.0001, "Width should match")
			require.InDelta(t, tc.expected[1], dim[1], 0.0001, "Height should match")
			require.InDelta(t, tc.expected[2], dim[2], 0.0001, "Depth should match")
		})
	}
}

func TestItem_GettersAtExtremeScales(t *testing.T) {
	t.Parallel()

	item := testPiece("tiny", 0.001, 0.002, 0.003, 0.1)
	require.InDelta(t, 0.001, item.Width(), 0.0001, "Very small width should work")
	require.InDelta(t, 0.002, item.Height(), 0.0001, "Very small height should work")
	require.InDelta(t, 0.003, item.Depth(), 0.0001, "Very small depth should work")

	item = testPiece("large", 10000, 20000, 30000, 1000)
	require.InDelta(t, 10000.0, item.Width(), 0.0001, "Very large width should work")
	require.InDelta(t, 20000.0, item.Height(), 0.0001, "Very large height should work")
	require.InDelta(t, 30000.0, item.Depth(), 0.0001, "Very large depth should work")

	item = testPiece("cube", 5, 5, 5, 1)
	require.InDelta(t, 5.0, item.Width(), 0.0001, "Equal dimensions should work")
	require.InDelta(t, 5.0, item.Height(), 0.0001, "Equal dimensions should work")
	require.InDelta(t, 5.0, item.Depth(), 0.0001, "Equal dimensions should work")

	item.setOrientation(OrientationHWD)
	dim := item.dimension()
	require.InDelta(t, 5.0, dim[0], 0.0001, "Rotation with equal dimensions should work")
	require.InDelta(t, 5.0, dim[1], 0.0001, "Rotation with equal dimensions should work")
	require.InDelta(t, 5.0, dim[2], 0.0001, "Rotation with equal dimensions should work")
}
