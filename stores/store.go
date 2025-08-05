package stores

type StoreKey comparable

type Store[K StoreKey] interface {
	Delete(key K) error
	Get(key K) (Location, error)
	Set(key K, value Location) error
}
