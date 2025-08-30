package messaging

import (
	"github.com/ristryder/maydinhed/mapping"
)

type Messenger[K mapping.ClusteredMarkerKey] interface {
	Close() error
	SendLocationUpdate(keyedLocation mapping.KeyedLocation[K]) error
	StartListening() error
	StopListening() error
}
