package valkey

import (
	"log"

	"github.com/cockroachdb/errors"
	"github.com/ristryder/maydinhed/mapping"
	"github.com/valkey-io/valkey-go"
)

type ValkeyStore[K mapping.ClusteredMarkerKey] struct {
	client valkey.Client
}

func New[K mapping.ClusteredMarkerKey](valkeyOptions valkey.ClientOption) (*ValkeyStore[K], error) {
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

func (v *ValkeyStore[K]) Get(key K) (mapping.Location, error) {
	//TODO: Do
	log.Printf("valkey store Get key %v\n", key)
	return mapping.InvalidLocation, nil
}

func (v *ValkeyStore[K]) Set(key K, value mapping.Location) error {
	//TODO: Do
	log.Printf("valkey store Set key %v\n", key)
	return nil
}

func (v *ValkeyStore[K]) SetAndForget(key K, value mapping.Location) {
	v.Set(key, value)
}
