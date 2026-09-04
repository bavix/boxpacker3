package boxpacker3

import "testing"

type notAtTheOrigin struct{}

func (notAtTheOrigin) Allow(_ *Container, item *Item, position Pivot, _ Dimension) bool {
	if item.ID() != "newcomer" {
		return true
	}

	return position != Pivot{}
}

func TestRepackHonoursConstraints(t *testing.T) {
	t.Parallel()

	box := openBox("box", 200, 200, 200, 100000)
	box.rules = Rules{MinSupportRatio: 0, Placement: []PlacementRule{notAtTheOrigin{}}}

	seated := testPiece("seated", 50, 50, 50, 10)
	if !fitInSpecificBox(box, seated) {
		t.Fatal("the first item did not go in")
	}

	newcomer := testPiece("newcomer", 100, 100, 100, 10)
	repacked := attemptRepack(box, newcomer)

	for _, item := range box.items {
		t.Logf("repacked=%v %s at %v size %v", repacked,
			item.ID(), item.position, item.dimension())

		if item.ID() == "newcomer" && item.position == (Pivot{}) {
			t.Errorf("%s was seated at the origin, which the constraint forbids", item.ID())
		}
	}

	if !repacked {
		if len(box.items) != 1 || box.items[0].ID() != "seated" {
			t.Errorf("a refused repack left the box as %d items", len(box.items))
		}
	}
}
