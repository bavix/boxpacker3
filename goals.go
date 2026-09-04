package boxpacker3

import (
	"math"
	"strings"
)

type Goal interface {
	Name() string
	Compare(a, b *Result) int
}

type Metrics struct {
	Unpacked int

	Boxes int

	BoxVolume float64

	ItemsVolume float64

	Fill float64

	WeightSpread float64
}

func MetricsOf(result *Result) Metrics {
	if result == nil {
		return Metrics{Unpacked: 0, Boxes: 0, BoxVolume: 0, ItemsVolume: 0, Fill: 0, WeightSpread: 0}
	}

	metrics := Metrics{
		Unpacked:     len(result.Unpacked),
		Boxes:        len(result.Boxes),
		BoxVolume:    0,
		ItemsVolume:  0,
		Fill:         0,
		WeightSpread: 0,
	}

	for _, box := range result.Boxes {
		metrics.BoxVolume += box.Box.volume
		metrics.ItemsVolume += box.Stats.ItemsVolume
		metrics.Fill += box.Stats.Fill
	}

	if metrics.Boxes > 0 {
		metrics.Fill /= float64(metrics.Boxes)
	}

	metrics.WeightSpread = weightSpread(result.Boxes)

	return metrics
}

func weightSpread(boxes []PackedBox) float64 {
	if len(boxes) <= 1 {
		return 0
	}

	var sum float64

	for _, box := range boxes {
		sum += box.Stats.ItemsWeight
	}

	mean := sum / float64(len(boxes))

	var variance float64

	for _, box := range boxes {
		diff := box.Stats.ItemsWeight - mean
		variance += diff * diff
	}

	return math.Sqrt(variance / float64(len(boxes)))
}

const epsilon = 1e-9

func sameReading(a, b float64) bool {
	return math.Abs(a-b) <= epsilon*max(math.Abs(a), math.Abs(b))
}

type Criterion struct {
	Name string

	Measure func(*Result) float64

	Higher bool
}

type Term struct {
	Criterion

	Weight float64
}

type lexicographic struct {
	name     string
	criteria []Criterion
}

func (g lexicographic) Name() string {
	return g.name
}

func (g lexicographic) Compare(a, b *Result) int {
	if decided, order := decideMissing(a, b); decided {
		return order
	}

	for _, criterion := range g.criteria {
		first, second := criterion.Measure(a), criterion.Measure(b)
		if sameReading(first, second) {
			continue
		}

		if criterion.Higher == (first > second) {
			return -1
		}

		return 1
	}

	return 0
}

func Lexicographic(name string, criteria ...Criterion) Goal { //nolint:ireturn // the caller passes it to a runner.
	return lexicographic{name: name, criteria: criteria}
}

type weighted struct {
	name  string
	terms []Term
}

func (g weighted) Name() string {
	return g.name
}

func (g weighted) Compare(a, b *Result) int {
	if decided, order := decideMissing(a, b); decided {
		return order
	}

	first, second := g.score(a), g.score(b)
	if sameReading(first, second) {
		return 0
	}

	if first < second {
		return -1
	}

	return 1
}

func (g weighted) score(result *Result) float64 {
	total := 0.0

	for _, term := range g.terms {
		reading := term.Weight * term.Measure(result)
		if term.Higher {
			reading = -reading
		}

		total += reading
	}

	return total
}

func Weighted(name string, terms ...Term) Goal { //nolint:ireturn // the caller passes it to a runner.
	return weighted{name: name, terms: terms}
}

type costGoal struct {
	name string
	cost func(*Result) float64
}

func (g costGoal) Name() string {
	return g.name
}

func (g costGoal) Compare(a, b *Result) int {
	if decided, order := decideMissing(a, b); decided {
		return order
	}

	first, second := g.cost(a), g.cost(b)
	if sameReading(first, second) {
		return 0
	}

	if first < second {
		return -1
	}

	return 1
}

func CostGoal(name string, cost func(*Result) float64) Goal { //nolint:ireturn // the caller passes it to a runner.
	return costGoal{name: name, cost: cost}
}

func decideMissing(a, b *Result) (bool, int) {
	switch {
	case a == nil && b == nil:
		return true, 0
	case a == nil:
		return true, 1
	case b == nil:
		return true, -1
	}

	return false, 0
}

type Tariff struct {
	PerBox map[string]float64

	PerKg float64

	DimDivisor float64

	Minimum float64
}

func (t Tariff) Price(box PackedBox) float64 {
	chargeable := box.Stats.GrossWeight

	if t.DimDivisor > 0 {
		volumetric := box.Box.outerWidth * box.Box.outerHeight * box.Box.outerDepth / t.DimDivisor
		chargeable = max(chargeable, volumetric)
	}

	return t.PerBox[box.ID()] + t.PerKg*chargeable
}

func (t Tariff) Total(result *Result) float64 {
	if result == nil {
		return math.Inf(1)
	}

	total := 0.0

	for _, box := range result.Boxes {
		total += t.Price(box)
	}

	return max(total, t.Minimum)
}

func ShippingCost(name string, tariff Tariff) Goal { //nolint:ireturn // the caller passes it to a runner.
	return Lexicographic(name,
		Criterion{Name: criterionUnpacked, Measure: unpackedCount, Higher: false},
		Criterion{Name: "cost", Measure: tariff.Total, Higher: false},
	)
}

func unpackedCount(result *Result) float64 {
	return float64(len(result.Unpacked))
}

func boxCount(result *Result) float64 {
	return float64(len(result.Boxes))
}

func boxVolume(result *Result) float64 {
	return MetricsOf(result).BoxVolume
}

func fillRate(result *Result) float64 {
	return MetricsOf(result).Fill
}

func loadSpread(result *Result) float64 {
	return weightSpread(result.Boxes)
}

const (
	criterionUnpacked = "unpacked"
	criterionBoxes    = "boxes"
	criterionVolume   = "box volume"
)

const (
	nameFewestBoxes    = "FewestBoxes"
	nameMostItems      = "MostItems"
	nameLeastVolume    = "LeastVolume"
	nameHighestFill    = "HighestFill"
	nameBalancedWeight = "BalancedWeight"
)

//nolint:gochecknoglobals // each is one fixed goal, exposed as a value.
var (
	FewestBoxes = Lexicographic(nameFewestBoxes,
		Criterion{Name: criterionUnpacked, Measure: unpackedCount, Higher: false},
		Criterion{Name: criterionBoxes, Measure: boxCount, Higher: false},
		Criterion{Name: criterionVolume, Measure: boxVolume, Higher: false},
	)

	MostItems = Lexicographic(nameMostItems,
		Criterion{Name: criterionUnpacked, Measure: unpackedCount, Higher: false},
	)

	LeastVolume = Lexicographic(nameLeastVolume,
		Criterion{Name: criterionUnpacked, Measure: unpackedCount, Higher: false},
		Criterion{Name: criterionVolume, Measure: boxVolume, Higher: false},
		Criterion{Name: criterionBoxes, Measure: boxCount, Higher: false},
	)

	HighestFill = Lexicographic(nameHighestFill,
		Criterion{Name: criterionUnpacked, Measure: unpackedCount, Higher: false},
		Criterion{Name: "fill", Measure: fillRate, Higher: true},
	)

	BalancedWeight = Lexicographic(nameBalancedWeight,
		Criterion{Name: criterionUnpacked, Measure: unpackedCount, Higher: false},
		Criterion{Name: "weight spread", Measure: loadSpread, Higher: false},
		Criterion{Name: criterionBoxes, Measure: boxCount, Higher: false},
	)
)

func Goals() []Goal {
	return []Goal{FewestBoxes, MostItems, LeastVolume, HighestFill, BalancedWeight}
}

func GoalNames() []string {
	names := make([]string, 0, len(Goals()))

	for _, goal := range Goals() {
		names = append(names, goal.Name())
	}

	return names
}

func GoalByName(name string) (Goal, bool) { //nolint:ireturn // the caller asked for one by name.
	for _, goal := range Goals() {
		if strings.EqualFold(goal.Name(), name) {
			return goal, true
		}
	}

	return nil, false
}

func prefers(goal Goal, candidate, best *Result) bool {
	if best == nil {
		return candidate != nil
	}

	return goal.Compare(candidate, best) < 0
}

func countUsedBoxes(boxes []*Container) int {
	return len(usedBoxes(boxes))
}
