package stores

import (
	"github.com/govalues/decimal"
)

type Location struct {
	Latitude  decimal.Decimal
	Longitude decimal.Decimal
}

var InvalidLocation Location = Location{
	Latitude:  decimal.Thousand,
	Longitude: decimal.Thousand,
}
