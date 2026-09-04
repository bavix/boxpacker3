package boxpacker3_test

import (
	"fmt"
	"testing"

	"github.com/bavix/boxpacker3/v2"
)

func qualityStrategies() []struct {
	name     string
	strategy boxpacker3.RuleSettings
} {
	return []struct {
		name     string
		strategy boxpacker3.RuleSettings
	}{
		{nameFirstFitDecreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectFirstFit}},
		{nameFirstFitIncreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectFirstFit}},
		{nameBestFitIncreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectBestFit}},
		{nameBestFitDecreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderDecreasing, Selection: boxpacker3.SelectBestFit}},
		{nameNextFitIncreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectNextFit}},
		{nameWorstFitIncreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectWorstFit}},
		{nameAlmostWorstFitIncreasing, boxpacker3.RuleSettings{Order: boxpacker3.OrderIncreasing, Selection: boxpacker3.SelectAlmostWorstFit}},
	}
}

const benchmarkProblems = 20

func BenchmarkPublished_Quality(b *testing.B) {
	problems := loadPublishedInstances(b, fileBR1)
	if len(problems) > benchmarkProblems {
		problems = problems[:benchmarkProblems]
	}

	for _, entry := range qualityStrategies() {
		b.Run(entry.name, func(b *testing.B) {
			packer := rulePacker(entry.strategy.Order, entry.strategy.Selection)

			var (
				utilisation float64
				leftBehind  int
			)

			b.ReportAllocs()

			for b.Loop() {
				utilisation, leftBehind = 0, 0

				for _, problem := range problems {
					result, err := packer.Pack(b.Context(),
						[]*boxpacker3.Box{problem.Box}, problem.Items)
					if err != nil {
						b.Fatal(err)
					}

					utilisation += volumeUtilisation(result.Boxes[0])
					leftBehind += len(result.Unpacked)
				}
			}

			b.StopTimer()

			count := float64(len(problems))
			b.ReportMetric(utilisation/count*100, "fill%")
			b.ReportMetric(float64(leftBehind)/count, "left-behind/problem")
		})
	}
}

func BenchmarkPublished_Scaling(b *testing.B) {
	for _, name := range []string{fileBR1, fileBR4, fileBR7} {
		problems := loadPublishedInstances(b, name)
		if len(problems) > benchmarkProblems {
			problems = problems[:benchmarkProblems]
		}

		b.Run(name, func(b *testing.B) {
			packer := boxpacker3.NewPacker()

			var utilisation float64

			b.ReportAllocs()

			for b.Loop() {
				utilisation = 0

				for _, problem := range problems {
					result, err := packer.Pack(b.Context(),
						[]*boxpacker3.Box{problem.Box}, problem.Items)
					if err != nil {
						b.Fatal(err)
					}

					utilisation += volumeUtilisation(result.Boxes[0])
				}
			}

			b.StopTimer()
			b.ReportMetric(utilisation/float64(len(problems))*100, "fill%")
		})
	}
}

func BenchmarkSearchBudgets(b *testing.B) {
	problems := loadPublishedInstances(b, fileBR1)
	if len(problems) > benchmarkProblems {
		problems = problems[:benchmarkProblems]
	}

	settings := []struct {
		width     int
		branching int
	}{{1, 1}, {2, 2}, {4, 3}, {8, 4}, {16, 6}}

	for _, setting := range settings {
		b.Run(fmt.Sprintf("w%d-k%d", setting.width, setting.branching), func(b *testing.B) {
			search, err := boxpacker3.NewSearch(setting.width*32, setting.branching)
			if err != nil {
				b.Fatal(err)
			}

			packer := boxpacker3.NewPacker(boxpacker3.WithAlgorithm(search))

			var utilisation float64

			b.ReportAllocs()

			for b.Loop() {
				utilisation = 0

				for _, problem := range problems {
					result, packErr := packer.Pack(b.Context(),
						[]*boxpacker3.Box{problem.Box}, problem.Items)
					if packErr != nil {
						b.Fatal(packErr)
					}

					utilisation += volumeUtilisation(result.Boxes[0])
				}
			}

			b.StopTimer()
			b.ReportMetric(utilisation/float64(len(problems))*100, "fill%")
		})
	}
}
