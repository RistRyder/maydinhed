package memory

import (
	"log"

	"github.com/ristryder/maydinhed/stores"
)

type MemoryStore[K stores.StoreKey] struct {
	store map[K]stores.Location
}

func New[K stores.StoreKey]() (*MemoryStore[K], error) {
	return &MemoryStore[K]{
		store: make(map[K]stores.Location),
	}, nil
}

func (v *MemoryStore[K]) Delete(key K) error {
	//TODO: Do
	log.Printf("memory store Delete key %v\n", key)
	return nil
}

func (v *MemoryStore[K]) Get(key K) (stores.Location, error) {
	//TODO: Do
	log.Printf("memory store Get key %v\n", key)
	return stores.InvalidLocation, nil
}

func (v *MemoryStore[K]) Set(key K, value stores.Location) error {
	//TODO: Do
	log.Printf("memory store Set key %v\n", key)
	return nil
}

func (v *MemoryStore[K]) SetAndForget(key K, value stores.Location) {
	v.Set(key, value)
}
