package raft

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/raft"
	"github.com/ristryder/maydinhed/stores"
)

type RaftFsm[K stores.StoreKey] Node[K]

func (f *RaftFsm[K]) applyDelete(key K) error {
	return f.storeImpl.Delete(key)
}

func (f *RaftFsm[K]) applySet(key K, value stores.Location) error {
	return f.storeImpl.Set(key, value)
}

func (f *RaftFsm[K]) Apply(log *raft.Log) interface{} {
	var cmd stores.Command[K]
	if unmarshalErr := json.Unmarshal(log.Data, &cmd); unmarshalErr != nil {
		panic(fmt.Sprintf("failed to apply Raft log entry, error unmarshaling command: %s", unmarshalErr.Error()))
	}

	switch cmd.Operation {
	case "delete":
		return f.applyDelete(cmd.Key)
	case "set":
		return f.applySet(cmd.Key, cmd.Value)
	default:
		panic(fmt.Sprintf("failed to apply Raft log entry, unrecognized command: %s", cmd.Operation))
	}
}

func (f *RaftFsm[K]) Restore(readCloser io.ReadCloser) error {
	//TODO: Do
	return nil
}

func (f *RaftFsm[K]) Snapshot() (raft.FSMSnapshot, error) {
	//TODO: Do
	return nil, nil
}
