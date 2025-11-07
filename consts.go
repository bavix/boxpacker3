package boxpacker3

type RotationType int

const (
	RotationTypeWhd RotationType = iota
	RotationTypeHwd
	RotationTypeHdw
	RotationTypeDhw
	RotationTypeDwh
	RotationTypeWdh
)

// rotationMatrix defines rotation matrices for all 6 possible 3D rotations of an item.
//
// Each item stores its dimensions as whd [3]float64 = [width, height, depth] = [0, 1, 2].
// The rotation matrix maps the original whd array indices to new positions for each axis.
//
// Structure:
//   - Each row represents one rotation type (RotationTypeWhd through RotationTypeWdh).
//   - Each row contains 3 integers: [WidthAxis mapping, HeightAxis mapping, DepthAxis mapping].
//   - The integers are indices into the original whd array [0=width, 1=height, 2=depth].
//
// How it works:
//   - For a given rotation type, matrix[i] gives the index in whd array for axis i.
//   - WidthAxis (0) -> whd[matrix[0]] gives the width dimension after rotation.
//   - HeightAxis (1) -> whd[matrix[1]] gives the height dimension after rotation.
//   - DepthAxis (2) -> whd[matrix[2]] gives the depth dimension after rotation.
//
// Examples:
//   - RotationTypeWhd {0, 1, 2}: [w, h, d] -> [w, h, d] (no rotation, original orientation)
//   - RotationTypeHwd {1, 0, 2}: [w, h, d] -> [h, w, d] (width and height swapped)
//   - RotationTypeHdw {1, 2, 0}: [w, h, d] -> [h, d, w] (cyclic rotation)
//   - RotationTypeDhw {2, 1, 0}: [w, h, d] -> [d, h, w] (width and depth swapped)
//   - RotationTypeDwh {2, 0, 1}: [w, h, d] -> [d, w, h] (cyclic rotation)
//   - RotationTypeWdh {0, 2, 1}: [w, h, d] -> [w, d, h] (height and depth swapped)
//
// Usage in code:
//   - GetDimension(): matrix := rotationMatrix[i.rotationType]; dim[axis] = i.whd[matrix[axis]]
//   - PutItem(): matrix := rotationMatrix[rt]; itemWidth = item.whd[matrix[WidthAxis]]
//
// This approach avoids switch statements and provides O(1) lookup for rotation calculations.
//
//nolint:gochecknoglobals
var rotationMatrix = [6][3]int{
	RotationTypeWhd: {0, 1, 2}, // [w, h, d] -> [w, h, d] (no rotation)
	RotationTypeHwd: {1, 0, 2}, // [w, h, d] -> [h, w, d] (swap width/height)
	RotationTypeHdw: {1, 2, 0}, // [w, h, d] -> [h, d, w] (cyclic: w->d, h->w, d->h)
	RotationTypeDhw: {2, 1, 0}, // [w, h, d] -> [d, h, w] (swap width/depth)
	RotationTypeDwh: {2, 0, 1}, // [w, h, d] -> [d, w, h] (cyclic: w->h, h->d, d->w)
	RotationTypeWdh: {0, 2, 1}, // [w, h, d] -> [w, d, h] (swap height/depth)
}

type Axis int

const (
	WidthAxis Axis = iota
	HeightAxis
	DepthAxis
)

type Pivot [3]float64

type Dimension [3]float64

type PackingStrategy int

const (
	// StrategyMinimizeBoxes is the default strategy that minimizes the number of boxes used.
	// It sorts items by volume in descending order (largest first) and uses First Fit algorithm
	// to place each item in the first box where it fits.
	// First Fit Decreasing is better for minimizing box count than Best Fit.
	StrategyMinimizeBoxes PackingStrategy = iota

	// StrategyGreedy is a greedy packing strategy (First Fit with ascending sort).
	// It sorts items by volume in ascending order (smallest first) and uses First Fit algorithm
	// to place each item in the first box where it fits.
	// Simple and fast, but may use more boxes than optimal strategies.
	StrategyGreedy

	// StrategyBestFit is a best-fit strategy.
	// For each item, it finds the box with the smallest remaining space that can accommodate the item.
	// This minimizes wasted space but requires checking all boxes for each item.
	StrategyBestFit

	// StrategyBestFitDecreasing is a best-fit decreasing strategy.
	// Items are sorted by volume in descending order (largest first),
	// and for each item, it finds the box with the smallest remaining space.
	// Typically provides 2-5% better space utilization than FFD.
	StrategyBestFitDecreasing

	// StrategyNextFit is a next-fit strategy.
	// Items are placed in the current box if it fits, otherwise a new box is used.
	// Simpler than First Fit but may use more boxes.
	StrategyNextFit

	// StrategyWorstFit is a worst-fit strategy.
	// For each item, it finds the box with the largest remaining space that can accommodate the item.
	// This can help distribute items more evenly across boxes.
	StrategyWorstFit

	// StrategyAlmostWorstFit is an almost-worst-fit strategy.
	// Similar to Worst Fit, but excludes boxes that are too large (almost empty).
	// This prevents items from being placed in boxes that are nearly empty.
	StrategyAlmostWorstFit
)
