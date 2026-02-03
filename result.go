package boxpacker3

import (
	"slices"
	"time"
)

type Result struct {
	Boxes    []PackedBox
	Unpacked []UnpackedItem
	Report   Report
}

type PackedBox struct {
	Box   *Box
	Index int
	Items []PackedItem
	Stats BoxStats
}

func (b PackedBox) ID() string {
	return b.Box.id
}

func (b PackedBox) RemainingVolume() float64 {
	return b.Box.volume - b.Stats.ItemsVolume
}

func (b PackedBox) RemainingWeight() float64 {
	return b.Box.maxWeight - b.Stats.GrossWeight
}

func (b PackedBox) RemainingArea() float64 {
	used := 0.0

	for _, item := range b.Items {
		used += item.Dimension[WidthAxis] * item.Dimension[HeightAxis]
	}

	return b.Box.Area() - used
}

type BoxStats struct {
	ItemsVolume float64
	ItemsWeight float64
	GrossWeight float64
	Fill        float64
	Envelope    Dimension
}

type PackedItem struct {
	Item        *Item
	Index       int
	Position    Pivot
	Dimension   Dimension
	Orientation Orientation
}

func (p PackedItem) ID() string {
	return p.Item.id
}

func (p PackedItem) Area() float64 {
	return p.Dimension[WidthAxis] * p.Dimension[HeightAxis]
}

type UnpackedItem struct {
	Item   *Item
	Index  int
	Reason Reason
}

func (u UnpackedItem) ID() string {
	return u.Item.id
}

type Reason uint8

const (
	ReasonNoRoom Reason = iota

	ReasonTooBig

	ReasonTooHeavy

	ReasonNoOrientationFits

	ReasonRefusedByBox

	ReasonGroupIncomplete

	ReasonCancelled
)

func (r Reason) String() string {
	switch r {
	case ReasonNoRoom:
		return "no room"
	case ReasonTooBig:
		return "too big"
	case ReasonTooHeavy:
		return "too heavy"
	case ReasonNoOrientationFits:
		return "no permitted orientation fits"
	case ReasonRefusedByBox:
		return "refused by every box"
	case ReasonGroupIncomplete:
		return "group incomplete"
	case ReasonCancelled:
		return "cancelled"
	}

	return "no room"
}

func (r Reason) Structural() bool {
	switch r {
	case ReasonTooBig, ReasonTooHeavy, ReasonNoOrientationFits, ReasonRefusedByBox:
		return true
	case ReasonNoRoom, ReasonGroupIncomplete, ReasonCancelled:
	}

	return false
}

func Reasons() []Reason {
	return []Reason{
		ReasonNoRoom, ReasonTooBig, ReasonTooHeavy, ReasonNoOrientationFits,
		ReasonRefusedByBox, ReasonGroupIncomplete, ReasonCancelled,
	}
}

func (r *Result) Unpackable() []UnpackedItem {
	unpackable := make([]UnpackedItem, 0)

	for _, item := range r.Unpacked {
		if item.Reason.Structural() {
			unpackable = append(unpackable, item)
		}
	}

	return unpackable
}

type Report struct {
	Algorithm string

	Elapsed time.Duration

	Candidates []CandidateReport

	Nodes int

	Bound Bound

	GapBoxes int

	Fill float64

	Truncated bool
}

type CandidateReport struct {
	Algorithm string

	Elapsed time.Duration

	Boxes int

	Unpacked int

	Chosen bool

	Err error
}

type Packing struct {
	boxes []*Container
	unfit []*piece

	shelf []*Box

	candidates []CandidateReport
	truncated  bool
	problem    *Problem
}

func (p *Packing) on(shelf []*Box) *Packing {
	p.shelf = shelf

	return p
}

func (p *Packing) offered() []*Box {
	if len(p.shelf) > 0 {
		return p.shelf
	}

	offered := make([]*Box, 0, len(p.boxes))
	for _, box := range p.boxes {
		offered = append(offered, box.Box)
	}

	return offered
}

func (p *Packing) result() *Result {
	result := &Result{
		Boxes:    make([]PackedBox, 0, len(p.boxes)),
		Unpacked: make([]UnpackedItem, 0, len(p.unfit)),
		Report:   Report{Algorithm: "", Elapsed: 0},
	}

	for _, box := range p.boxes {
		if box != nil && len(box.items) > 0 {
			result.Boxes = append(result.Boxes, box.packed())
		}
	}

	for _, item := range p.unfit {
		if item == nil {
			continue
		}

		result.Unpacked = append(result.Unpacked, UnpackedItem{
			Item: item.Item, Index: item.index, Reason: reasonFor(item, p.offered()),
		})
	}

	return result
}

func reasonFor(item *piece, boxes []*Box) Reason {
	if item.reason == ReasonGroupIncomplete || item.reason == ReasonCancelled {
		return item.reason
	}

	if !fitsNoBox(item, boxes) {
		return ReasonNoRoom
	}

	return refusal(item, boxes)
}

func refusal(item *piece, boxes []*Box) Reason {
	tooHeavy, refused, turned := true, true, false

	for _, box := range compact(boxes) {
		if withinLimit(box.emptyWeight+item.weight, box.maxWeight) {
			tooHeavy = false
		}

		if len(box.accepts) == 0 || slices.Contains(box.accepts, item.class) {
			refused = false
		}

		if box.fitsFreelyTurned(item) {
			turned = true
		}
	}

	switch {
	case tooHeavy:
		return ReasonTooHeavy
	case refused:
		return ReasonRefusedByBox
	case turned:
		return ReasonNoOrientationFits
	}

	return ReasonTooBig
}

func (b PackedBox) Volume() float64 {
	return b.Box.volume
}

func (p PackedItem) Volume() float64 {
	return p.Item.volume
}

func (p PackedItem) Weight() float64 {
	return p.Item.weight
}

func (u UnpackedItem) Volume() float64 {
	return u.Item.volume
}

func (u UnpackedItem) Weight() float64 {
	return u.Item.weight
}

func (p *Problem) report(algorithm string, started time.Time, result *Result, packing *Packing) Report {
	bound := p.Bound()

	return Report{
		Algorithm:  algorithm,
		Elapsed:    time.Since(started),
		Candidates: packing.candidates,
		Nodes:      packing.spentNodes(),
		Bound:      bound,
		GapBoxes:   max(len(result.Boxes)-bound.Boxes, 0),
		Fill:       MetricsOf(result).Fill,
		Truncated:  packing.truncated,
	}
}

// spentNodes is how many states a search expanded on the way to this packing,
// zero when none ran.
func (p *Packing) spentNodes() int {
	if p == nil || p.problem == nil {
		return 0
	}

	return int(p.problem.spent.Load())
}

func (p *Packing) by(problem *Problem) *Packing {
	p.problem = problem

	return p
}

func newPacking(boxes []*Container, unfit []*piece) *Packing {
	return &Packing{boxes: boxes, unfit: unfit, shelf: nil, candidates: nil, truncated: false, problem: nil}
}
