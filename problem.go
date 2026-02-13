package boxpacker3

import (
	"context"
	"slices"
	"sync/atomic"
	"time"
)

type Instance struct {
	Item  *Item
	Index int
}

type Algorithm interface {
	Name() string
	Pack(ctx context.Context, problem *Problem) (*Packing, error)
}

type Problem struct {
	boxes     []*Box
	pieces    []*piece
	rules     Rules
	finishers []Finisher
	budget    Budget
	observer  Observer
	spent     atomic.Int64

	lookup map[Instance]*piece
}

func NewProblem(boxes []*Box, items []*Item) *Problem {
	return newProblem(compact(boxes), piecesOf(items), Rules{}, nil, Budget{}, nil)
}

func newProblem(
	boxes []*Box, pieces []*piece, rules Rules, finishers []Finisher,
	budget Budget, observer Observer,
) *Problem {
	problem := &Problem{
		boxes:     boxes,
		pieces:    pieces,
		rules:     rules,
		finishers: finishers,
		budget:    budget,
		observer:  observer,
		spent:     atomic.Int64{},
		lookup:    make(map[Instance]*piece, len(pieces)),
	}

	for _, item := range pieces {
		problem.lookup[Instance{Item: item.Item, Index: item.index}] = item
	}

	return problem
}

// spend records the states a search expanded, so the report can say what the
// budget bought.
func (p *Problem) spend(nodes int) {
	p.spent.Add(int64(nodes))
}

func (p *Problem) clone() *Problem {
	return newProblem(p.boxes, cloneSlice(p.pieces), p.rules, p.finishers, p.budget, p.observer)
}

func (p *Problem) sub(boxes []*Box, pieces []*piece) *Problem {
	return newProblem(boxes, pieces, p.rules, p.finishers, p.budget, p.observer)
}

func (p *Problem) Boxes() []*Box {
	return slices.Clone(p.boxes)
}

func (p *Problem) Instances() []Instance {
	instances := make([]Instance, 0, len(p.pieces))

	for _, item := range p.pieces {
		instances = append(instances, Instance{Item: item.Item, Index: item.index})
	}

	return instances
}

func (p *Problem) Rules() Rules {
	return p.rules
}

func (p *Problem) Finishers() []Finisher {
	return compactFinishers(p.finishers)
}

func (p *Problem) Budget() Budget {
	return p.budget
}

func (p *Problem) Deadline() (time.Time, bool) {
	return p.budget.deadline(time.Now())
}

func (p *Problem) Bound() Bound {
	return boundOf(p.boxes, p.pieces)
}

func (p *Problem) observe(report CandidateReport) {
	if p.observer != nil {
		p.observer.Candidate(report)
	}
}

func (p *Problem) improved(result *Result) {
	if p.observer != nil {
		p.observer.Improved(result)
	}
}

func (p *Problem) Open(box *Box, index int) *Container {
	return newContainer(box, index, p.rules, p)
}

func (p *Problem) open(boxes []*Box) []*Container {
	opened := make([]*Container, 0, len(boxes))

	for _, box := range boxes {
		opened = append(opened, p.Open(box, 0))
	}

	return opened
}

func (p *Problem) Packing(boxes []*Container, leftover []Instance) *Packing {
	unfit := make([]*piece, 0, len(leftover))

	for _, instance := range leftover {
		if item, known := p.lookup[instance]; known {
			unfit = append(unfit, item)
		}
	}

	return newPacking(compact(boxes), unfit).on(p.boxes).by(p)
}

func (p *Problem) pieceFor(instance Instance) *piece {
	if p == nil {
		return nil
	}

	return p.lookup[instance]
}

func (c *Container) TryPlace(instance Instance, position Pivot, orientation Orientation) bool {
	item := c.problem.pieceFor(instance)
	if item == nil || !c.canQuota(item) {
		return false
	}

	allowUnstable := !c.hasStableOrientation(item)
	if !c.acceptsPlacement(item, position, rotatedDimension(item.Item, orientation), allowUnstable) {
		return false
	}

	c.place(item, position, orientation)

	return true
}

func (c *Container) Fit(instance Instance) bool {
	return fitInSpecificBox(c, c.problem.pieceFor(instance))
}

func (c *Container) Stats() BoxStats {
	return c.stats()
}
