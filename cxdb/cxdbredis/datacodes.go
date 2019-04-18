package cxdbredis

const (
	// Token is used as a prefix for queries for auth tokens
	Token = 0x00
	// Account is used as a prefix for queries for username / account queries
	Account = 0x01
	// Balance is a prefix used for balance of a specific coin
	Balance = 0x02
)

const (
	// Bitcoin is a prefix used with Balance for bitcoin tokens
	Bitcoin = 0x00
	// Litecoin is a prefix used with Balance for Litecoin tokens
	Litecoin = 0x01
	// Vertcoin is a prefix used with Balance for Vertcoin tokens
	Vertcoin = 0x02
	// Dogecoin is a prefix used with Balance for Dogecoin tokens
	Dogecoin = 0x03
)
