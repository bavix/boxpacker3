package boxpacker3

import (
	"context"
)

type SingleContainerSearch struct {
	inner Algorithm
}

func SingleContainer(inner Algorithm) *SingleContainerSearch {
	if inner == nil {
		inner = NewGreedy(OrderDecreasing, SelectFullestBox)
	}

	return &SingleContainerSearch{inner: inner}
}

func (s *SingleContainerSearch) Name() string {
	return "SingleContainer(" + s.inner.Name() + ")"
}

func (s *SingleContainerSearch) Pack(ctx context.Context, problem *Problem) (*Packing, error) {
	pieces := sortItems(compact(problem.pieces), OrderDecreasing)

	if len(problem.boxes) == 0 {
		return newPacking([]*Container{}, pieces).on(problem.boxes), nil
	}

	single := clonePtr(problem.boxes[0])
	single.quantity = 1

	best, err := s.densestSubset(ctx, problem, single, pieces)
	if err != nil {
		return nil, err
	}

	if best == nil {
		best = newPacking([]*Container{}, pieces)
	}

	return withLeftovers(best, pieces).on(problem.boxes), nil
}

const (
	singleBoxMaxRounds   = 16
	singleBoxStaleRounds = 3
)

func (s *SingleContainerSearch) densestSubset(
	ctx context.Context, problem *Problem, box *Box, candidates []*piece,
) (*Packing, error) {
	var best *Packing

	stale := 0

	for round := 0; len(candidates) > 0 && round < singleBoxMaxRounds; round++ {
		err := checkContext(ctx)
		if err != nil {
			return nil, err
		}

		if settled(best, box, candidates) {
			return best, nil
		}

		packing, err := s.fill(ctx, problem, box, candidates)
		if err != nil {
			return nil, err
		}

		best, stale = denser(best, packing, stale)

		if stale >= singleBoxStaleRounds || len(packing.unfit) == 0 {
			return best, nil
		}

		candidates = withoutTheHead(candidates)
	}

	return best, nil
}

func (s *SingleContainerSearch) fill(
	ctx context.Context, problem *Problem, box *Box, candidates []*piece,
) (*Packing, error) {
	attempt := problem.sub([]*Box{box}, cloneSlice(candidates))

	packing, err := s.inner.Pack(ctx, attempt)
	if err != nil {
		return nil, err
	}

	return attempt.finish(ctx, packing)
}

func settled(best *Packing, box *Box, candidates []*piece) bool {
	if best == nil {
		return false
	}

	return remainingVolume(candidates) <= filledVolume(best) ||
		filledVolume(best) >= box.volume-epsilon*box.volume
}

func withoutTheHead(candidates []*piece) []*piece {
	head := candidates[0]
	rest := make([]*piece, 0, len(candidates)-1)

	for _, item := range candidates[1:] {
		if head.group != "" && item.group == head.group {
			continue
		}

		rest = append(rest, item)
	}

	return rest
}

func denser(best, attempt *Packing, stale int) (*Packing, int) {
	if best == nil || filledVolume(attempt) > filledVolume(best) {
		return attempt, 0
	}

	return best, stale + 1
}

func filledVolume(result *Packing) float64 {
	var volume float64

	for _, box := range result.boxes {
		volume += box.itemsVolume
	}

	return volume
}

func remainingVolume(items []*piece) float64 {
	var volume float64

	for _, item := range items {
		volume += item.volume
	}

	return volume
}

func withLeftovers(result *Packing, offered []*piece) *Packing {
	packed := map[Instance]bool{}

	for _, box := range result.boxes {
		for _, item := range box.items {
			packed[Instance{Item: item.Item, Index: item.index}] = true
		}
	}

	left := make([]*piece, 0, len(offered))

	for _, item := range offered {
		if !packed[Instance{Item: item.Item, Index: item.index}] {
			left = append(left, item)
		}
	}

	result.unfit = left

	return result
}
