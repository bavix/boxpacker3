package boxpacker3

type Rotation int

const (
	RotationBestFit Rotation = iota
	RotationKeepFlat
	RotationNever
)

func (r Rotation) String() string {
	switch r {
	case RotationKeepFlat:
		return "keep-flat"
	case RotationNever:
		return "never"
	case RotationBestFit:
	}

	return labelBestFit
}

func Rotations() []Rotation {
	return []Rotation{RotationBestFit, RotationKeepFlat, RotationNever}
}

type Orientation int

const (
	OrientationWHD Orientation = iota
	OrientationHWD
	OrientationHDW
	OrientationDHW
	OrientationDWH
	OrientationWDH
)

//nolint:gochecknoglobals
var rotationMatrix = [6][3]int{
	OrientationWHD: {0, 1, 2},
	OrientationHWD: {1, 0, 2},
	OrientationHDW: {1, 2, 0},
	OrientationDHW: {2, 1, 0},
	OrientationDWH: {2, 0, 1},
	OrientationWDH: {0, 2, 1},
}

type Axis int

const (
	WidthAxis Axis = iota
	HeightAxis
	DepthAxis
)

type Pivot [3]float64

type Dimension [3]float64
