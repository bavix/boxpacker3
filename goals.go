package boxpacker3

import (
	"math"
)

// MetricFunc calculates a specific numeric value from a packing Result.
type MetricFunc func(res *Result) float64

// Direction defines optimization direction.
type Direction int

const (
	LessIsBetter Direction = iota
	MoreIsBetter
)

// Criterion represents a single step in the decision-making process.
type Criterion struct {
	Metric    MetricFunc
	Direction Direction
}

// MakeGoal creates a ComparatorFunc from a list of criteria.
// It compares results sequentially.
func MakeGoal(criteria ...Criterion) ComparatorFunc {
	return func(cand, best *Result) bool {
		// If there is no current best, the candidate automatically wins
		if best == nil {
			return true
		}

		const epsilon = 0.00001

		for _, c := range criteria {
			valCand := c.Metric(cand)
			valBest := c.Metric(best)

			// If values are essentially equal, continue to the next tie-breaker
			if math.Abs(valCand-valBest) < epsilon {
				continue
			}

			if c.Direction == LessIsBetter {
				return valCand < valBest
			}

			return valCand > valBest
		}

		// If all criteria are equal, the candidate is not strictly "better"
		return false
	}
}

//nolint:gochecknoglobals // Metrics are stateless functions defined as vars for composition.
var (
	// UnfitCountMetric: Number of unpacked items.
	UnfitCountMetric MetricFunc = func(res *Result) float64 {
		return float64(len(res.UnfitItems))
	}

	// BoxCountMetric: Number of non-empty boxes.
	BoxCountMetric MetricFunc = func(res *Result) float64 {
		return float64(countUsedBoxes(res.Boxes))
	}

	// TotalVolumeMetric: Sum of the capacity of all used boxes (shipping air + items).
	TotalVolumeMetric MetricFunc = func(res *Result) float64 {
		return getUsedVolume(res.Boxes)
	}

	// AverageFillRateMetric: Average percentage of box volume used (0.0 - 1.0).
	AverageFillRateMetric MetricFunc = func(res *Result) float64 {
		return getAverageFillRate(res.Boxes)
	}

	// WeightStdDevMetric: Standard deviation of box weights (lower is more balanced).
	WeightStdDevMetric MetricFunc = func(res *Result) float64 {
		return getWeightStdDev(res.Boxes)
	}
)

//nolint:gochecknoglobals // Strategies are stateless configurations.
var (
	// MinimizeBoxesGoal prioritizes using the fewest number of boxes possible.
	// This is the classic bin packing goal, ideal for reducing shipping label costs.
	//
	// 1. Maximize items packed (minimize unfit items).
	// 2. Minimize number of boxes used.
	// 3. Minimize total capacity of boxes used (prefer smaller boxes -> higher fill rate).
	MinimizeBoxesGoal = MakeGoal(
		Criterion{UnfitCountMetric, LessIsBetter},
		Criterion{BoxCountMetric, LessIsBetter},
		Criterion{TotalVolumeMetric, LessIsBetter},
	)

	// TightestPackingGoal prioritizes high density / volume utilization.
	//
	// 1. Maximize items packed (minimize unfit items).
	// 2. Minimize total capacity of boxes used (highest density).
	// 3. Minimize number of boxes used.
	TightestPackingGoal = MakeGoal(
		Criterion{UnfitCountMetric, LessIsBetter},
		Criterion{TotalVolumeMetric, LessIsBetter},
		Criterion{BoxCountMetric, LessIsBetter},
	)

	// MaxAverageFillRateGoal prioritizes the "Unboxing Experience".
	//
	// 1. Maximize items packed (minimize unfit items).
	// 2. Maximize average fill rate (each box should be as full as possible).
	MaxAverageFillRateGoal = MakeGoal(
		Criterion{UnfitCountMetric, LessIsBetter},
		Criterion{AverageFillRateMetric, MoreIsBetter},
	)

	// BalancedPackingGoal prioritizes safe distribution and handling.
	//
	// 1. Maximize items packed (minimize unfit items).
	// 2. Minimize weight deviation (balance the load).
	// 3. Minimize number of boxes used.
	BalancedPackingGoal = MakeGoal(
		Criterion{UnfitCountMetric, LessIsBetter},
		Criterion{WeightStdDevMetric, LessIsBetter},
		Criterion{BoxCountMetric, LessIsBetter},
	)
)

// countUsedBoxes helps calculate how many boxes actually contain items.
func countUsedBoxes(boxes []*Box) int {
	n := 0

	for _, b := range boxes {
		if len(b.items) > 0 {
			n++
		}
	}

	return n
}

// getUsedVolume calculates the total capacity of all boxes that contain items.
func getUsedVolume(boxes []*Box) float64 {
	var v float64

	for _, b := range boxes {
		// Only count boxes that actually have items in them
		if len(b.items) > 0 {
			v += b.volume // Use the box's volume (capacity), not itemsVolume
		}
	}

	return v
}

// getAverageFillRate calculates the average fill percentage of used boxes.
func getAverageFillRate(boxes []*Box) float64 {
	var (
		totalRate float64
		count     int
	)

	for _, b := range boxes {
		if len(b.items) > 0 && b.volume > 0 {
			totalRate += b.itemsVolume / b.volume
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return totalRate / float64(count)
}

// getWeightStdDev calculates the standard deviation of box weights.
func getWeightStdDev(boxes []*Box) float64 {
	var (
		weights []float64
		sum     float64
	)

	for _, b := range boxes {
		if len(b.items) > 0 {
			w := b.itemsWeight
			weights = append(weights, w)
			sum += w
		}
	}

	count := float64(len(weights))
	if count <= 1 {
		return 0
	}

	mean := sum / count

	var variance float64

	for _, w := range weights {
		diff := w - mean
		variance += diff * diff
	}

	return math.Sqrt(variance / count)
}
