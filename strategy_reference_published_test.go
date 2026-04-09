//go:build measure

package boxpacker3_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type brItem struct {
	ID     string  `json:"id"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Depth  float64 `json:"depth"`
	Weight float64 `json:"weight"`
}

type brProblem struct {
	ID     string   `json:"id"`
	Width  float64  `json:"width"`
	Height float64  `json:"height"`
	Depth  float64  `json:"depth"`
	Weight float64  `json:"weight"`
	Items  []brItem `json:"items"`
}

func brFixture(problem publishedProblem) brProblem {
	one := brProblem{
		ID:     problem.ID,
		Width:  problem.Box.Width(),
		Height: problem.Box.Height(),
		Depth:  problem.Box.Depth(),
		Weight: problem.Box.MaxWeight(),
		Items:  make([]brItem, 0, len(problem.Items)),
	}

	for _, item := range problem.Items {
		one.Items = append(one.Items, brItem{
			ID:     item.ID(),
			Width:  item.Width(),
			Height: item.Height(),
			Depth:  item.Depth(),
			Weight: item.Weight(),
		})
	}

	return one
}

func TestWriteBRFixtures(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"br1.txt", fileBR2, fileBR3, "br4.txt",
		fileBR5, fileBR6, "br7.txt", "loh-nee.txt",
	} {
		problems := loadPublishedInstances(t, name)
		out := make([]brProblem, 0, len(problems))

		for _, problem := range problems {
			out = append(out, brFixture(problem))
		}

		encoded, err := json.Marshal(out)
		if err != nil {
			t.Fatal(err)
		}

		path := filepath.Join(referenceDir, name+".json")

		err = os.WriteFile(path, encoded, 0o600)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("wrote %s: %d problems", path, len(out))
	}
}
