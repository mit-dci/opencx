package chainutils

import (
	"github.com/mit-dci/lit/coinparam"
)

// HostParams are a struct that hold a param and a host, so we can automate the creation of wallets and chainhooks,
// as well as better keep track of stuff.
type HostParams struct {
	Param *coinparam.Params
	Host  string
}

// HostParamList exists so we can easily, with utils, get a list of coin params separated from the hosts.
type HostParamList []*HostParams

// CoinListFromHostParams generates a list of coinparams from an existing host param list.
func (hpList HostParamList) CoinListFromHostParams() (coinList []*coinparam.Params) {
	for _, hostParam := range hpList {
		if hostParam.Host != "" {
			coinList = append(coinList, hostParam.Param)
		}
	}
	return
}

func HostParamsFromCoinList(coinList []*coinparam.Params) (hParams []*HostParams) {
	hParams = make([]*HostParams, len(coinList))
	// if you want to go in circles with this method and CoinListFromHostParams, you're losing information
	for i, coin := range coinList {
		hParams[i] = &HostParams{
			Param: coin,
			// Default host
			Host: "",
		}
	}
	return
}
