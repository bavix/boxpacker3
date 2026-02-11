package boxpacker3

import (
	"cmp"
	"slices"
	"strings"
)

type blockKey struct {
	whd          [3]float64
	rotation     Rotation
	verticalAxes string
	weight       float64
	class        string
	group        string
	separateFrom string
	maxLoadOnTop float64
	nothingOnTop bool
}

func keyOf(item *piece) blockKey {
	axes := make([]byte, 0, len(item.verticalAxes))

	for _, axis := range item.verticalAxes {
		axes = append(axes, "whd"[axis%3])
	}

	return blockKey{
		whd:          item.whd,
		rotation:     item.rotation,
		verticalAxes: string(axes),
		weight:       item.weight,
		class:        item.class,
		group:        item.group,
		separateFrom: strings.Join(item.separateFrom, "\x00"),
		maxLoadOnTop: item.maxLoadOnTop,
		nothingOnTop: item.nothingOnTop,
	}
}

type stock struct {
	kinds  []blockKey
	sample map[blockKey]*piece
	left   map[blockKey][]*piece
}

func stockOf(items []*piece) *stock {
	held := &stock{
		kinds:  make([]blockKey, 0, len(items)),
		sample: make(map[blockKey]*piece, len(items)),
		left:   make(map[blockKey][]*piece, len(items)),
	}

	for _, item := range items {
		key := keyOf(item)

		if _, known := held.sample[key]; !known {
			held.kinds = append(held.kinds, key)
			held.sample[key] = item
		}

		held.left[key] = append(held.left[key], item)
	}

	slices.SortStableFunc(held.kinds, func(a, b blockKey) int {
		return cmp.Compare(volumeOfKey(b), volumeOfKey(a))
	})

	return held
}

func volumeOfKey(key blockKey) float64 {
	return key.whd[0] * key.whd[1] * key.whd[2]
}

func (s *stock) count(key blockKey) int {
	return len(s.left[key])
}

func (s *stock) take(key blockKey, count int) []*piece {
	taken := s.left[key][:count]
	s.left[key] = s.left[key][count:]

	return taken
}

func (s *stock) give(key blockKey, pieces []*piece) {
	s.left[key] = append(slices.Clone(pieces), s.left[key]...)
}

func (s *stock) remaining() []*piece {
	left := make([]*piece, 0)

	for _, key := range s.kinds {
		left = append(left, s.left[key]...)
	}

	return left
}

func (s *stock) empty() bool {
	for _, key := range s.kinds {
		if len(s.left[key]) > 0 {
			return false
		}
	}

	return true
}

type block struct {
	key         blockKey
	orientation Orientation
	counts      [3]int
	each        Dimension
}

func (b block) size() Dimension {
	return Dimension{
		b.each[WidthAxis] * float64(b.counts[WidthAxis]),
		b.each[HeightAxis] * float64(b.counts[HeightAxis]),
		b.each[DepthAxis] * float64(b.counts[DepthAxis]),
	}
}

func (b block) pieces() int {
	return b.counts[WidthAxis] * b.counts[HeightAxis] * b.counts[DepthAxis]
}

func (b block) volume() float64 {
	size := b.size()

	return size[WidthAxis] * size[HeightAxis] * size[DepthAxis]
}

func blocksAt(box *Container, held *stock, key blockKey, position Pivot, limit int) []block {
	sample := held.sample[key]
	available := held.count(key)

	built := make([]block, 0, limit)

	for _, orientation := range sample.rotations {
		each := rotatedDimension(sample.Item, orientation)
		built = appendStacks(built, box, key, orientation, each, position, available)
	}

	slices.SortStableFunc(built, func(a, b block) int {
		return cmp.Compare(b.volume(), a.volume())
	})

	return built[:min(limit, len(built))]
}

func appendStacks(
	built []block, box *Container, key blockKey, orientation Orientation,
	each Dimension, position Pivot, available int,
) []block {
	room := [3]int{}

	for _, axis := range allAxes() {
		if each[axis] <= minDimension {
			return built
		}

		room[axis] = int((box.boxDimension(axis) - position[axis]) / each[axis])
	}

	for x := 1; x <= room[WidthAxis]; x++ {
		for y := 1; y <= room[HeightAxis]; y++ {
			for z := 1; z <= room[DepthAxis]; z++ {
				if x*y*z > available {
					continue
				}

				built = append(built, block{
					key: key, orientation: orientation,
					counts: [3]int{x, y, z}, each: each,
				})
			}
		}
	}

	return built
}

func (c *Container) eachCopy(candidate block, position Pivot, visit func(Pivot) bool) bool {
	for z := range candidate.counts[DepthAxis] {
		for y := range candidate.counts[HeightAxis] {
			for x := range candidate.counts[WidthAxis] {
				at := Pivot{
					position[WidthAxis] + candidate.each[WidthAxis]*float64(x),
					position[HeightAxis] + candidate.each[HeightAxis]*float64(y),
					position[DepthAxis] + candidate.each[DepthAxis]*float64(z),
				}

				if !visit(at) {
					return false
				}
			}
		}
	}

	return true
}

func (c *Container) placeBlock(candidate block, held *stock, position Pivot) ([]*piece, bool) {
	sample := held.sample[candidate.key]
	allowUnstable := !c.hasStableOrientation(sample)

	state := saveContainer(c)
	origins := held.take(candidate.key, candidate.pieces())
	next := 0

	laid := c.eachCopy(candidate, position, func(at Pivot) bool {
		copied := clonePtr(origins[next])

		if !c.canQuota(copied) || !c.acceptsPlacement(copied, at, candidate.each, allowUnstable) {
			return false
		}

		c.place(copied, at, candidate.orientation)

		next++

		return true
	})

	if !laid {
		state.rollback()
		held.give(candidate.key, origins)

		return nil, false
	}

	return origins, true
}
