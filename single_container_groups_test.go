package boxpacker3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bavix/boxpacker3/v2"
)

func TestSingleBox_GroupsStayWholeWhenTheSearchDropsAMember(t *testing.T) {
	t.Parallel()

	box := boxpacker3.NewBox("b", 100, 100, 100, 1000)

	oversized, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
		ID: "oversized", Width: 150, Height: 150, Depth: 150, Weight: 1, Group: "pair",
	})
	require.NoError(t, err)

	small, err := boxpacker3.NewItemFromSpec(boxpacker3.ItemSpec{
		ID: "small", Width: 10, Height: 10, Depth: 10, Weight: 1, Group: "pair",
	})
	require.NoError(t, err)

	loose := boxpacker3.NewItem("loose", 20, 20, 20, 1)

	result, err := singlePacker().Pack(context.Background(), []*boxpacker3.Box{box}, []*boxpacker3.Item{oversized, small, loose})
	require.NoError(t, err)

	packed := map[string]bool{}
	for _, item := range result.Boxes[0].Items {
		packed[item.ID()] = true
	}

	require.True(t, packed["loose"])
	require.Equal(t, packed["oversized"], packed["small"], "the pair must travel together or not at all")
	require.Len(t, result.Unpacked, 2)
}
