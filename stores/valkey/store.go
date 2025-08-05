package valkey

import (
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
	return nil
}

func (v *ValkeyStore[K]) Get(key K) (stores.Location, error) {
	//TODO: Do
	return stores.InvalidLocation, nil
}

func (v *ValkeyStore[K]) Set(key K, value stores.Location) error {
	//TODO: Do
	return nil
}
