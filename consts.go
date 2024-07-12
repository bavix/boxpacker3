package boxpacker3

// RotationType represents the type of rotation for an item.
type RotationType int

// RotationTypeWhd represents the rotation type where the width is the longest dimension.
const (
	RotationTypeWhd RotationType = iota
	// RotationTypeHwd represents the rotation type where the height is the longest dimension.
	RotationTypeHwd
	// RotationTypeHdw represents the rotation type where the depth is the longest dimension.
	RotationTypeHdw
	// RotationTypeDhw represents the rotation type where the depth is the longest dimension.
	RotationTypeDhw
	// RotationTypeDwh represents the rotation type where the width is the longest dimension.
	RotationTypeDwh
	// RotationTypeWdh represents the rotation type where the height is the longest dimension.
	RotationTypeWdh
)

// Axis represents the axis of a dimension.
type Axis int

// WidthAxis represents the width axis.
const (
	WidthAxis Axis = iota
	// HeightAxis represents the height axis.
	HeightAxis
	// DepthAxis represents the depth axis.
	DepthAxis
)

// Pivot represents the position of an item within a box.
type Pivot [3]float64

// Dimension represents the dimensions of an item or a box.
type Dimension [3]float64
