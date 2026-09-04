package boxpacker3

import (
	"cmp"
	"math"
	"slices"
)

const minBoxesToBalance = 2

func usedBoxes(boxes []*Container) []*Container {
	used := make([]*Container, 0, len(boxes))

	for _, box := range boxes {
		if box != nil && len(box.items) > 0 {
			used = append(used, box)
		}
	}

	return used
}

func redistributeWeight(result *Packing, maxBoxes int) *Packing {
	if result == nil || maxBoxes < minBoxesToBalance {
		return result
	}

	boxes := usedBoxes(result.boxes)
	if len(boxes) < minBoxesToBalance || len(boxes) > maxBoxes {
		return result
	}

	for moveOne(boxes) {
		boxes = usedBoxes(result.boxes)
		if len(boxes) < minBoxesToBalance {
			break
		}
	}

	return result
}

func moveOne(boxes []*Container) bool {
	slices.SortStableFunc(boxes, func(a, b *Container) int {
		return cmp.Compare(b.itemsWeight, a.itemsWeight)
	})

	for i, heavier := range boxes {
		for _, lighter := range boxes[i+1:] {
			if heavier.itemsWeight == lighter.itemsWeight {
				continue
			}

			if moveFrom(heavier, lighter) {
				return true
			}
		}
	}

	return false
}

func moveFrom(from, to *Container) bool {
	for _, item := range from.items {
		if item.group != "" {
			continue
		}

		if !improvesBalance(from, to, item) {
			continue
		}

		if relocate(from, to, item) {
			return true
		}
	}

	return false
}

func improvesBalance(from, to *Container, item *piece) bool {
	before := math.Abs(from.itemsWeight - to.itemsWeight)
	after := math.Abs((from.itemsWeight - item.weight) - (to.itemsWeight + item.weight))

	return after < before
}

func relocate(from, to *Container, item *piece) bool {
	fromState, toState := saveContainer(from), saveContainer(to)

	remaining := make([]*piece, 0, len(from.items))

	for _, other := range from.items {
		if other != item {
			remaining = append(remaining, other)
		}
	}

	receiving := append(slices.Clone(to.items), item)

	if len(rebuildBox(to, receiving)) == 0 && len(rebuildBox(from, remaining)) == 0 {
		return true
	}

	fromState.rollback()
	toState.rollback()

	return false
}
