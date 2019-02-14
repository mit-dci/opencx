package cxbenchmark

import (
	"fmt"
	"testing"

	"github.com/mit-dci/opencx/cmd/benchclient"
	"github.com/mit-dci/opencx/logging"
)

// func BenchmarkPlaceOrders(b *testing.B) {
//  logging.SetLogLevel(0)
// 	client := SetupBenchmark()
// 	runs := []int{1, 10, 100, 1000}
// 	for _, varRuns := range runs {
// 		testTitle := fmt.Sprintf("PlaceAndFill%d", varRuns)
// 		b.Run(testTitle, func(b *testing.B) {
// 			PlaceAndFill(client, "tester", "othertester", "btc/ltc", varRuns)
// 			PlaceAndFill(client, "tester", "othertester", "btc/vtc", varRuns)
// 			PlaceAndFill(client, "tester", "othertester", "vtc/ltc", varRuns)
// 		})
// 	}
// }

func BenchmarkPlaceBuySell1(b *testing.B) {
	logging.SetLogLevel(0)
	client := SetupBenchmark()
	for i := 0; i < b.N; i++ {
		PlaceBuySellX(client, 1)
	}
}

func BenchmarkPlaceBuySell10(b *testing.B) {
	logging.SetLogLevel(0)
	client := SetupBenchmark()
	for i := 0; i < b.N; i++ {
		PlaceBuySellX(client, 10)
	}
}

func BenchmarkPlaceBuySell100(b *testing.B) {
	logging.SetLogLevel(0)
	client := SetupBenchmark()
	for i := 0; i < b.N; i++ {
		PlaceBuySellX(client, 100)
	}
}
func BenchmarkPlaceBuySell1000(b *testing.B) {
	logging.SetLogLevel(0)
	client := SetupBenchmark()
	for i := 0; i < b.N; i++ {
		PlaceBuySellX(client, 1000)
	}
}

func BenchmarkPlaceBuyThenSell(b *testing.B) {
	logging.SetLogLevel(0)
	client := SetupBenchmark()
	runs := []int{1, 10, 100, 1000}
	for _, varRuns := range runs {
		testTitle := fmt.Sprintf("PlaceBuyThenSell%d", varRuns)
		b.Run(testTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceBuySellX(client, varRuns)
			}
		})
	}
}

func PlaceBuySellX(client *benchclient.BenchClient, varRuns int) (numRuns int) {
	PlaceManyBuy(client, "tester", "othertester", "btc/ltc", varRuns)
	PlaceManyBuy(client, "tester", "othertester", "btc/vtc", varRuns)
	PlaceManyBuy(client, "tester", "othertester", "vtc/ltc", varRuns)
	PlaceManySell(client, "tester", "othertester", "btc/ltc", varRuns)
	PlaceManySell(client, "tester", "othertester", "btc/vtc", varRuns)
	PlaceManySell(client, "tester", "othertester", "vtc/ltc", varRuns)
	return
}
