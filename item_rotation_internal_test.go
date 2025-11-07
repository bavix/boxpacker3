package boxpacker3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestItem_GetDimension_AllRotations tests all 6 rotation types.
func TestItem_GetDimension_AllRotations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		rotationType   RotationType
		expectedWidth  float64
		expectedHeight float64
		expectedDepth  float64
	}{
		{
			name:           "RotationTypeWhd",
			rotationType:   RotationTypeWhd,
			expectedWidth:  10, // w
			expectedHeight: 20, // h
			expectedDepth:  30, // d
		},
		{
			name:           "RotationTypeHwd",
			rotationType:   RotationTypeHwd,
			expectedWidth:  20, // h
			expectedHeight: 10, // w
			expectedDepth:  30, // d
		},
		{
			name:           "RotationTypeHdw",
			rotationType:   RotationTypeHdw,
			expectedWidth:  20, // h
			expectedHeight: 30, // d
			expectedDepth:  10, // w
		},
		{
			name:           "RotationTypeDhw",
			rotationType:   RotationTypeDhw,
			expectedWidth:  30, // d
			expectedHeight: 20, // h
			expectedDepth:  10, // w
		},
		{
			name:           "RotationTypeDwh",
			rotationType:   RotationTypeDwh,
			expectedWidth:  30, // d
			expectedHeight: 10, // w
			expectedDepth:  20, // h
		},
		{
			name:           "RotationTypeWdh",
			rotationType:   RotationTypeWdh,
			expectedWidth:  10, // w
			expectedHeight: 30, // d
			expectedDepth:  20, // h
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Create item with distinct dimensions: width=10, height=20, depth=30
			item := NewItem("test", 10, 20, 30, 1)
			item.setRotationType(tc.rotationType)
			dim := item.GetDimension()

			require.InDelta(t, tc.expectedWidth, dim[0], 0.0001, "Width should match")
			require.InDelta(t, tc.expectedHeight, dim[1], 0.0001, "Height should match")
			require.InDelta(t, tc.expectedDepth, dim[2], 0.0001, "Depth should match")
		})
	}
}

// TestItem_GetDimension_Consistency tests that GetDimension is consistent with GetWidth/Height/Depth for default rotation.
func TestItem_GetDimension_Consistency(t *testing.T) {
	t.Parallel()

	item := NewItem("test", 10, 20, 30, 1)
	item.setRotationType(RotationTypeWhd) // Default rotation

	dim := item.GetDimension()

	require.InDelta(t, item.GetWidth(), dim[0], 0.0001, "Width from GetDimension should match GetWidth")
	require.InDelta(t, item.GetHeight(), dim[1], 0.0001, "Height from GetDimension should match GetHeight")
	require.InDelta(t, item.GetDepth(), dim[2], 0.0001, "Depth from GetDimension should match GetDepth")
}

// TestItem_RotationMatrix_AllCombinations tests all rotation combinations with different item sizes.
func TestItem_RotationMatrix_AllCombinations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		width        float64
		height       float64
		depth        float64
		rotationType RotationType
		expected     [3]float64
	}{
		// Small values
		{"Small_Whd", 1, 2, 3, RotationTypeWhd, [3]float64{1, 2, 3}},
		{"Small_Hwd", 1, 2, 3, RotationTypeHwd, [3]float64{2, 1, 3}},
		{"Small_Hdw", 1, 2, 3, RotationTypeHdw, [3]float64{2, 3, 1}},
		{"Small_Dhw", 1, 2, 3, RotationTypeDhw, [3]float64{3, 2, 1}},
		{"Small_Dwh", 1, 2, 3, RotationTypeDwh, [3]float64{3, 1, 2}},
		{"Small_Wdh", 1, 2, 3, RotationTypeWdh, [3]float64{1, 3, 2}},

		// Large values
		{"Large_Whd", 100, 200, 300, RotationTypeWhd, [3]float64{100, 200, 300}},
		{"Large_Hwd", 100, 200, 300, RotationTypeHwd, [3]float64{200, 100, 300}},
		{"Large_Hdw", 100, 200, 300, RotationTypeHdw, [3]float64{200, 300, 100}},
		{"Large_Dhw", 100, 200, 300, RotationTypeDhw, [3]float64{300, 200, 100}},
		{"Large_Dwh", 100, 200, 300, RotationTypeDwh, [3]float64{300, 100, 200}},
		{"Large_Wdh", 100, 200, 300, RotationTypeWdh, [3]float64{100, 300, 200}},

		// Equal values
		{"Equal_Whd", 5, 5, 5, RotationTypeWhd, [3]float64{5, 5, 5}},
		{"Equal_Hwd", 5, 5, 5, RotationTypeHwd, [3]float64{5, 5, 5}},
		{"Equal_Hdw", 5, 5, 5, RotationTypeHdw, [3]float64{5, 5, 5}},

		// Very small values (edge case)
		{"Tiny_Whd", 0.1, 0.2, 0.3, RotationTypeWhd, [3]float64{0.1, 0.2, 0.3}},
		{"Tiny_Hwd", 0.1, 0.2, 0.3, RotationTypeHwd, [3]float64{0.2, 0.1, 0.3}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			item := NewItem("test", tc.width, tc.height, tc.depth, 1)
			item.setRotationType(tc.rotationType)

			dim := item.GetDimension()

			require.InDelta(t, tc.expected[0], dim[0], 0.0001, "Width should match")
			require.InDelta(t, tc.expected[1], dim[1], 0.0001, "Height should match")
			require.InDelta(t, tc.expected[2], dim[2], 0.0001, "Depth should match")
		})
	}
}

// TestItem_RotationMatrix_BoundaryValues tests boundary values for rotation matrix indices.
func TestItem_RotationMatrix_BoundaryValues(t *testing.T) {
	t.Parallel()
	// Test with minimum valid rotation
	item := NewItem("test", 10, 20, 30, 1)
	item.setRotationType(RotationTypeWhd)
	dim := item.GetDimension()
	require.InDelta(t, 10.0, dim[0], 0.0001, "Minimum rotation should work")

	// Test with maximum valid rotation
	item.setRotationType(RotationTypeWdh)
	dim = item.GetDimension()
	require.InDelta(t, 10.0, dim[0], 0.0001, "Maximum rotation should work")
	require.InDelta(t, 30.0, dim[1], 0.0001, "Maximum rotation should work")
	require.InDelta(t, 20.0, dim[2], 0.0001, "Maximum rotation should work")
}

// TestItem_WhdArray_EdgeCases tests edge cases for whd array access.
func TestItem_WhdArray_EdgeCases(t *testing.T) {
	t.Parallel()
	// Test with very small values
	item := NewItem("tiny", 0.001, 0.002, 0.003, 0.1)
	require.InDelta(t, 0.001, item.GetWidth(), 0.0001, "Very small width should work")
	require.InDelta(t, 0.002, item.GetHeight(), 0.0001, "Very small height should work")
	require.InDelta(t, 0.003, item.GetDepth(), 0.0001, "Very small depth should work")

	// Test with very large values
	item = NewItem("large", 10000, 20000, 30000, 1000)
	require.InDelta(t, 10000.0, item.GetWidth(), 0.0001, "Very large width should work")
	require.InDelta(t, 20000.0, item.GetHeight(), 0.0001, "Very large height should work")
	require.InDelta(t, 30000.0, item.GetDepth(), 0.0001, "Very large depth should work")

	// Test with equal dimensions
	item = NewItem("cube", 5, 5, 5, 1)
	require.InDelta(t, 5.0, item.GetWidth(), 0.0001, "Equal dimensions should work")
	require.InDelta(t, 5.0, item.GetHeight(), 0.0001, "Equal dimensions should work")
	require.InDelta(t, 5.0, item.GetDepth(), 0.0001, "Equal dimensions should work")

	// Test rotation with equal dimensions
	item.setRotationType(RotationTypeHwd)
	dim := item.GetDimension()
	require.InDelta(t, 5.0, dim[0], 0.0001, "Rotation with equal dimensions should work")
	require.InDelta(t, 5.0, dim[1], 0.0001, "Rotation with equal dimensions should work")
	require.InDelta(t, 5.0, dim[2], 0.0001, "Rotation with equal dimensions should work")
}

// TestItem_RotationMatrix_IndexBoundaries tests that matrix indices are always valid.
func TestItem_RotationMatrix_IndexBoundaries(t *testing.T) {
	t.Parallel()

	item := NewItem("test", 10, 20, 30, 1)

	// Test all valid rotation types don't cause index out of bounds
	for rt := RotationTypeWhd; rt <= RotationTypeWdh; rt++ {
		item.setRotationType(rt)
		dim := item.GetDimension()

		// All dimensions should be positive and from the original set {10, 20, 30}
		require.Greater(t, dim[0], 0.0, "Width should be positive")
		require.Greater(t, dim[1], 0.0, "Height should be positive")
		require.Greater(t, dim[2], 0.0, "Depth should be positive")

		// All dimensions should be one of the original values
		require.Contains(t, []float64{10.0, 20.0, 30.0}, dim[0], "Width should be from original dimensions")
		require.Contains(t, []float64{10.0, 20.0, 30.0}, dim[1], "Height should be from original dimensions")
		require.Contains(t, []float64{10.0, 20.0, 30.0}, dim[2], "Depth should be from original dimensions")
	}
}

// TestItem_GetDimension_MatrixDirectAccess tests direct matrix access without intermediate arrays.
func TestItem_GetDimension_MatrixDirectAccess(t *testing.T) {
	t.Parallel()

	item := NewItem("test", 10, 20, 30, 1)

	// Test that GetDimension uses whd array directly (no intermediate array creation)
	// This is verified by checking that all rotations work correctly
	rotations := []RotationType{
		RotationTypeWhd,
		RotationTypeHwd,
		RotationTypeHdw,
		RotationTypeDhw,
		RotationTypeDwh,
		RotationTypeWdh,
	}

	expectedResults := [][3]float64{
		{10, 20, 30}, // Whd
		{20, 10, 30}, // Hwd
		{20, 30, 10}, // Hdw
		{30, 20, 10}, // Dhw
		{30, 10, 20}, // Dwh
		{10, 30, 20}, // Wdh
	}

	for i, rt := range rotations {
		item.setRotationType(rt)
		dim := item.GetDimension()
		require.InDelta(t, expectedResults[i][0], dim[0], 0.0001, "Rotation %d width should match", i)
		require.InDelta(t, expectedResults[i][1], dim[1], 0.0001, "Rotation %d height should match", i)
		require.InDelta(t, expectedResults[i][2], dim[2], 0.0001, "Rotation %d depth should match", i)
	}
}
