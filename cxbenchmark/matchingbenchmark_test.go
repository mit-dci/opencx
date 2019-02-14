package cxbenchmark

import (
	"testing"

	"github.com/mit-dci/opencx/logging"
)

func BenchmarkPlaceOrders(b *testing.B) {
	client := SetupBenchmark()
	var numRuns int
	b.Run("VariablePlacingAndFilling", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			numRuns += PlaceAndFill(client, "tester", "othertester", "btc/ltc", 2)
			numRuns += PlaceAndFill(client, "tester", "othertester", "btc/vtc", 2)
			numRuns += PlaceAndFill(client, "tester", "othertester", "vtc/ltc", 2)
		}
	})
	logging.Infof("Number of runs: %d", numRuns)
}
