package boxpacker3

import (
	"fmt"
	"math"
	"slices"
)

const dimensionEpsilon = 1e-9

func withinLimit(used, limit float64) bool {
	return used <= limit+math.Abs(limit)*dimensionEpsilon
}

type Box struct {
	id string

	width  float64
	height float64
	depth  float64

	outerWidth  float64
	outerHeight float64
	outerDepth  float64

	emptyWeight float64
	maxWeight   float64
	volume      float64

	maxLength float64

	accepts []string

	quantity int
}

type BoxSpec struct {
	ID                                  string
	OuterWidth, OuterHeight, OuterDepth float64
	InnerWidth, InnerHeight, InnerDepth float64
	EmptyWeight                         float64
	MaxWeight                           float64
	Quantity                            int

	Accepts []string
}

func NewBoxFromSpec(spec BoxSpec) (*Box, error) {
	spec = spec.withDefaults()

	err := spec.validate()
	if err != nil {
		return nil, err
	}

	box := newBox(spec.ID, spec.InnerWidth, spec.InnerHeight, spec.InnerDepth, spec.MaxWeight)
	box.quantity = max(spec.Quantity, 1)
	box.outerWidth = spec.OuterWidth
	box.outerHeight = spec.OuterHeight
	box.outerDepth = spec.OuterDepth
	box.emptyWeight = spec.EmptyWeight
	box.accepts = slices.Clone(spec.Accepts)

	return box, nil
}

func (s BoxSpec) withDefaults() BoxSpec {
	if s.InnerWidth == 0 {
		s.InnerWidth = s.OuterWidth
	}

	if s.InnerHeight == 0 {
		s.InnerHeight = s.OuterHeight
	}

	if s.InnerDepth == 0 {
		s.InnerDepth = s.OuterDepth
	}

	return s
}

func measurable(size float64) bool {
	return size > 0 && !math.IsInf(size, 0) && !math.IsNaN(size)
}

func (s BoxSpec) validateSides() error {
	sizes := map[string]float64{
		"outer width": s.OuterWidth, "outer height": s.OuterHeight, "outer depth": s.OuterDepth,
		"inner width": s.InnerWidth, "inner height": s.InnerHeight, "inner depth": s.InnerDepth,
	}

	for axis, size := range sizes {
		if !measurable(size) {
			return fmt.Errorf("%w: box %q has %s %g", ErrInvalidDimension, s.ID, axis, size)
		}
	}

	if volume := s.OuterWidth * s.OuterHeight * s.OuterDepth; math.IsInf(volume, 0) {
		return fmt.Errorf("%w: box %q is too large to measure", ErrInvalidDimension, s.ID)
	}

	if s.InnerWidth > s.OuterWidth || s.InnerHeight > s.OuterHeight || s.InnerDepth > s.OuterDepth {
		return fmt.Errorf("%w: box %q", ErrInnerExceedsOuter, s.ID)
	}

	return nil
}

func (s BoxSpec) validate() error {
	err := s.validateSides()
	if err != nil {
		return err
	}

	if s.EmptyWeight < 0 {
		return fmt.Errorf("%w: box %q has tare %g", ErrInvalidWeight, s.ID, s.EmptyWeight)
	}

	if s.Quantity < 0 {
		return fmt.Errorf("%w: box %q has quantity %d", ErrInvalidQuantity, s.ID, s.Quantity)
	}

	if s.MaxWeight < s.EmptyWeight {
		return fmt.Errorf("%w: box %q allows %g gross but weighs %g empty",
			ErrInvalidWeight, s.ID, s.MaxWeight, s.EmptyWeight)
	}

	if slices.Contains(s.Accepts, "") {
		return fmt.Errorf("%w: box %q accepts an empty class", ErrInvalidClass, s.ID)
	}

	return nil
}

func NewBox(id string, w, h, d, mw float64) *Box {
	return newBox(id, w, h, d, mw)
}

func newBox(id string, w, h, d, mw float64) *Box {
	//nolint:exhaustruct_v5
	return &Box{
		id:          id,
		width:       w,
		height:      h,
		depth:       d,
		outerWidth:  w,
		outerHeight: h,
		outerDepth:  d,
		maxWeight:   mw,
		maxLength:   max(w, h, d),
		volume:      w * h * d,
		quantity:    1,
	}
}

func NewBox2D(id string, w, h, mw float64) *Box {
	return NewBox(id, w, h, 1, mw)
}

func (b *Box) Flat() bool {
	return b.depth == 1
}

func (b *Box) Area() float64 {
	return b.width * b.height
}

func (b *Box) ID() string {
	return b.id
}

func (b *Box) Width() float64 {
	return b.width
}

func (b *Box) Height() float64 {
	return b.height
}

func (b *Box) Depth() float64 {
	return b.depth
}

func (b *Box) Volume() float64 {
	return b.volume
}

func (b *Box) MaxWeight() float64 {
	return b.maxWeight
}

func (b *Box) OuterWidth() float64 {
	return b.outerWidth
}

func (b *Box) OuterHeight() float64 {
	return b.outerHeight
}

func (b *Box) OuterDepth() float64 {
	return b.outerDepth
}

func (b *Box) Quantity() int {
	return b.quantity
}

func (b *Box) EmptyWeight() float64 {
	return b.emptyWeight
}

func (b *Box) Accepts() []string {
	return slices.Clone(b.accepts)
}
