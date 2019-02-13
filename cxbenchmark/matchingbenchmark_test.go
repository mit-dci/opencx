package cxbenchmark

import (
	"testing"
)

func BenchmarkPlaceOrders1000(b *testing.B) {
	client := SetupBenchmark()
	for i := 0; i < b.N; i++ {
		PlaceAndFill(client, "tester", "othertester", "btc/ltc", 2)
		PlaceAndFill(client, "tester", "othertester", "btc/vtc", 2)
		PlaceAndFill(client, "tester", "othertester", "vtc/ltc", 2)
	}
}
