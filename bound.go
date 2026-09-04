package boxpacker3

import (
	"math"
	"slices"
)

type Bound struct {
	Boxes int

	Volume float64
}

const l2Limit = 2000

func boundOf(boxes []*Box, pieces []*piece) Bound {
	bound := Bound{Boxes: 0, Volume: 0}

	shippable := make([]*piece, 0, len(pieces))

	for _, item := range pieces {
		if fitsNoBox(item, boxes) {
			continue
		}

		shippable = append(shippable, item)
		bound.Volume += item.volume
	}

	roomiest := roomiestBox(boxes)
	if roomiest == nil || len(shippable) == 0 {
		return bound
	}

	bound.Boxes = max(continuousBound(roomiest, shippable), oneDimensionalBound(roomiest, shippable))

	if len(shippable) <= l2Limit {
		bound.Boxes = max(bound.Boxes, twoDimensionalBound(roomiest, shippable))
	}

	return bound
}

func roomiestBox(boxes []*Box) *Box {
	var roomiest *Box

	for _, box := range boxes {
		if box != nil && (roomiest == nil || box.volume > roomiest.volume) {
			roomiest = box
		}
	}

	return roomiest
}

func continuousBound(box *Box, pieces []*piece) int {
	if box.volume <= 0 {
		return 0
	}

	volume := 0.0
	for _, item := range pieces {
		volume += item.volume
	}

	return int(math.Ceil(volume/box.volume - epsilon))
}

func smallestFit(item *piece, axis Axis) float64 {
	smallest := math.Inf(1)

	for _, orientation := range item.rotations {
		smallest = min(smallest, rotatedDimension(item.Item, orientation)[axis])
	}

	if math.IsInf(smallest, 1) {
		return 0
	}

	return smallest
}

func oneDimensionalBound(box *Box, pieces []*piece) int {
	bound := 0

	for _, axes := range [3][3]Axis{
		{WidthAxis, HeightAxis, DepthAxis},
		{WidthAxis, DepthAxis, HeightAxis},
		{HeightAxis, DepthAxis, WidthAxis},
	} {
		bound = max(bound, stackedAlong(box, pieces, axes[0], axes[1], axes[2]))
	}

	return bound
}

func stackedAlong(box *Box, pieces []*piece, first, second, third Axis) int {
	lengths := make([]float64, 0, len(pieces))

	for _, item := range pieces {
		if smallestFit(item, first) > box.boxDimension(first)/2 &&
			smallestFit(item, second) > box.boxDimension(second)/2 {
			lengths = append(lengths, smallestFit(item, third))
		}
	}

	return martelloToth(lengths, box.boxDimension(third))
}

func martelloToth(lengths []float64, capacity float64) int {
	if len(lengths) == 0 || capacity <= 0 {
		return 0
	}

	sorted := slices.Clone(lengths)
	slices.Sort(sorted)

	bound := 0

	for _, cut := range cutPoints(sorted, capacity) {
		bound = max(bound, martelloTothAt(sorted, capacity, cut))
	}

	return bound
}

func cutPoints(sorted []float64, capacity float64) []float64 {
	points := make([]float64, 0, len(sorted)+1)
	points = append(points, 0)

	for _, length := range sorted {
		if length > 0 && length <= capacity/2 {
			points = append(points, length)
		}
	}

	return points
}

func martelloTothAt(sorted []float64, capacity, cut float64) int {
	var (
		big     int
		large   int
		largeUp float64
		small   float64
	)

	for _, length := range sorted {
		switch {
		case length > capacity-cut+epsilon:
			big++
		case length > capacity/2+epsilon:
			large++
			largeUp += capacity - length
		case length >= cut-epsilon:
			small += length
		}
	}

	spare := 0.0
	if small > largeUp {
		spare = math.Ceil((small - largeUp) / capacity)
	}

	return big + large + int(spare)
}

func twoDimensionalBound(box *Box, pieces []*piece) int {
	bound := 0

	for _, axes := range [3][3]Axis{
		{WidthAxis, HeightAxis, DepthAxis},
		{WidthAxis, DepthAxis, HeightAxis},
		{HeightAxis, DepthAxis, WidthAxis},
	} {
		bound = max(bound, volumeOverStack(box, pieces, axes[0], axes[1], axes[2]))
	}

	return bound
}

func volumeOverStack(box *Box, pieces []*piece, first, second, third Axis) int {
	stacked := stackedAlong(box, pieces, first, second, third)
	if stacked == 0 || box.volume <= 0 {
		return 0
	}

	var inStack, rest float64

	for _, item := range pieces {
		if smallestFit(item, first) > box.boxDimension(first)/2 &&
			smallestFit(item, second) > box.boxDimension(second)/2 {
			inStack += item.volume

			continue
		}

		rest += item.volume
	}

	spare := float64(stacked)*box.volume - inStack
	if rest <= spare {
		return stacked
	}

	return stacked + int(math.Ceil((rest-spare)/box.volume-epsilon))
}
