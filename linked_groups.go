package boxpacker3

import (
	"maps"
	"slices"
)

func hasGroups(result *Packing) bool {
	for _, box := range result.boxes {
		for _, item := range box.items {
			if item.group != "" {
				return true
			}
		}
	}

	for _, item := range result.unfit {
		if item.group != "" {
			return true
		}
	}

	return false
}

func groupMembers(result *Packing) map[string][]*piece {
	members := map[string][]*piece{}

	for _, box := range result.boxes {
		for _, item := range box.items {
			if item.group != "" {
				members[item.group] = append(members[item.group], item)
			}
		}
	}

	for _, item := range result.unfit {
		if item.group != "" {
			members[item.group] = append(members[item.group], item)
		}
	}

	return members
}

func boxHoldingAll(result *Packing, members []*piece) *Container {
	var holder *Container

	for _, box := range result.boxes {
		for _, item := range box.items {
			if !slices.Contains(members, item) {
				continue
			}

			if holder != nil && holder != box {
				return nil
			}

			holder = box
		}
	}

	if holder == nil {
		return nil
	}

	for _, member := range members {
		if !slices.Contains(holder.items, member) {
			return nil
		}
	}

	return holder
}

func detachMembers(result *Packing, members []*piece) []*piece {
	stranded := make([]*piece, 0)

	for _, box := range result.boxes {
		kept := make([]*piece, 0, len(box.items))
		removed := false

		for _, item := range box.items {
			if slices.Contains(members, item) {
				removed = true

				continue
			}

			kept = append(kept, item)
		}

		if removed {
			stranded = append(stranded, rebuildBox(box, kept)...)
		}
	}

	result.unfit = slices.DeleteFunc(result.unfit, func(item *piece) bool {
		return slices.Contains(members, item)
	})

	return stranded
}

func rebuildBox(b *Container, items []*piece) []*piece {
	b.reset()

	unplaced := make([]*piece, 0)

	for _, item := range items {
		if !fitInSpecificBox(b, item) {
			unplaced = append(unplaced, item)
		}
	}

	return unplaced
}

func placeGroupTogether(result *Packing, members []*piece) bool {
	for _, box := range result.boxes {
		state := saveContainer(box)
		memberPlacements := snapshot(members)

		placed := true

		for _, member := range members {
			if !fitInSpecificBox(box, member) {
				placed = false

				break
			}
		}

		if placed {
			return true
		}

		restore(members, memberPlacements)
		state.rollback()
	}

	return false
}

const maxGroupRepairs = 2

func allUnfit(result *Packing, members []*piece) bool {
	for _, member := range members {
		if !slices.Contains(result.unfit, member) {
			return false
		}
	}

	return true
}

func splitGroup(result *Packing, members map[string][]*piece, spent map[string]int) string {
	for _, group := range slices.Sorted(maps.Keys(members)) {
		if spent != nil && spent[group] >= maxGroupRepairs {
			continue
		}

		if boxHoldingAll(result, members[group]) != nil || allUnfit(result, members[group]) {
			continue
		}

		return group
	}

	return ""
}

func gatherGroups(result *Packing) []*piece {
	spent := map[string]int{}
	stranded := make([]*piece, 0)

	for {
		members := groupMembers(result)

		target := splitGroup(result, members, spent)
		if target == "" {
			return stranded
		}

		group := members[target]
		spent[target]++

		lost := detachMembers(result, group)
		result.unfit = append(result.unfit, lost...)
		stranded = append(stranded, lost...)

		if !placeGroupTogether(result, group) {
			leaveBehind(result, group)
		}
	}
}

func leaveBehind(result *Packing, group []*piece) {
	for _, member := range group {
		member.setReason(ReasonGroupIncomplete)
	}

	result.unfit = append(result.unfit, group...)
}

func dropSplitGroups(result *Packing) []*piece {
	stranded := make([]*piece, 0)

	for {
		members := groupMembers(result)

		target := splitGroup(result, members, nil)
		if target == "" {
			return stranded
		}

		group := members[target]
		lost := detachMembers(result, group)
		result.unfit = append(result.unfit, lost...)
		result.unfit = append(result.unfit, group...)
		stranded = append(stranded, lost...)
	}
}

func reofferStranded(result *Packing, stranded []*piece) {
	for _, item := range stranded {
		if item == nil || item.group != "" || !anyBoxTakes(result, item) {
			continue
		}

		at := slices.Index(result.unfit, item)
		if at >= 0 {
			result.unfit = slices.Delete(result.unfit, at, at+1)
		}
	}
}

func enforceGroups(result *Packing) {
	if result == nil || !hasGroups(result) {
		return
	}

	stranded := gatherGroups(result)
	stranded = append(stranded, dropSplitGroups(result)...)
	reofferStranded(result, stranded)
}
