package boxpacker3

import (
	"context"
	"time"
)

type Option func(*Packer)

func WithAlgorithm(algorithm Algorithm) Option {
	return func(p *Packer) {
		if algorithm != nil {
			p.algorithm = algorithm
		}
	}
}

func WithRules(rules Rules) Option {
	return func(p *Packer) {
		p.rules = p.rules.merged(rules.clamped())
	}
}

func WithFinishers(finishers ...Finisher) Option {
	return func(p *Packer) {
		p.finishers = compactFinishers(finishers)
	}
}

func WithMerit(merit Merit) Option {
	return func(p *Packer) {
		if merit != nil {
			p.rules.Merit = merit
		}
	}
}

func WithFiller(filler Filler) Option {
	return func(p *Packer) {
		if filler != nil {
			p.rules.Filler = filler
		}
	}
}

func WithBudget(budget Budget) Option {
	return func(p *Packer) {
		p.budget = budget
	}
}

func WithObserver(observer Observer) Option {
	return func(p *Packer) {
		p.observer = observer
	}
}

func compactFinishers(finishers []Finisher) []Finisher {
	kept := make([]Finisher, 0, len(finishers))

	for _, finisher := range finishers {
		if finisher != nil {
			kept = append(kept, finisher)
		}
	}

	return kept
}

type Packer struct {
	algorithm Algorithm
	rules     Rules
	finishers []Finisher
	budget    Budget
	observer  Observer
}

func NewPacker(opts ...Option) *Packer {
	p := &Packer{
		algorithm: NewGreedy(OrderDecreasing, SelectFullestBox),
		rules:     Rules{},
		finishers: DefaultFinishers(),
		budget:    Budget{},
		observer:  nil,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *Packer) Algorithm() Algorithm { //nolint:ireturn // it hands back what it was given.
	return p.algorithm
}

func (p *Packer) Rules() Rules {
	return p.rules
}

func (p *Packer) Finishers() []Finisher {
	return compactFinishers(p.finishers)
}

func (p *Packer) Budget() Budget {
	return p.budget
}

func (p *Packer) problem(boxes []*Box, items []*Item) *Problem {
	return newProblem(compact(boxes), piecesOf(items), p.rules, p.finishers, p.budget, p.observer)
}

func (p *Packer) Pack(ctx context.Context, boxes []*Box, items []*Item) (*Result, error) {
	started := time.Now()
	problem := p.problem(boxes, items)

	budgeted, done := p.budget.apply(ctx)
	defer done()

	packing, err := p.algorithm.Pack(budgeted, problem)
	if err != nil {
		return nil, err
	}

	packing, err = problem.finish(budgeted, packing)
	if err != nil {
		return nil, err
	}

	result := packing.result()
	result.Report = problem.report(p.algorithm.Name(), started, result, packing)

	return result, nil
}
