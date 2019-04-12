package cxbenchmark

import (
	"fmt"
	"testing"
)

func BenchmarkPlaceOrders(b *testing.B) {
	client := SetupBenchmark()
	runs := []int{1, 10, 100, 1000}
	for _, varRuns := range runs {
		placeFillTitle := fmt.Sprintf("PlaceAndFill%d", varRuns)
		b.Run(placeFillTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceFillX(client, varRuns)
			}
		})
		placeBuySellTitle := fmt.Sprintf("PlaceBuyThenSell%d", varRuns)
		b.Run(placeBuySellTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceBuySellX(client, varRuns)
			}
		})
	}
}

func BenchmarkNoisePlaceOrders(b *testing.B) {
	client := SetupNoiseBenchmark()
	runs := []int{1, 10, 100, 1000}
	for _, varRuns := range runs {
		placeFillTitle := fmt.Sprintf("PlaceAndFill%d", varRuns)
		b.Run(placeFillTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceFillX(client, varRuns)
			}
		})
		placeBuySellTitle := fmt.Sprintf("PlaceBuyThenSell%d", varRuns)
		b.Run(placeBuySellTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceBuySellX(client, varRuns)
			}
		})
	}
}
