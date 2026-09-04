package boxpacker3

func rotatedDimension(item *Item, rotation Orientation) Dimension {
	return item.oriented[rotation]
}

func (c *Container) acceptsPlacement(item *piece, position Pivot, dimension Dimension, allowUnstable bool) bool {
	if !allowUnstable && !c.isStableOrientation(dimension) {
		return false
	}

	if !c.fitsAt(position, dimension) {
		return false
	}

	if c.rules.MinSupportRatio > 0 &&
		c.supportRatio(position, dimension) < c.rules.MinSupportRatio-dimensionEpsilon {
		return false
	}

	if !c.bearsTheLoad(item, position, dimension) {
		return false
	}

	if !c.takesTheClass(item) {
		return false
	}

	if len(c.rules.Placement) == 0 {
		return true
	}

	return c.rules.permits(c, item, position, dimension)
}

func bestPlacement(box *Container, item *piece) (Pivot, Orientation, bool) {
	var (
		bestRotation Orientation
		bestScore    MeritScore
		found        bool
	)

	merit := box.meritOf()
	rotations := item.rotations
	allowUnstable := !box.hasStableOrientation(item)

	bestPoint := Pivot{}

	for _, rotation := range rotations {
		dimension := rotatedDimension(item.Item, rotation)
		perfect := false

		box.eachPlacementPoint(dimension, func(point Pivot) {
			if perfect || !box.acceptsPlacement(item, point, dimension, allowUnstable) {
				return
			}

			score := merit.Score(box, point, dimension)
			if found && !merit.Better(score, bestScore) {
				return
			}

			bestPoint, bestRotation, bestScore, found = point, rotation, score, true

			perfect = score.Contact >= wholeSurface(dimension)
		})

		if perfect {
			return bestPoint, bestRotation, true
		}
	}

	if found {
		return bestPoint, bestRotation, true
	}

	return Pivot{}, bestRotation, false
}

func (c *Container) place(item *piece, position Pivot, rotation Orientation) {
	item.position = position
	item.setOrientation(rotation)
	c.insert(item)
}
