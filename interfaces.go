package boxpacker3

import "context"

type PackingAlgorithm interface {
	Name() string
	Pack(ctx context.Context, boxes []*Box, items []*Item) (*Result, error)
}

type ComparatorFunc func(candidate, currentBest *Result) bool
