package boxpacker3

import (
	"context"
	"sync"
)

// ParallelStrategy is a meta-strategy that runs multiple packing algorithms concurrently
// and selects the best result based on a configured goal (ComparatorFunc).
type ParallelStrategy struct {
	algorithms []PackingAlgorithm
	goal       ComparatorFunc
}

// ParallelOption defines functional options for configuring the ParallelStrategy.
type ParallelOption func(*ParallelStrategy)

// NewParallelStrategy creates a new parallel runner.
// If no algorithms are provided via options, it defaults to running no algorithms
// (effectively returning an empty result) unless configured.
//
// Usage:
//
//	strategy := NewParallelStrategy(
//	    WithAlgorithms(NewMinimizeBoxesStrategy(), NewBestFitStrategy()),
//	    WithGoal(TightestPackingGoal),
//	)
func NewParallelStrategy(opts ...ParallelOption) *ParallelStrategy {
	ps := &ParallelStrategy{
		algorithms: []PackingAlgorithm{},
		goal:       MinimizeBoxesGoal,
	}

	for _, opt := range opts {
		opt(ps)
	}

	return ps
}

// WithAlgorithms appends algorithms to the execution list.
func WithAlgorithms(algos ...PackingAlgorithm) ParallelOption {
	return func(p *ParallelStrategy) {
		p.algorithms = append(p.algorithms, algos...)
	}
}

// WithGoal sets the comparator function used to determine the "best" result.
// See goals.go for standard implementations.
func WithGoal(goal ComparatorFunc) ParallelOption {
	return func(p *ParallelStrategy) {
		p.goal = goal
	}
}

// Name returns the identifier for this strategy.
func (s *ParallelStrategy) Name() string {
	return "ParallelStrategy"
}

// Pack executes all configured algorithms in parallel.
//
// It performs the following steps:
// 1. Deep copies the input boxes and items for each algorithm (to ensure thread safety).
// 2. Launches a goroutine for each algorithm.
// 3. Collects valid results.
// 4. Uses the configured ComparatorFunc (goal) to select the winner.
func (s *ParallelStrategy) Pack(ctx context.Context, boxes []*Box, items []*Item) (*Result, error) {
	// If no algorithms are configured, return all items as unfit immediately.
	if len(s.algorithms) == 0 {
		return &Result{UnfitItems: items, Boxes: []*Box{}}, nil
	}

	results := make(chan *Result, len(s.algorithms))

	var wg sync.WaitGroup

	// Launch each algorithm in a separate goroutine
	for _, algo := range s.algorithms {
		wg.Add(1)

		go func(a PackingAlgorithm) {
			defer wg.Done()

			// Check context before doing work
			if ctx.Err() != nil {
				return
			}

			res, err := a.Pack(ctx, CopySlicePtr(boxes), CopySlicePtr(items))
			if err == nil && res != nil {
				results <- res
			}
		}(algo)
	}

	// Closer goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// Select the best result
	var bestResult *Result

	// Read from channel until closed
	for res := range results {
		// Check if the new result is better than what we have so far
		if s.goal(res, bestResult) {
			bestResult = res
		}
	}

	// If context was canceled or all strategies failed, handle gracefully
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// If no strategy produced a valid result (rare, but possible if all fail),
	// return a result with all items marked as unfit.
	if bestResult == nil {
		return &Result{
			UnfitItems: items, // Return original items
			Boxes:      []*Box{},
		}, nil
	}

	return bestResult, nil
}
