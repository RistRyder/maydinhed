package stores

import (
	"github.com/ristryder/maydinhed/mapping"
)

type Command[K mapping.ClusteredMarkerKey] struct {
	Key       K                `json:"key,omitempty"`
	Operation string           `json:"op,omitempty"`
	Value     mapping.Location `json:"value,omitempty"`
}
