package valkey

import (
	"log"

	"github.com/cockroachdb/errors"
	"github.com/ristryder/maydinhed/stores"
	"github.com/valkey-io/valkey-go"
)

type ValkeyStore[K stores.StoreKey] struct {
	client valkey.Client
}

func New[K stores.StoreKey](valkeyOptions valkey.ClientOption) (*ValkeyStore[K], error) {
	valkeyClient, valkeyClientErr := valkey.NewClient(valkeyOptions)
	if valkeyClientErr != nil {
		return nil, errors.Wrap(valkeyClientErr, "failed to create Valkey client")
	}

	return &ValkeyStore[K]{
		client: valkeyClient,
	}, nil
}

func (v *ValkeyStore[K]) Delete(key K) error {
	//TODO: Do
	log.Printf("valkey store Delete key %v\n", key)
	return nil
}

func (v *ValkeyStore[K]) Get(key K) (stores.Location, error) {
	//TODO: Do
	log.Printf("valkey store Get key %v\n", key)
	return stores.InvalidLocation, nil
}

func (v *ValkeyStore[K]) Set(key K, value stores.Location) error {
	//TODO: Do
	log.Printf("valkey store Set key %v\n", key)
	return nil
}

func (v *ValkeyStore[K]) SetAndForget(key K, value stores.Location) {
	v.Set(key, value)
}
