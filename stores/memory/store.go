package memory

import (
	"log"

	"github.com/ristryder/maydinhed/mapping"
)

type MemoryStore[K mapping.ClusteredMarkerKey] struct {
	store map[K]mapping.Location
}

func New[K mapping.ClusteredMarkerKey]() (*MemoryStore[K], error) {
	return &MemoryStore[K]{
		store: make(map[K]mapping.Location),
	}, nil
}

func (v *MemoryStore[K]) Delete(key K) error {
	//TODO: Do
	log.Printf("memory store Delete key %v\n", key)
	return nil
}

func (v *MemoryStore[K]) Get(key K) (mapping.Location, error) {
	//TODO: Do
	log.Printf("memory store Get key %v\n", key)
	return mapping.InvalidLocation, nil
}

func (v *MemoryStore[K]) Set(key K, value mapping.Location) error {
	//TODO: Do
	log.Printf("memory store Set key %v\n", key)
	return nil
}

func (v *MemoryStore[K]) SetAndForget(key K, value mapping.Location) {
	v.Set(key, value)
}
