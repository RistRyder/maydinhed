package stores

import (
	"github.com/ristryder/maydinhed/mapping"
)

type Store[K mapping.ClusteredMarkerKey] interface {
	Delete(key K) error
	Get(key K) (mapping.Location, error)
	Set(key K, value mapping.Location) error
	SetAndForget(key K, value mapping.Location)
}
