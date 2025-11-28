package boxpacker3

import (
	"context"
)

// PackerOption is a functional option for configuring a Packer.
type PackerOption func(*Packer)

// WithStrategy sets the packing strategy using the legacy enum constants.
// This ensures backward compatibility with existing codebases.
func WithStrategy(strategy PackingStrategy) PackerOption {
	return func(p *Packer) {
		switch strategy {
		case StrategyMinimizeBoxes:
			p.algorithm = NewMinimizeBoxesStrategy()
		case StrategyGreedy:
			p.algorithm = NewGreedyStrategy()
		case StrategyBestFit:
			p.algorithm = NewBestFitStrategy()
		case StrategyBestFitDecreasing:
			p.algorithm = NewBestFitDecreasingStrategy()
		case StrategyNextFit:
			p.algorithm = NewNextFitStrategy()
		case StrategyWorstFit:
			p.algorithm = NewWorstFitStrategy()
		case StrategyAlmostWorstFit:
			p.algorithm = NewAlmostWorstFitStrategy()
		default:
			p.algorithm = NewMinimizeBoxesStrategy()
		}
	}
}

// WithAlgorithm sets a specific packing algorithm instance.
// This allows for custom implementations or the use of the ParallelStrategy runner.
func WithAlgorithm(algo PackingAlgorithm) PackerOption {
	return func(p *Packer) {
		p.algorithm = algo
	}
}

// Packer packs items into boxes using a configurable algorithm.
type Packer struct {
	algorithm PackingAlgorithm
}

// Result represents the result of packing items into boxes.
type Result struct {
	UnfitItems itemSlice
	Boxes      boxSlice
}

// NewPacker creates a new Packer.
// By default, it uses the MinimizeBoxes strategy (First Fit Decreasing) to match historical behavior.
func NewPacker(opts ...PackerOption) *Packer {
	p := &Packer{
		algorithm: NewMinimizeBoxesStrategy(),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// PackCtx packs items into boxes with context support for cancellation.
// It delegates the actual logic to the configured PackingAlgorithm.
func (p *Packer) PackCtx(ctx context.Context, inputBoxes []*Box, inputItems []*Item) (*Result, error) {
	if inputBoxes == nil {
		inputBoxes = []*Box{}
	}

	if inputItems == nil {
		inputItems = []*Item{}
	}

	return p.algorithm.Pack(ctx, CopySlicePtr(inputBoxes), CopySlicePtr(inputItems))
}

// Pack packs items into boxes.
//
// Deprecated: Use PackCtx instead. This function is kept for backward compatibility
// but PackCtx provides better control with context support for cancellation.
//
// Parameters:
// - inputBoxes: a list of boxes.
// - inputItems: a list of items.
//
// Returns:
// - a Result struct that contains two slices:
//   - Boxes: a list of boxes with items.
//   - UnfitItems: a list of items that didn't fit into boxes.
func (p *Packer) Pack(inputBoxes []*Box, inputItems []*Item) *Result {
	res, _ := p.PackCtx(context.Background(), inputBoxes, inputItems)

	return res
}
