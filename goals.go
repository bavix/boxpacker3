package boxpacker3

import (
	"math"
)

type metricFunc func(res *Result) float64

type direction int

const (
	lessIsBetter direction = iota
	moreIsBetter

	epsilon = 0.00001
)

type criterion struct {
	metric    metricFunc
	direction direction
}

func makeGoal(criteria ...criterion) ComparatorFunc {
	return func(candidate, currentBest *Result) bool {
		if currentBest == nil {
			return true
		}

		for _, c := range criteria {
			valCand := c.metric(candidate)
			valBest := c.metric(currentBest)

			if math.Abs(valCand-valBest) < epsilon {
				continue
			}

			if c.direction == lessIsBetter {
				return valCand < valBest
			}

			return valCand > valBest
		}

		return false
	}
}

func unfitCountMetric(res *Result) float64 {
	return float64(len(res.UnfitItems))
}

func boxCountMetric(res *Result) float64 {
	return float64(countUsedBoxes(res.Boxes))
}

func totalVolumeMetric(res *Result) float64 {
	return getUsedVolume(res.Boxes)
}

func averageFillRateMetric(res *Result) float64 {
	return getAverageFillRate(res.Boxes)
}

func weightStdDevMetric(res *Result) float64 {
	return getWeightStdDev(res.Boxes)
}

// MinimizeBoxesGoal prioritizes using the fewest number of boxes possible.
// This is the classic bin packing goal, ideal for reducing shipping label costs.
//
// 1. Maximize items packed (minimize unfit items).
// 2. Minimize number of boxes used.
// 3. Minimize total volume of boxes used (prefer smaller boxes).
func MinimizeBoxesGoal(candidate, currentBest *Result) bool {
	return makeGoal(
		criterion{unfitCountMetric, lessIsBetter},
		criterion{boxCountMetric, lessIsBetter},
		criterion{totalVolumeMetric, lessIsBetter},
	)(candidate, currentBest)
}

// MaximizeItemsGoal prioritizes fitting the maximum number of items, regardless of box efficiency.
// Ideal for fixed-container scenarios (e.g., loading a truck) where leaving items behind is the worst outcome.
//
// 1. Maximize items packed (minimize unfit items).
func MaximizeItemsGoal(candidate, currentBest *Result) bool {
	return makeGoal(
		criterion{unfitCountMetric, lessIsBetter},
	)(candidate, currentBest)
}

// TightestPackingGoal prioritizes high density / volume utilization.
// Ideal when shipping costs are calculated based on dimensional weight or total volume.
//
// 1. Maximize items packed (minimize unfit items).
// 2. Minimize total volume of boxes used.
// 3. Minimize number of boxes used.
func TightestPackingGoal(candidate, currentBest *Result) bool {
	return makeGoal(
		criterion{unfitCountMetric, lessIsBetter},
		criterion{totalVolumeMetric, lessIsBetter},
		criterion{boxCountMetric, lessIsBetter},
	)(candidate, currentBest)
}

// MaxAverageFillRateGoal prioritizes maximizing the average fill rate of used boxes.
// Ideal when shipping costs are influenced by dimensional weight or when higher density reduces cost.
//
// 1. Maximize items packed (minimize unfit items).
// 2. Maximize average fill rate across used boxes.
func MaxAverageFillRateGoal(candidate, currentBest *Result) bool {
	return makeGoal(
		criterion{unfitCountMetric, lessIsBetter},
		criterion{averageFillRateMetric, moreIsBetter},
	)(candidate, currentBest)
}

// BalancedPackingGoal prioritizes a balanced distribution of weights and box sizes.
// Ideal for scenarios where the total weight of items needs to be distributed evenly across boxes.
//
// 1. Maximize items packed (minimize unfit items).
// 2. Minimize weight standard deviation.
// 3. Minimize number of boxes used.
func BalancedPackingGoal(candidate, currentBest *Result) bool {
	return makeGoal(
		criterion{unfitCountMetric, lessIsBetter},
		criterion{weightStdDevMetric, lessIsBetter},
		criterion{boxCountMetric, lessIsBetter},
	)(candidate, currentBest)
}

func countUsedBoxes(boxes []*Box) int {
	n := 0

	for _, b := range boxes {
		if len(b.items) > 0 {
			n++
		}
	}

	return n
}

func getUsedVolume(boxes []*Box) float64 {
	var v float64

	for _, b := range boxes {
		if len(b.items) > 0 {
			v += b.volume
		}
	}

	return v
}

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
