package cxbenchmark

import (
	"fmt"
	"testing"
)

// TODO: fix problems where you do something unexpected network-wise and the entire rpc server crashes!
func BenchmarkPlaceOrders(b *testing.B) {
	client := SetupBenchmark()
	// ugh when we run this benchmark and the server is noise then we basically crash the rpc server... need to figure out how to have that not happen, why is that fatal?
	runs := []int{}
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
	runs := []int{1}
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
