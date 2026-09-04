package boxpacker3

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"time"
)

type Portfolio struct {
	algorithms []Algorithm
	goal       Goal
	name       string
}

func NewPortfolio(goal Goal, algorithms ...Algorithm) *Portfolio {
	if goal == nil {
		goal = FewestBoxes
	}

	return &Portfolio{algorithms: compactAlgorithms(algorithms), goal: goal, name: ""}
}

func compactAlgorithms(algorithms []Algorithm) []Algorithm {
	kept := make([]Algorithm, 0, len(algorithms))

	for _, algorithm := range algorithms {
		if algorithm != nil {
			kept = append(kept, algorithm)
		}
	}

	return kept
}

func (s *Portfolio) Name() string {
	if s.name != "" {
		return s.name
	}

	names := make([]string, 0, len(s.algorithms))

	for _, algorithm := range s.algorithms {
		names = append(names, algorithm.Name())
	}

	if len(names) == 0 {
		return "BestFor(" + s.goal.Name() + ": nothing)"
	}

	return "BestFor(" + s.goal.Name() + ": " + strings.Join(names, ", ") + ")"
}

func (s *Portfolio) Goal() Goal { //nolint:ireturn // it hands back what it was given.
	return s.goal
}

func (s *Portfolio) Algorithms() []Algorithm {
	return slices.Clone(s.algorithms)
}

func (s *Portfolio) Pack(ctx context.Context, problem *Problem) (*Packing, error) {
	if len(s.algorithms) == 0 {
		return newPacking([]*Container{}, cloneSlice(problem.pieces)).on(problem.boxes), nil
	}

	packings, elapsed, errs := s.runAll(ctx, problem)

	best, reports, err := s.pick(ctx, problem, packings, elapsed, errs)
	if err != nil {
		return nil, err
	}

	if best != nil {
		best.candidates = reports
		best.truncated = truncated(reports)

		return best, nil
	}

	joined := errors.Join(errs...)
	if joined != nil {
		return nil, joined
	}

	empty := newPacking([]*Container{}, cloneSlice(problem.pieces)).on(problem.boxes)
	empty.candidates, empty.truncated = reports, truncated(reports)

	return empty, nil
}

func truncated(reports []CandidateReport) bool {
	for _, report := range reports {
		if errors.Is(report.Err, context.DeadlineExceeded) {
			return true
		}
	}

	return false
}

func (s *Portfolio) runAll(ctx context.Context, problem *Problem) ([]*Packing, []time.Duration, []error) {
	packings := make([]*Packing, len(s.algorithms))
	elapsed := make([]time.Duration, len(s.algorithms))
	errs := make([]error, len(s.algorithms))

	workers := make(chan struct{}, problem.budget.workers())

	var group sync.WaitGroup

	for i, algorithm := range s.algorithms {
		group.Go(func() {
			workers <- struct{}{}
			defer func() { <-workers }()

			started := time.Now()

			err := ctx.Err()
			if err != nil {
				errs[i] = err

				return
			}

			packings[i], errs[i] = algorithm.Pack(ctx, problem.clone())
			elapsed[i] = time.Since(started)
		})
	}

	group.Wait()

	return packings, elapsed, errs
}

func (s *Portfolio) pick(
	ctx context.Context, problem *Problem, packings []*Packing, elapsed []time.Duration, errs []error,
) (*Packing, []CandidateReport, error) {
	var (
		best      *Packing
		finished  *Result
		bestAt    = -1
		reports   = make([]CandidateReport, 0, len(packings))
		finishCtx = context.WithoutCancel(ctx)
	)

	for i, packing := range packings {
		report := CandidateReport{
			Algorithm: s.algorithms[i].Name(), Elapsed: elapsed[i],
			Boxes: 0, Unpacked: 0, Chosen: false, Err: errs[i],
		}

		if packing == nil {
			reports = append(reports, report)
			problem.observe(report)

			continue
		}

		packing, err := problem.finish(finishCtx, packing)
		if err != nil {
			return nil, nil, err
		}

		answer := packing.result()
		report.Boxes, report.Unpacked = len(answer.Boxes), len(answer.Unpacked)

		if prefers(s.goal, answer, finished) {
			best, finished, bestAt = packing, answer, i

			problem.improved(answer)
		}

		reports = append(reports, report)
		problem.observe(report)
	}

	if bestAt >= 0 {
		reports[bestAt].Chosen = true
	}

	return best, reports, nil
}

func BestFor(goal Goal) *Portfolio {
	if goal == nil {
		goal = FewestBoxes
	}

	runner := NewPortfolio(goal, repairedRules(goal)...)
	runner.name = goal.Name()

	return runner
}

func BestForName(name string) (*Portfolio, bool) {
	goal, known := GoalByName(name)
	if !known {
		return nil, false
	}

	return BestFor(goal), true
}

func EveryRule(goal Goal) *Portfolio {
	return NewPortfolio(goal, everyRuleAlgorithms()...)
}

func everyRuleAlgorithms() []Algorithm {
	settings := EveryRuleSettings()
	algorithms := make([]Algorithm, 0, len(settings))

	for _, setting := range settings {
		algorithms = append(algorithms, NewGreedy(setting.Order, setting.Selection))
	}

	return algorithms
}

func repairedRules(goal Goal) []Algorithm {
	repairs := []Finisher{Rescue}
	if goal.Name() != nameBalancedWeight {
		repairs = []Finisher{Consolidate, Rescue}
	}

	algorithms := everyRuleAlgorithms()
	both := make([]Algorithm, 0, len(algorithms)*2) //nolint:mnd // as it stands and repaired.

	for _, algorithm := range algorithms {
		both = append(both, algorithm, Repaired(algorithm, repairs...))
	}

	return both
}

func Repaired(inner Algorithm, finishers ...Finisher) Algorithm { //nolint:ireturn // it wraps what it was given.
	return &repairedAlgorithm{inner: inner, finishers: compactFinishers(finishers)}
}

type repairedAlgorithm struct {
	inner     Algorithm
	finishers []Finisher
}

func (a *repairedAlgorithm) Name() string {
	return a.inner.Name() + "+repaired"
}

func (a *repairedAlgorithm) Pack(ctx context.Context, problem *Problem) (*Packing, error) {
	packing, err := a.inner.Pack(ctx, problem)
	if err != nil {
		return nil, err
	}

	for _, finisher := range a.finishers {
		err = finisher.Finish(ctx, problem, packing)
		if err != nil {
			return nil, err
		}
	}

	return packing, nil
}
