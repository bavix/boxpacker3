package boxpacker3

import (
	"cmp"
	"context"
	"slices"
	"strings"
)

const (
	labelBestFit  = "best-fit"
	labelFirstFit = "first-fit"
)

type ItemSorter interface {
	Name() string
	Compare(a, b Instance) int
}

type itemOrder struct {
	name    string
	suffix  string
	compare func(a, b *Item) int
}

func (o *itemOrder) Name() string {
	return o.name
}

func (o *itemOrder) Compare(a, b Instance) int {
	return o.compare(a.Item, b.Item)
}

//nolint:gochecknoglobals // each is one fixed rule, exposed as a value.
var (
	OrderDecreasing ItemSorter = &itemOrder{name: "decreasing", suffix: "Decreasing", compare: byVolumeDesc}

	OrderIncreasing ItemSorter = &itemOrder{name: "increasing", suffix: "Increasing", compare: byVolumeAsc}

	OrderAsGiven ItemSorter = &itemOrder{name: "as-given", suffix: "", compare: func(_, _ *Item) int { return 0 }}

	OrderHeaviestFirst ItemSorter = &itemOrder{name: "heaviest-first", suffix: "HeaviestFirst", compare: byWeightDesc}

	OrderLongestFirst ItemSorter = &itemOrder{name: "longest-first", suffix: "LongestFirst", compare: byLengthDesc}
)

func byVolumeAsc(a, b *Item) int {
	return cmp.Compare(a.volume, b.volume)
}

func byVolumeDesc(a, b *Item) int {
	return byVolumeAsc(b, a)
}

func byWeightDesc(a, b *Item) int {
	return cmp.Compare(b.weight, a.weight)
}

func byLengthDesc(a, b *Item) int {
	return cmp.Compare(b.maxLength, a.maxLength)
}

func ItemOrders() []ItemSorter {
	return []ItemSorter{OrderDecreasing, OrderIncreasing, OrderAsGiven, OrderHeaviestFirst, OrderLongestFirst}
}

func ItemOrderByName(name string) (ItemSorter, bool) { //nolint:ireturn // the caller asked for one by name.
	for _, order := range ItemOrders() {
		if order.Name() == name {
			return order, true
		}
	}

	return nil, false
}

func orderSuffix(order ItemSorter) string {
	if builtin, ok := order.(*itemOrder); ok {
		return builtin.suffix
	}

	parts := strings.Split(order.Name(), "-")
	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}

	return strings.Join(parts, "")
}

func sortItems(items []*piece, order ItemSorter) []*piece {
	if order == nil || order == OrderAsGiven {
		return items
	}

	slices.SortStableFunc(items, func(a, b *piece) int {
		return order.Compare(Instance{Item: a.Item, Index: a.index}, Instance{Item: b.Item, Index: b.index})
	})

	return items
}

func byBoxVolumeAsc(a, b *Container) int {
	return cmp.Compare(a.volume, b.volume)
}

type BoxSelection int

const (
	SelectFirstFit BoxSelection = iota
	SelectBestFit
	SelectWorstFit
	SelectAlmostWorstFit
	SelectNextFit

	SelectFullestBox
)

func (s BoxSelection) String() string {
	switch s {
	case SelectFirstFit:
		return labelFirstFit
	case SelectBestFit:
		return labelBestFit
	case SelectWorstFit:
		return "worst-fit"
	case SelectAlmostWorstFit:
		return "almost-worst-fit"
	case SelectNextFit:
		return "next-fit"
	case SelectFullestBox:
		return "fullest-box"
	}

	return labelFirstFit
}

func (s BoxSelection) title() string {
	switch s {
	case SelectFirstFit:
		return "FirstFit"
	case SelectBestFit:
		return "BestFit"
	case SelectWorstFit:
		return "WorstFit"
	case SelectAlmostWorstFit:
		return "AlmostWorstFit"
	case SelectNextFit:
		return "NextFit"
	case SelectFullestBox:
		return "FullestBox"
	}

	return "FirstFit"
}

func BoxSelections() []BoxSelection {
	return []BoxSelection{
		SelectFirstFit, SelectBestFit, SelectWorstFit, SelectAlmostWorstFit, SelectNextFit,
		SelectFullestBox,
	}
}

type Greedy struct {
	order     ItemSorter
	selection BoxSelection
}

func NewGreedy(order ItemSorter, selection BoxSelection) *Greedy {
	if order == nil {
		order = OrderAsGiven
	}

	return &Greedy{
		order:     order,
		selection: selection,
	}
}

func (s *Greedy) Order() ItemSorter { //nolint:ireturn // it hands back what it was given.
	return s.order
}

func (s *Greedy) Selection() BoxSelection {
	return s.selection
}

func (s *Greedy) Name() string {
	return s.selection.title() + orderSuffix(s.order)
}

func (s *Greedy) Pack(ctx context.Context, problem *Problem) (*Packing, error) {
	items := sortItems(compact(problem.pieces), s.order)
	boxes := problem.boxes

	switch s.selection {
	case SelectFirstFit:
		return runFirstFit(ctx, boxes, items, problem)
	case SelectFullestBox:
		return runFullestBox(ctx, boxes, items, problem)
	case SelectBestFit, SelectWorstFit, SelectAlmostWorstFit, SelectNextFit:
	}

	return runItemMajor(ctx, boxes, items, problem, s.selection)
}

type RuleSettings struct {
	Order     ItemSorter
	Selection BoxSelection
}

func (r RuleSettings) Name() string {
	return NewGreedy(r.Order, r.Selection).Name()
}

func EveryRuleSettings() []RuleSettings {
	return []RuleSettings{
		{Order: OrderDecreasing, Selection: SelectFullestBox},
		{Order: OrderDecreasing, Selection: SelectFirstFit},
		{Order: OrderIncreasing, Selection: SelectFirstFit},
		{Order: OrderDecreasing, Selection: SelectBestFit},
		{Order: OrderIncreasing, Selection: SelectBestFit},
		{Order: OrderIncreasing, Selection: SelectNextFit},
		{Order: OrderIncreasing, Selection: SelectWorstFit},
		{Order: OrderIncreasing, Selection: SelectAlmostWorstFit},
	}
}
