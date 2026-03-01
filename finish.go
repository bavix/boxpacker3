package boxpacker3

import (
	"cmp"
	"context"
	"slices"
)

type Finisher interface {
	Name() string
	Finish(ctx context.Context, problem *Problem, packing *Packing) error
}

func DefaultFinishers() []Finisher {
	return []Finisher{BalanceWeight{MaxBoxes: defaultBalanceBoxes}}
}

const defaultBalanceBoxes = 12

type BalanceWeight struct {
	MaxBoxes int
}

func (BalanceWeight) Name() string {
	return "BalanceWeight"
}

func (f BalanceWeight) Finish(_ context.Context, _ *Problem, packing *Packing) error {
	redistributeWeight(packing, f.MaxBoxes)

	return nil
}

//nolint:gochecknoglobals // one fixed pass, exposed as a value.
var Consolidate Finisher = finisherFunc{name: "Consolidate", run: consolidate}

func consolidate(ctx context.Context, problem *Problem, packing *Packing) error {
	bound := problem.Bound().Boxes

	for range len(packing.boxes) {
		if bound > 0 && countUsedBoxes(packing.boxes) <= bound {
			return nil
		}

		err := checkContext(ctx)
		if err != nil {
			return err
		}

		if !freedOne(packing) {
			return nil
		}
	}

	return nil
}

//nolint:gochecknoglobals // one fixed pass, exposed as a value.
var Rescue Finisher = finisherFunc{name: "Rescue", run: rescue}

func rescue(ctx context.Context, _ *Problem, packing *Packing) error {
	for range len(packing.unfit) {
		err := checkContext(ctx)
		if err != nil {
			return err
		}

		if !placedOneLeftBehind(packing) {
			return nil
		}
	}

	return nil
}

//nolint:gochecknoglobals // one fixed pass, exposed as a value.
var Rehome Finisher = finisherFunc{name: "Rehome", run: rehome}

func rehome(ctx context.Context, problem *Problem, packing *Packing) error {
	for at, box := range packing.boxes {
		err := checkContext(ctx)
		if err != nil {
			return err
		}

		if box == nil || len(box.items) == 0 {
			continue
		}

		smaller := smallestHome(ctx, problem, packing, box)
		if smaller == nil {
			continue
		}

		packing.boxes[at] = smaller
	}

	return nil
}

func smallestHome(ctx context.Context, problem *Problem, packing *Packing, box *Container) *Container {
	kinds := slices.Clone(problem.boxes)
	slices.SortStableFunc(kinds, func(a, b *Box) int {
		return cmp.Compare(a.volume, b.volume)
	})

	moved := slices.Clone(box.items)
	saved := snapshot(moved)

	for _, kind := range kinds {
		if kind == nil || kind.volume >= box.volume {
			continue
		}

		index, free := packing.spareIndex(kind, box)
		if !free {
			continue
		}

		fresh := problem.Open(kind, index)

		left, err := packToBox(ctx, fresh, moved)
		if err == nil && len(left) == 0 {
			return fresh
		}

		restore(moved, saved)
	}

	restore(moved, saved)

	return nil
}

func (p *Packing) spareIndex(kind *Box, leaving *Container) (int, bool) {
	taken := make(map[int]bool, kind.quantity)

	for _, box := range p.boxes {
		if box == nil || box == leaving || box.Box != kind || len(box.items) == 0 {
			continue
		}

		taken[box.index] = true
	}

	for index := range max(kind.quantity, 1) {
		if !taken[index] {
			return index, true
		}
	}

	return 0, false
}

type finisherFunc struct {
	name string
	run  func(ctx context.Context, problem *Problem, packing *Packing) error
}

func (f finisherFunc) Name() string {
	return f.name
}

func (f finisherFunc) Finish(ctx context.Context, problem *Problem, packing *Packing) error {
	return f.run(ctx, problem, packing)
}

func (p *Problem) finish(ctx context.Context, packing *Packing) (*Packing, error) {
	if packing == nil {
		return nil, nil //nolint:nilnil
	}

	enforceGroups(packing)

	for _, finisher := range p.finishers {
		err := finisher.Finish(ctx, p, packing)
		if err != nil {
			return nil, err
		}
	}

	return packing, nil
}
