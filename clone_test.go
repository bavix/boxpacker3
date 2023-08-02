package boxpacker3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopyPtr(t *testing.T) {
	a := 1
	b := struct{ a int }{}

	require.Same(t, &a, &a)
	require.Same(t, &b, &b)
	require.NotSame(t, &a, copyPtr(&a))
	require.NotSame(t, &b, copyPtr(&b))
}

func TestCopySlicePtr(t *testing.T) {
	a := struct{ a int }{}
	b := struct{ a int }{}
	c := []*struct{ a int }{&a, &b}

	d := make([]*struct{ a int }, len(c))
	copy(d, c)

	e := copySlicePtr(c)

	for i := range d {
		require.Same(t, c[i], d[i])
		require.NotSame(t, c[i], e[i])
	}
}
