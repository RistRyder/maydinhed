package messaging

import (
	"github.com/ristryder/maydinhed/stores"
)

type Messenger[K stores.StoreKey] interface {
	SendLocationUpdate(key K, value stores.Location) error
}
