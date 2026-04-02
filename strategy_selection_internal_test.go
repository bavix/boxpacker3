package boxpacker3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	boxFullest  = "fullest"
	boxMiddle   = "middle"
	boxEmptiest = "emptiest"
	itemLarge   = "large"
	itemMedium  = "medium"
	itemSmall   = "small"
)

func boxesWithKnownSlack(t *testing.T) []*Container {
	t.Helper()

	fullest := openBox(boxFullest, 100, 100, 100, 1e9)
	middle := openBox(boxMiddle, 100, 100, 100, 1e9)
	emptiest := openBox(boxEmptiest, 100, 100, 100, 1e9)

	require.True(t, fullest.putItem(testPiece("fill-a", 90, 100, 100, 1), Pivot{}))
	require.True(t, middle.putItem(testPiece("fill-b", 50, 100, 100, 1), Pivot{}))
	require.True(t, emptiest.putItem(testPiece("fill-c", 10, 100, 100, 1), Pivot{}))

	require.InDelta(t, 100000.0, fullest.remainingVolume(), 1e-9)
	require.InDelta(t, 500000.0, middle.remainingVolume(), 1e-9)
	require.InDelta(t, 900000.0, emptiest.remainingVolume(), 1e-9)

	return []*Container{fullest, middle, emptiest}
}

func TestSelectBox_DefiningProperty(t *testing.T) {
	t.Parallel()

	cases := map[BoxSelection]string{
		SelectFirstFit:       boxFullest,
		SelectNextFit:        boxFullest,
		SelectBestFit:        boxFullest,
		SelectWorstFit:       boxEmptiest,
		SelectAlmostWorstFit: boxMiddle,
	}

	for selection, want := range cases {
		t.Run(selection.String(), func(t *testing.T) {
			t.Parallel()

			boxes := boxesWithKnownSlack(t)

			choice, found := selectBox(boxes, testPiece("small", 10, 10, 10, 1), selection)
			require.True(t, found)
			require.Equal(t, want, boxes[choice.index].ID())
		})
	}
}

func TestSelectBox_AlmostWorstFitFallsBackToTheOnlyBox(t *testing.T) {
	t.Parallel()

	boxes := []*Container{openBox("only", 100, 100, 100, 1e9)}

	choice, found := selectBox(boxes, testPiece("small", 10, 10, 10, 1), SelectAlmostWorstFit)
	require.True(t, found)
	require.Equal(t, "only", boxes[choice.index].ID())
}

func TestSelectBox_AlmostWorstFitDiffersFromWorstFit(t *testing.T) {
	t.Parallel()

	item := testPiece("small", 10, 10, 10, 1)

	worst, found := selectBox(boxesWithKnownSlack(t), item, SelectWorstFit)
	require.True(t, found)

	almost, found := selectBox(boxesWithKnownSlack(t), item, SelectAlmostWorstFit)
	require.True(t, found)

	require.NotEqual(t, worst.index, almost.index,
		"almost worst fit leaves the emptiest box alone")
}

func TestSelectBox_NoBoxTakesTheItem(t *testing.T) {
	t.Parallel()

	boxes := []*Container{openBox("small", 5, 5, 5, 1e9)}

	_, found := selectBox(boxes, testPiece("oversized", 50, 50, 50, 1), SelectBestFit)
	require.False(t, found)
}

func TestSelectBox_SkipsBoxesOverWeight(t *testing.T) {
	t.Parallel()

	light := openBox("light", 100, 100, 100, 5)
	heavy := openBox("heavy", 100, 100, 100, 1000)

	choice, found := selectBox([]*Container{light, heavy}, testPiece("brick", 10, 10, 10, 100), SelectBestFit)
	require.True(t, found)
	require.Equal(t, "heavy", []*Container{light, heavy}[choice.index].ID())
}

func TestSortItems(t *testing.T) {
	t.Parallel()

	build := func() []*piece {
		return []*piece{
			testPiece(itemMedium, 2, 2, 2, 1),
			testPiece(itemLarge, 3, 3, 3, 1),
			testPiece(itemSmall, 1, 1, 1, 1),
		}
	}

	ids := func(items []*piece) []string {
		out := make([]string, 0, len(items))
		for _, item := range items {
			out = append(out, item.ID())
		}

		return out
	}

	require.Equal(t, []string{itemLarge, itemMedium, itemSmall}, ids(sortItems(build(), OrderDecreasing)))
	require.Equal(t, []string{itemSmall, itemMedium, itemLarge}, ids(sortItems(build(), OrderIncreasing)))
	require.Equal(t, []string{itemMedium, itemLarge, itemSmall}, ids(sortItems(build(), OrderAsGiven)))
}

func TestStrategy_Name(t *testing.T) {
	t.Parallel()

	require.Equal(t, "BestFitDecreasing", NewGreedy(OrderDecreasing, SelectBestFit).Name())
	require.Equal(t, "NextFit", NewGreedy(OrderAsGiven, SelectNextFit).Name())
	require.Equal(t, "AlmostWorstFitIncreasing",
		NewGreedy(OrderIncreasing, SelectAlmostWorstFit).Name())
}
