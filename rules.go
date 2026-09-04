package boxpacker3

import "context"

type Rules struct {
	MinSupportRatio float64

	Admission []AdmissionRule

	Placement []PlacementRule

	Pair []PairRule

	Merit Merit

	Filler Filler

	FreeSpaceCorners bool
}

type PlacementRule interface {
	Allow(box *Container, item *Item, position Pivot, dimension Dimension) bool
}

type PlacementRuleFunc func(box *Container, item *Item, position Pivot, dimension Dimension) bool

func (f PlacementRuleFunc) Allow(box *Container, item *Item, position Pivot, dimension Dimension) bool {
	return f(box, item, position, dimension)
}

type AdmissionRule interface {
	Allow(box *Container, item *Item) bool
}

type AdmissionRuleFunc func(box *Container, item *Item) bool

func (f AdmissionRuleFunc) Allow(box *Container, item *Item) bool {
	return f(box, item)
}

type PairRule interface {
	Together(a, b *Item) bool
}

type PairRuleFunc func(a, b *Item) bool

func (f PairRuleFunc) Together(a, b *Item) bool {
	return f(a, b)
}

func (r Rules) merged(other Rules) Rules {
	merit := r.Merit
	if other.Merit != nil {
		merit = other.Merit
	}

	filler := r.Filler
	if other.Filler != nil {
		filler = other.Filler
	}

	return Rules{
		MinSupportRatio:  max(r.MinSupportRatio, other.MinSupportRatio),
		Admission:        append(append([]AdmissionRule(nil), r.Admission...), other.Admission...),
		Placement:        append(append([]PlacementRule(nil), r.Placement...), other.Placement...),
		Pair:             append(append([]PairRule(nil), r.Pair...), other.Pair...),
		Merit:            merit,
		Filler:           filler,
		FreeSpaceCorners: r.FreeSpaceCorners || other.FreeSpaceCorners,
	}
}

func (r Rules) clamped() Rules {
	r.MinSupportRatio = min(max(r.MinSupportRatio, 0), 1)

	return r
}

func (r Rules) permits(box *Container, item *piece, position Pivot, dimension Dimension) bool {
	for _, rule := range r.Placement {
		if !rule.Allow(box, item.Item, position, dimension) {
			return false
		}
	}

	return true
}

func (r Rules) admits(box *Container, item *piece) bool {
	for _, rule := range r.Admission {
		if !rule.Allow(box, item.Item) {
			return false
		}
	}

	return true
}

func (r Rules) together(a, b *piece) bool {
	for _, rule := range r.Pair {
		if !rule.Together(a.Item, b.Item) || !rule.Together(b.Item, a.Item) {
			return false
		}
	}

	return true
}

type Filler interface {
	Name() string
	Fill(ctx context.Context, box *Container, items []Instance) ([]Instance, error)
}
