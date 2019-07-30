package cxdbsql

import (
	"testing"
)

func TestCreateDepositStoreAllParams(t *testing.T) {
	var err error

	var tc *testerContainer
	if tc, err = CreateTesterContainer(); err != nil {
		t.Errorf("Error creating tester container: %s", err)
	}

	defer func() {
		if err = tc.Kill(); err != nil {
			t.Errorf("Error killing tester container: %s", err)
			return
		}
	}()

	var ds *SQLDepositStore
	for _, coin := range constCoinParams() {
		if ds, err = CreateDepositStoreStructWithConf(coin, testConfig()); err != nil {
			t.Errorf("Error creating deposit store for coin: %s", err)
		}

		if err = ds.DestroyHandler(); err != nil {
			t.Errorf("Error destroying handler for deposit store: %s", err)
		}
	}

}
