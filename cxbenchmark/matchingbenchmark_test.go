package cxbenchmark

import (
	"testing"
)

func BenchmarkPlaceOrders1000(b *testing.B) {
	client := SetupBenchmark()
	for n := 0; n < b.N; n++ {
		PlaceAndFill(client, "tester", "othertester", "btc/ltc", 10)
		PlaceAndFill(client, "tester", "othertester", "btc/vtc", 10)
		PlaceAndFill(client, "tester", "othertester", "vtc/ltc", 10)
	}
}
