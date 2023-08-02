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

type Axis int

const (
	WidthAxis Axis = iota
	HeightAxis
	DepthAxis
)

type (
	Pivot     [3]float64
	Dimension [3]float64
)
