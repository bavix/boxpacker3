package boxpacker3

import "testing"

func pieceOf(item *Item) *piece {
	return newPiece(item, 0)
}

func resultOf(boxes []*Container, unfit ...*piece) *Result {
	return newPacking(boxes, unfit).result()
}

func testPiece(id string, w, h, d, wg float64) *piece {
	return pieceOf(NewItem(id, w, h, d, wg))
}

func rawPack(tb testing.TB, packer *Packer, boxes []*Box, pieces []*piece) *Packing {
	tb.Helper()

	problem := newProblem(compact(boxes), cloneSlice(pieces), packer.rules, packer.finishers, packer.budget, packer.observer)

	packing, err := packer.algorithm.Pack(tb.Context(), problem)
	if err != nil {
		tb.Fatal(err)
	}

	packing, err = problem.finish(tb.Context(), packing)
	if err != nil {
		tb.Fatal(err)
	}

	return packing
}

func problemOf(boxes []*Box, pieces []*piece) *Problem {
	return newProblem(compact(boxes), pieces, Rules{}, nil, Budget{}, nil)
}

func repairPacking(tb testing.TB, packing *Packing) {
	tb.Helper()

	problem := problemOf(nil, nil)

	for _, finisher := range []Finisher{Consolidate, Rescue} {
		err := finisher.Finish(tb.Context(), problem, packing)
		if err != nil {
			tb.Fatal(err)
		}
	}
}

func openBox(id string, w, h, d, mw float64) *Container {
	return newContainer(NewBox(id, w, h, d, mw), 0, Rules{}, nil)
}

func openUnder(box *Box) *Container {
	return newContainer(box, 0, Rules{}, nil)
}
