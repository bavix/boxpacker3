package boxpacker3

import (
	"cmp"
	"slices"
)

func placedOneLeftBehind(result *Packing) bool {
	if len(result.unfit) == 0 {
		return false
	}

	for at, item := range result.unfit {
		if !anyBoxTakes(result, item) {
			continue
		}

		result.unfit = slices.Delete(result.unfit, at, at+1)

		return true
	}

	return false
}

func anyBoxTakes(result *Packing, item *piece) bool {
	for _, box := range result.boxes {
		if box != nil && packItem(box, item, true) {
			return true
		}
	}

	return false
}

func freedOne(result *Packing) bool {
	const needsTwo = 2

	filled := usedBoxes(result.boxes)
	if len(filled) < needsTwo {
		return false
	}

	slices.SortStableFunc(filled, func(a, b *Container) int {
		return cmp.Compare(a.itemsVolume, b.itemsVolume)
	})

	for _, victim := range filled {
		if emptyInto(victim, filled) {
			return true
		}
	}

	return false
}

func emptyInto(victim *Container, filled []*Container) bool {
	states := make([]containerState, 0, len(filled))
	for _, box := range filled {
		states = append(states, saveContainer(box))
	}

	moved := victim.items
	victim.reset()

	for _, item := range moved {
		if !placedElsewhere(item, victim, filled) {
			for _, state := range states {
				state.rollback()
			}

			return false
		}
	}

	return true
}

func placedElsewhere(item *piece, victim *Container, filled []*Container) bool {
	for _, box := range filled {
		if box == victim {
			continue
		}

		if packItem(box, item, true) {
			return true
		}
	}

	return false
}
