package stores

import (
	"encoding/json"

	"github.com/govalues/decimal"
)

type Location struct {
	Latitude  decimal.Decimal `json:"lat"`
	Longitude decimal.Decimal `json:"lng"`
}

var InvalidLocation Location = Location{
	Latitude:  decimal.Thousand,
	Longitude: decimal.Thousand,
}

func (l *Location) Bytes() ([]byte, error) {
	return json.Marshal(l)
}
