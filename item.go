package boxpacker3

import (
	"fmt"
	"math"
	"slices"
)

const minDimension = 1e-10

type Item struct {
	id     string
	whd    [3]float64
	weight float64
	volume float64

	maxLength float64

	rotation     Rotation
	verticalAxes []Axis
	rotations    []Orientation
	oriented     [6]Dimension
	group        string

	maxLoadOnTop float64
	nothingOnTop bool

	class        string
	separateFrom []string
	quantity     int
}

type ItemSpec struct {
	ID                   string
	Width, Height, Depth float64
	Weight               float64
	Rotation             Rotation
	VerticalAxes         []Axis
	Group                string

	Quantity int

	MaxLoadOnTop float64

	NothingOnTop bool

	Class string

	SeparateFrom []string
}

func NewItemFromSpec(spec ItemSpec) (*Item, error) {
	err := spec.validate()
	if err != nil {
		return nil, err
	}

	item := newItem(spec.ID, spec.Width, spec.Height, spec.Depth, spec.Weight, spec.Rotation)
	item.group = spec.Group
	item.maxLoadOnTop = spec.MaxLoadOnTop
	item.nothingOnTop = spec.NothingOnTop
	item.class = spec.Class
	item.separateFrom = slices.Clone(spec.SeparateFrom)
	item.quantity = max(spec.Quantity, 1)

	if len(spec.VerticalAxes) > 0 {
		item.verticalAxes = slices.Clone(spec.VerticalAxes)
		item.rotations = distinctRotations(item)
	}

	return item, nil
}

func (s ItemSpec) validate() error {
	err := s.validateSides()
	if err != nil {
		return err
	}

	if s.Weight < 0 || math.IsInf(s.Weight, 0) || math.IsNaN(s.Weight) {
		return fmt.Errorf("%w: item %q has weight %g", ErrInvalidWeight, s.ID, s.Weight)
	}

	if s.Quantity < 0 {
		return fmt.Errorf("%w: item %q has quantity %d", ErrInvalidQuantity, s.ID, s.Quantity)
	}

	if s.MaxLoadOnTop < 0 || math.IsNaN(s.MaxLoadOnTop) {
		return fmt.Errorf("%w: item %q may carry %g",
			ErrInvalidWeight, s.ID, s.MaxLoadOnTop)
	}

	err = s.validateSeparation()
	if err != nil {
		return err
	}

	return s.validateAxes()
}

func (s ItemSpec) validateSeparation() error {
	if slices.Contains(s.SeparateFrom, "") {
		return fmt.Errorf("%w: item %q names an empty class", ErrInvalidClass, s.ID)
	}

	return nil
}

func (s ItemSpec) validateSides() error {
	for axis, size := range map[string]float64{"width": s.Width, "height": s.Height, "depth": s.Depth} {
		if !measurable(size) {
			return fmt.Errorf("%w: item %q has %s %g", ErrInvalidDimension, s.ID, axis, size)
		}
	}

	if volume := s.Width * s.Height * s.Depth; math.IsInf(volume, 0) {
		return fmt.Errorf("%w: item %q is too large to measure", ErrInvalidDimension, s.ID)
	}

	return nil
}

func (s ItemSpec) validateAxes() error {
	for _, axis := range s.VerticalAxes {
		if axis != WidthAxis && axis != HeightAxis && axis != DepthAxis {
			return fmt.Errorf("%w: item %q names axis %d", ErrInvalidAxis, s.ID, axis)
		}
	}

	return nil
}

func NewItem(id string, w, h, d, wg float64) *Item {
	return newItem(id, w, h, d, wg, RotationBestFit)
}

func newItem(id string, w, h, d, wg float64, rotation Rotation) *Item {
	//nolint:exhaustruct_v5
	item := &Item{
		id:           id,
		whd:          [3]float64{w, h, d},
		weight:       wg,
		volume:       w * h * d,
		maxLength:    max(w, h, d),
		rotation:     rotation,
		verticalAxes: verticalAxesFor(rotation),
		quantity:     1,
	}

	for rotation := range item.oriented {
		matrix := rotationMatrix[rotation]
		item.oriented[rotation] = Dimension{
			item.whd[matrix[WidthAxis]],
			item.whd[matrix[HeightAxis]],
			item.whd[matrix[DepthAxis]],
		}
	}

	item.rotations = distinctRotations(item)

	return item
}

func verticalAxesFor(rotation Rotation) []Axis {
	if rotation == RotationBestFit {
		return []Axis{WidthAxis, HeightAxis, DepthAxis}
	}

	return []Axis{DepthAxis}
}

func verticalEdge(rotation Orientation) Axis {
	return Axis(rotationMatrix[rotation][DepthAxis])
}

func candidateRotations(item *Item) []Orientation {
	if item.rotation == RotationNever {
		return []Orientation{OrientationWHD}
	}

	all := []Orientation{
		OrientationWHD, OrientationHWD, OrientationHDW,
		OrientationDHW, OrientationDWH, OrientationWDH,
	}

	candidates := make([]Orientation, 0, len(all))

	for _, rotation := range all {
		if slices.Contains(item.verticalAxes, verticalEdge(rotation)) {
			candidates = append(candidates, rotation)
		}
	}

	return candidates
}

func distinctRotations(item *Item) []Orientation {
	candidates := candidateRotations(item)

	distinct := make([]Orientation, 0, len(candidates))
	seen := make([]Dimension, 0, len(candidates))

	for _, rotation := range candidates {
		dimension := rotatedDimension(item, rotation)
		if slices.Contains(seen, dimension) {
			continue
		}

		seen = append(seen, dimension)
		distinct = append(distinct, rotation)
	}

	return distinct
}

func NewItem2D(id string, w, h, wg float64) *Item {
	return newItem(id, w, h, 1, wg, RotationKeepFlat)
}

func (i *Item) Flat() bool {
	return i.rotation == RotationKeepFlat
}

func (i *Item) Rotation() Rotation {
	return i.rotation
}

func (i *Item) Group() string {
	return i.group
}

func (i *Item) Orientations() []Dimension {
	orientations := make([]Dimension, 0, len(i.rotations))

	for _, rotation := range i.rotations {
		orientations = append(orientations, rotatedDimension(i, rotation))
	}

	return orientations
}

func (i *Item) MaxLoadOnTop() float64 {
	return i.maxLoadOnTop
}

func (i *Item) NothingOnTop() bool {
	return i.nothingOnTop
}

func (i *Item) Class() string {
	return i.class
}

func (i *Item) SeparateFrom() []string {
	return slices.Clone(i.separateFrom)
}

func (i *Item) travelsWith(other *Item) bool {
	if i == nil || other == nil {
		return true
	}

	return !slices.Contains(i.separateFrom, other.class) &&
		!slices.Contains(other.separateFrom, i.class)
}

func (i *Item) VerticalAxes() []Axis {
	return slices.Clone(i.verticalAxes)
}

func (i *Item) ID() string {
	return i.id
}

func (i *Item) Width() float64 {
	return i.whd[0]
}

func (i *Item) Height() float64 {
	return i.whd[1]
}

func (i *Item) Depth() float64 {
	return i.whd[2]
}

func (i *Item) Volume() float64 {
	return i.volume
}

func (i *Item) Weight() float64 {
	return i.weight
}

type piece struct {
	*Item

	index       int
	position    Pivot
	orientation Orientation
	reason      Reason
}

func newPiece(item *Item, index int) *piece {
	return &piece{Item: item, index: index, position: Pivot{}, orientation: OrientationWHD, reason: ReasonNoRoom}
}

func piecesOf(items []*Item) []*piece {
	pieces := make([]*piece, 0, len(items))

	for _, item := range items {
		if item == nil {
			continue
		}

		for index := range max(item.quantity, 1) {
			pieces = append(pieces, newPiece(item, index))
		}
	}

	return pieces
}

func (p *piece) setOrientation(orientation Orientation) {
	p.orientation = orientation
}

func (p *piece) dimension() Dimension {
	return rotatedDimension(p.Item, p.orientation)
}

func (p *piece) intersects(other *piece) bool {
	if p == nil || other == nil {
		return false
	}

	return overlaps(p.position, p.dimension(), other.position, other.dimension())
}

func (p *piece) packed() PackedItem {
	return PackedItem{
		Item:        p.Item,
		Index:       p.index,
		Position:    p.position,
		Dimension:   p.dimension(),
		Orientation: p.orientation,
	}
}

func (i *Item) Quantity() int {
	return i.quantity
}

func (p *piece) setReason(reason Reason) {
	p.reason = reason
}
