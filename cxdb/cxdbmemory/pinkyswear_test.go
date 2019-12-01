package cxdbmemory

import (
	"testing"

	"github.com/Rjected/lit/coinparam"
	"github.com/mit-dci/opencx/match"
)

var (
	testWhitelist = [][33]byte{
		// The "zero" pubkey
		[33]byte{},
	}
	btc, _         = match.AssetFromCoinParam(&coinparam.BitcoinParams)
	testExecByZero = &match.SettlementExecution{
		// The person that is on the whitelist
		Pubkey: [33]byte{},
		// Just 1BTC because why not
		Amount: uint64(100000000),
		// This could be the right or wrong param depending on our engine
		Asset: btc,
		Type:  match.Debit,
	}
)

func TestCreatePinkySwearEngine(t *testing.T) {
	var err error

	if _, err = CreatePinkySwearEngine(&coinparam.BitcoinParams, testWhitelist, false); err != nil {
		t.Errorf("Error creating pinky swear engine for TestCreatePinkySwear: %s", err)
		return
	}

	return
}

func TestPinkySwearWrongAsset(t *testing.T) {
	var err error

	var engine match.SettlementEngine
	if engine, err = CreatePinkySwearEngine(&coinparam.VertcoinParams, testWhitelist, false); err != nil {
		t.Errorf("Error creating pinky swear engine for TestPinkySwearWrongAsset: %s", err)
		return
	}

	var valid bool
	if valid, err = engine.CheckValid(testExecByZero); err != nil {
		t.Errorf("Error checking valid for TestPinkySwearWrongAsset")
		return
	}

	expected := false
	// I know I can just do if valid, but this is more explicit and it's a test. It should be explicit,
	// and tests are easy to write. The last thing we want is a test that nobody can understand
	if valid != expected {
		t.Errorf("testExecByZero input to CheckValid should have been %t but was %t", expected, valid)
		return
	}

}

func TestPinkySwearRightAsset(t *testing.T) {
	var err error

	var engine match.SettlementEngine
	if engine, err = CreatePinkySwearEngine(&coinparam.BitcoinParams, testWhitelist, false); err != nil {
		t.Errorf("Error creating pinky swear engine for TestPinkySwearRightAsset: %s", err)
		return
	}

	var valid bool
	if valid, err = engine.CheckValid(testExecByZero); err != nil {
		t.Errorf("Error checking valid for TestPinkySwearRightAsset")
		return
	}

	expected := true
	if valid != expected {
		t.Errorf("testExecByZero input to CheckValid should have been %t but was %t", expected, valid)
		return
	}

}
