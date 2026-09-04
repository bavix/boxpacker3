package boxpacker3

import "slices"

type space struct {
	origin Pivot
	size   Dimension
}

func (s space) far(axis Axis) float64 {
	return s.origin[axis] + s.size[axis]
}

func (s space) holds(dimension Dimension) bool {
	for _, axis := range allAxes() {
		if dimension[axis] > s.size[axis]+s.size[axis]*dimensionEpsilon {
			return false
		}
	}

	return true
}

func (s space) within(other space) bool {
	for axis := range 3 {
		slack := max(s.size[axis], other.size[axis]) * dimensionEpsilon

		if s.origin[axis] < other.origin[axis]-slack {
			return false
		}

		if s.origin[axis]+s.size[axis] > other.origin[axis]+other.size[axis]+slack {
			return false
		}
	}

	return true
}

func (s space) meets(position Pivot, dimension Dimension) bool {
	return overlaps(s.origin, s.size, position, dimension)
}

func (c *Container) resetSpaces() {
	c.spaces = append(c.spaces[:0], space{
		origin: Pivot{},
		size:   Dimension{c.width, c.height, c.depth},
	})
}

func (c *Container) refreshSpaces(placed *piece) {
	position := placed.position
	dimension := placed.dimension()

	kept := c.spaces[:0]
	slabs := make([]space, 0, 6) //nolint:mnd // two slabs per axis at most.

	for _, free := range c.spaces {
		if !free.meets(position, dimension) {
			kept = append(kept, free)

			continue
		}

		slabs = appendSlabs(slabs, free, position, dimension)
	}

	c.spaces = addSlabs(kept, slabs)
}

func addSlabs(kept []space, slabs []space) []space {
	for i, slab := range slabs {
		if tooThin(slab) || swallowedBy(kept, slab) || swallowedByEarlier(slabs, i, slab) {
			continue
		}

		write := 0

		for _, free := range kept {
			if free.within(slab) {
				continue
			}

			kept[write] = free
			write++
		}

		kept = append(kept[:write], slab)
	}

	return kept
}

func tooThin(free space) bool {
	return free.size[WidthAxis] <= minDimension ||
		free.size[HeightAxis] <= minDimension ||
		free.size[DepthAxis] <= minDimension
}

func swallowedBy(spaces []space, slab space) bool {
	for _, free := range spaces {
		if slab.within(free) {
			return true
		}
	}

	return false
}

func swallowedByEarlier(slabs []space, at int, slab space) bool {
	for j, other := range slabs[:at] {
		if slab.within(other) && (!other.within(slab) || j <= at) {
			return true
		}
	}

	return false
}

func appendSlabs(kept []space, free space, position Pivot, dimension Dimension) []space {
	for _, axis := range allAxes() {
		if near := position[axis] - free.origin[axis]; near > 0 {
			slab := free
			slab.size[axis] = near
			kept = append(kept, slab)
		}

		far := position[axis] + dimension[axis]
		if room := free.far(axis) - far; room > 0 {
			slab := free
			slab.origin[axis] = far
			slab.size[axis] = room
			kept = append(kept, slab)
		}
	}

	return kept
}

func spaceCornersInto(into *[4]Pivot, free space, dimension Dimension) []Pivot {
	corners := into[:0]
	corners = append(corners, free.origin)

	for _, axis := range allAxes() {
		room := free.size[axis] - dimension[axis]
		if room <= minDimension {
			continue
		}

		corner := free.origin
		corner[axis] += room
		corners = append(corners, corner)
	}

	return corners
}

func (c *Container) eachPlacementPoint(dimension Dimension, visit func(Pivot)) {
	for _, point := range c.points {
		visit(point)
	}

	if !c.rules.FreeSpaceCorners {
		return
	}

	var corners [4]Pivot

	for _, free := range c.spaces {
		if !free.holds(dimension) {
			continue
		}

		for _, corner := range spaceCornersInto(&corners, free, dimension) {
			if corner == free.origin && slices.Contains(c.points, corner) {
				continue
			}

			visit(corner)
		}
	}
}
