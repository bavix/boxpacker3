package boxpacker3_test

import (
	"bufio"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

type publishedProblem struct {
	ID    string
	Box   *boxpacker3.Box
	Items []*boxpacker3.Item
}

const (
	fileBR1 = "br1.txt"
	fileBR2 = "br2.txt"
	fileBR3 = "br3.txt"
	fileBR4 = "br4.txt"
	fileBR5 = "br5.txt"
	fileBR6 = "br6.txt"
	fileBR7 = "br7.txt"
)

func loadPublishedInstances(tb testing.TB, name string) []publishedProblem {
	tb.Helper()

	file, err := os.Open(filepath.Join("testdata", name))
	require.NoError(tb, err)

	defer func() { require.NoError(tb, file.Close()) }()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	fields := func() []string {
		require.True(tb, scanner.Scan(), "unexpected end of %s", name)

		return strings.Fields(scanner.Text())
	}

	number := func(s string) float64 {
		value, convErr := strconv.ParseFloat(s, 64)
		require.NoError(tb, convErr)

		return value
	}

	problemCount := int(number(fields()[0]))
	problems := make([]publishedProblem, 0, problemCount)

	for range problemCount {
		problems = append(problems, readProblem(fields, number))
	}

	require.NoError(tb, scanner.Err())

	return problems
}

func readProblem(fields func() []string, number func(string) float64) publishedProblem {
	id := fields()[0]

	dimensions := fields()
	box := boxpacker3.NewBox(
		"container-"+id,
		number(dimensions[0]),
		number(dimensions[1]),
		number(dimensions[2]),
		math.MaxFloat64,
	)

	typeCount := int(number(fields()[0]))
	items := make([]*boxpacker3.Item, 0, typeCount)

	for range typeCount {
		spec := fields()

		for copyIndex := range int(number(spec[7])) {
			item, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
				ID:           "item-" + spec[0] + "-" + strconv.Itoa(copyIndex),
				Width:        number(spec[1]),
				Height:       number(spec[3]),
				Depth:        number(spec[5]),
				Weight:       0,
				VerticalAxes: verticalAxes(spec, number),
			})
			if err != nil {
				panic(err)
			}

			items = append(items, item)
		}
	}

	return publishedProblem{ID: id, Box: box, Items: items}
}

func verticalAxes(spec []string, number func(string) float64) []boxpacker3.Axis {
	flags := [3]struct {
		axis boxpacker3.Axis
		set  bool
	}{
		{boxpacker3.WidthAxis, number(spec[2]) != 0},
		{boxpacker3.HeightAxis, number(spec[4]) != 0},
		{boxpacker3.DepthAxis, number(spec[6]) != 0},
	}

	axes := make([]boxpacker3.Axis, 0, len(flags))

	for _, flag := range flags {
		if flag.set {
			axes = append(axes, flag.axis)
		}
	}

	if len(axes) == 0 {
		return []boxpacker3.Axis{boxpacker3.WidthAxis, boxpacker3.HeightAxis, boxpacker3.DepthAxis}
	}

	return axes
}
