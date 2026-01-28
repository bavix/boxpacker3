package boxpacker3

type Merit interface {
	Name() string

	Score(box *Container, position Pivot, dimension Dimension) MeritScore

	Better(a, b MeritScore) bool
}

type MeritScore struct {
	Position Pivot

	Contact float64

	Weighted float64

	Residual Dimension

	ExactAxes int

	MinGap float64

	Footprint float64
}

func (c *Container) meritOf() Merit { //nolint:ireturn // it hands back what it was given.
	if c.rules.Merit != nil {
		return c.rules.Merit
	}

	return ContactFirst
}

//nolint:gochecknoglobals // each is one fixed rule, exposed as a value.
var (
	ContactFirst Merit = &contactFirst{}

	CornerFirst Merit = &cornerFirst{}

	ResidualFit Merit = &residualFit{}

	WeightedContact Merit = &weightedContact{}
)

func Merits() []Merit {
	return []Merit{ContactFirst, CornerFirst, ResidualFit, WeightedContact}
}

func MeritByName(name string) (Merit, bool) { //nolint:ireturn // the caller asked for one by name.
	for _, merit := range Merits() {
		if merit.Name() == name {
			return merit, true
		}
	}

	return nil, false
}

func (c *Container) score(position Pivot, dimension Dimension) MeritScore {
	reading := MeritScore{
		Position:  position,
		Contact:   c.contactArea(position, dimension),
		Weighted:  0,
		Residual:  Dimension{},
		ExactAxes: 0,
		MinGap:    0,
		Footprint: dimension[WidthAxis] * dimension[HeightAxis],
	}

	first := true

	for _, axis := range allAxes() {
		gap := c.residualGap(position, dimension, axis)
		reading.Residual[axis] = gap

		if gap <= c.boxDimension(axis)*dimensionEpsilon {
			reading.ExactAxes++
		}

		if axis == DepthAxis {
			continue
		}

		if first || gap < reading.MinGap {
			reading.MinGap = gap
			first = false
		}
	}

	return reading
}

func positionBefore(a, b Pivot) (bool, bool) {
	for _, axis := range [3]Axis{DepthAxis, HeightAxis, WidthAxis} {
		if a[axis] != b[axis] {
			return a[axis] < b[axis], true
		}
	}

	return false, false
}

type contactFirst struct{}

func (*contactFirst) Name() string { return "contact-first" }

func (*contactFirst) Score(box *Container, position Pivot, dimension Dimension) MeritScore {
	return box.score(position, dimension)
}

func (*contactFirst) Better(a, b MeritScore) bool {
	if a.Contact != b.Contact {
		return a.Contact > b.Contact
	}

	if before, decided := positionBefore(a.Position, b.Position); decided {
		return before
	}

	if a.ExactAxes != b.ExactAxes {
		return a.ExactAxes > b.ExactAxes
	}

	if a.MinGap != b.MinGap {
		return a.MinGap < b.MinGap
	}

	return a.Footprint < b.Footprint
}

type cornerFirst struct{}

func (*cornerFirst) Name() string { return "corner-first" }

func (*cornerFirst) Score(box *Container, position Pivot, dimension Dimension) MeritScore {
	return box.score(position, dimension)
}

func (*cornerFirst) Better(a, b MeritScore) bool {
	if before, decided := positionBefore(a.Position, b.Position); decided {
		return before
	}

	if a.Contact != b.Contact {
		return a.Contact > b.Contact
	}

	return a.Footprint < b.Footprint
}

type residualFit struct{}

func (*residualFit) Name() string { return "residual-fit" }

func (*residualFit) Score(box *Container, position Pivot, dimension Dimension) MeritScore {
	return box.score(position, dimension)
}

func (*residualFit) Better(a, b MeritScore) bool {
	first, second := residualSum(a), residualSum(b)
	if first != second {
		return first < second
	}

	if a.Contact != b.Contact {
		return a.Contact > b.Contact
	}

	if before, decided := positionBefore(a.Position, b.Position); decided {
		return before
	}

	return a.Footprint < b.Footprint
}

func residualSum(reading MeritScore) float64 {
	return reading.Residual[WidthAxis] + reading.Residual[HeightAxis] + reading.Residual[DepthAxis]
}

const (
	weightFloor = 2.0
	weightBack  = 4.0
	weightLeft  = 2.0
)

type weightedContact struct{}

func (*weightedContact) Name() string { return "weighted-contact" }

func (*weightedContact) Score(box *Container, position Pivot, dimension Dimension) MeritScore {
	reading := box.score(position, dimension)
	reading.Weighted = reading.Contact + box.wallBonus(position, dimension)

	return reading
}

func (c *Container) wallBonus(position Pivot, dimension Dimension) float64 {
	bonus := 0.0

	if position[DepthAxis] <= c.depth*dimensionEpsilon {
		bonus += (weightFloor - 1) * dimension[WidthAxis] * dimension[HeightAxis]
	}

	if position[HeightAxis] <= c.height*dimensionEpsilon {
		bonus += (weightBack - 1) * dimension[WidthAxis] * dimension[DepthAxis]
	}

	if position[WidthAxis] <= c.width*dimensionEpsilon {
		bonus += (weightLeft - 1) * dimension[HeightAxis] * dimension[DepthAxis]
	}

	return bonus
}

func (*weightedContact) Better(a, b MeritScore) bool {
	if a.Weighted != b.Weighted {
		return a.Weighted > b.Weighted
	}

	if before, decided := positionBefore(a.Position, b.Position); decided {
		return before
	}

	if a.ExactAxes != b.ExactAxes {
		return a.ExactAxes > b.ExactAxes
	}

	return a.Footprint < b.Footprint
}
