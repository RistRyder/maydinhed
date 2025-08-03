package raft

type StoreKey interface {
}

type StoreValue interface {
}

type Store interface {
	AddNode(address, id string) error
	Delete(key StoreKey) error
	Get(key StoreKey) (StoreValue, error)
	Set(key StoreKey, value StoreValue) error
}
