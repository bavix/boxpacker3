package boxpacker3

import "testing"

func TestLoadBearing_AClearItemRefusesEvenAWeightlessPieceAbove(t *testing.T) {
	t.Parallel()

	box := openBox("box", 10, 10, 20, 1e9)

	cake, err := NewItemFromSpec(ItemSpec{ID: "cake", Width: 10, Height: 10, Depth: 10, Weight: 5, NothingOnTop: true})
	if err != nil {
		t.Fatal(err)
	}

	feather, err := NewItemFromSpec(ItemSpec{ID: "feather", Width: 10, Height: 10, Depth: 10, Weight: 0})
	if err != nil {
		t.Fatal(err)
	}

	if !box.putItem(pieceOf(feather), Pivot{0, 0, 10}) {
		t.Fatal("the feather should sit at the top of the box")
	}

	if box.putItem(pieceOf(cake), Pivot{0, 0, 0}) {
		t.Error("the cake went under the feather although nothing may rest on it")
	}
}
