package cxdbmemory

import (
	"fmt"
	"sync"

	"github.com/Rjected/lit/coinparam"
	"github.com/mit-dci/opencx/match"
)

type MemorySettlementEngine struct {
	// Balances
	balances    map[[33]byte]uint64
	balancesMtx *sync.Mutex

	// this coin
	coin *coinparam.Params
}

// CreateSettlementEngine creates a settlement engine for a specific coin
func CreateSettlementEngine(coin *coinparam.Params) (engine match.SettlementEngine, err error) {

	// Set values
	me := &MemorySettlementEngine{
		balances:    make(map[[33]byte]uint64),
		balancesMtx: new(sync.Mutex),
		coin:        coin,
	}

	// Now we actually set what we want
	engine = me
	return
}

// ApplySettlementExecution applies the settlementExecution, this assumes that the settlement execution is
// valid
func (me *MemorySettlementEngine) ApplySettlementExecution(setExec *match.SettlementExecution) (setRes *match.SettlementResult, err error) {

	me.balancesMtx.Lock()
	var curBal uint64
	var ok bool
	if curBal, ok = me.balances[setExec.Pubkey]; !ok && setExec.Type == match.Credit {
		err = fmt.Errorf("Trying to apply settlement execution credit to order with no balance")
		me.balancesMtx.Unlock()
		return
	}

	var newBal uint64
	if setExec.Type == match.Debit {
		newBal = curBal + setExec.Amount
	} else if setExec.Type == match.Credit {
		newBal = curBal - setExec.Amount
	}

	me.balances[setExec.Pubkey] = newBal
	me.balancesMtx.Unlock()
	// Finally set return value
	setRes = &match.SettlementResult{
		NewBal:         newBal,
		SuccessfulExec: setExec,
	}

	return
}

// CheckValid returns true if the settlement execution would be valid
func (me *MemorySettlementEngine) CheckValid(setExec *match.SettlementExecution) (valid bool, err error) {
	if setExec.Type == match.Debit {
		// No settlement will be an invalid debit
		valid = true
		return
	}
	me.balancesMtx.Lock()
	curBal := me.balances[setExec.Pubkey]
	me.balancesMtx.Unlock()
	valid = setExec.Amount > curBal
	return
}

// CreateSettlementEngineMap creates a map of coin to settlement engine, given a list of coins.
func CreateSettlementEngineMap(coins []*coinparam.Params) (setMap map[*coinparam.Params]match.SettlementEngine, err error) {

	setMap = make(map[*coinparam.Params]match.SettlementEngine)
	var curSetEng match.SettlementEngine
	for _, coin := range coins {
		if curSetEng, err = CreateSettlementEngine(coin); err != nil {
			err = fmt.Errorf("Error creating single settlement engine while creating settlement engine map: %s", err)
			return
		}
		setMap[coin] = curSetEng
	}

	return
}
