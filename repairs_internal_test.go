package boxpacker3

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func spreadOrder(random *rand.Rand) ([]*Box, []*piece) {
	boxes := make([]*Box, 0, 6)

	for i := range 6 {
		boxes = append(boxes, NewBox(fmt.Sprintf("crate-%d", i), 300, 300, 300, 1e9))
	}

	items := make([]*piece, 0, 16)

	for i := range 6 + random.Intn(10) {
		side := float64(101 + random.Intn(48))
		items = append(items, testPiece(fmt.Sprintf("item-%d", i), side, side, side, 10))
	}

	return boxes, items
}

func TestEmptyABox_CollapsesASpreadPacking(t *testing.T) {
	t.Parallel()

	freed, orders := 0, 0

	for seed := range 200 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec
		boxes, items := spreadOrder(random)

		result := rawPack(t, NewPacker(WithAlgorithm(NewGreedy(OrderIncreasing, SelectWorstFit))), boxes, items)

		before := countUsedBoxes(result.boxes)
		repairPacking(t, result)
		after := countUsedBoxes(result.boxes)

		require.LessOrEqualf(t, after, before,
			"seed %d: emptying opened boxes, %d became %d", seed, before, after)

		if after < before {
			freed += before - after
			orders++
		}
	}

	t.Logf("emptied %d boxes across %d of 200 orders", freed, orders)

	require.Positive(t, freed, "the emptying pass never freed a box; it would be dead weight")
}

func TestEmptyABox_KeepsEveryPiece(t *testing.T) {
	t.Parallel()

	for seed := range 200 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec
		boxes, items := spreadOrder(random)

		result := rawPack(t, NewPacker(WithAlgorithm(NewGreedy(OrderIncreasing, SelectWorstFit))), boxes, items)

		repairPacking(t, result)

		requireWholeOrder(t, seed, items, result)
		requireNoOverlap(t, seed, result)
	}
}

func requireWholeOrder(t *testing.T, seed int, offered []*piece, result *Packing) {
	t.Helper()

	wanted := map[string]int{}
	for _, item := range offered {
		wanted[item.id]++
	}

	answered := map[string]int{}

	for _, box := range result.boxes {
		for _, item := range box.items {
			answered[item.id]++
		}
	}

	for _, item := range result.unfit {
		answered[item.id]++
	}

	require.Equalf(t, wanted, answered,
		"seed %d: the boxes and the unpacked list together are not the order that was offered", seed)
}

func requireNoOverlap(t *testing.T, seed int, result *Packing) {
	t.Helper()

	for _, box := range result.boxes {
		for _, one := range box.items {
			for _, other := range box.items {
				if one == other {
					continue
				}

				require.Falsef(t, one.intersects(other),
					"seed %d: %s and %s overlap after emptying", seed, one.id, other.id)
			}
		}
	}
}

func TestEmptyABox_LeavesATightPackingAlone(t *testing.T) {
	t.Parallel()

	box := NewBox("crate", 300, 300, 300, 1e9)
	items := []*piece{
		testPiece("a", 150, 150, 150, 10),
		testPiece("b", 150, 150, 150, 10),
	}

	result := rawPack(t, NewPacker(), []*Box{box}, items)

	positions := map[string]Pivot{}
	for _, packed := range result.boxes[0].items {
		positions[packed.id] = packed.position
	}

	repairPacking(t, result)

	for _, packed := range result.boxes[0].items {
		require.Equalf(t, positions[packed.id], packed.position,
			"%s moved in a packing that could not be improved", packed.id)
	}
}

func crowdedOrder(random *rand.Rand) ([]*Box, []*piece) {
	boxes := make([]*Box, 0, 3)

	for i := range 3 {
		boxes = append(boxes, NewBox(fmt.Sprintf("crate-%d", i), 300, 300, 300, 1e9))
	}

	items := make([]*piece, 0, 24)

	for i := range 18 + random.Intn(6) {
		side := float64(120 + random.Intn(60))
		items = append(items, testPiece(fmt.Sprintf("item-%d", i), side, side, side, 10))
	}

	return boxes, items
}

func TestEmptyABox_OffersLeftBehindGoodsToAFreedBox(t *testing.T) {
	t.Parallel()

	rescued, orders := 0, 0

	for seed := range 200 {
		random := rand.New(rand.NewSource(int64(seed))) //nolint:gosec
		boxes, items := crowdedOrder(random)

		result := rawPack(t, NewPacker(WithAlgorithm(NewGreedy(OrderIncreasing, SelectWorstFit))), boxes, items)

		before := len(result.unfit)
		repairPacking(t, result)
		after := len(result.unfit)

		require.LessOrEqualf(t, after, before,
			"seed %d: emptying left more goods behind, %d became %d", seed, before, after)

		requireWholeOrder(t, seed, items, result)
		requireNoOverlap(t, seed, result)

		if after < before {
			rescued += before - after
			orders++
		}
	}

	t.Logf("packed %d goods that nothing would take, across %d of 200 orders", rescued, orders)
}
