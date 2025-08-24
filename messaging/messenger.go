package messaging

import (
	"github.com/ristryder/maydinhed/stores"
)

type Messenger[K stores.StoreKey] interface {
	Close() error
	SendLocationUpdate(key K, value stores.Location) error
	StartListening() error
	StopListening() error
}
