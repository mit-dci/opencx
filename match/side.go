package match

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Side is a representation of buy or sell side
type Side bool

// iota trickery
const (
	Buy = Side(iota%2 == 0)
	Sell
)

const (
	buyString  = "buy"
	sellString = "sell"
)

// String returns the string representation of a buy or sell side
func (s *Side) String() string {
	if *s {
		return buyString
	}
	return sellString
}

func (s *Side) UnmarshalJSON(b []byte) (err error) {
	var str string
	if err = json.Unmarshal(b, &str); err != nil {
		return
	}
	switch strings.ToLower(str) {
	default:
		err = fmt.Errorf("Cannot unmarshal side json, not buy or sell")
		return
	case buyString:
		*s = Buy
	case sellString:
		*s = Sell
	}
	return
}

func (s *Side) FromString(str string) (err error) {
	switch strings.ToLower(str) {
	default:
		err = fmt.Errorf("Cannot get side from string, not buy or sell")
		return
	case buyString:
		*s = Buy
	case sellString:
		*s = Sell
	}
	return
}
