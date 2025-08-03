package raft

import (
	"io"

	"github.com/hashicorp/raft"
)

type raftFsm Node

func (f *raftFsm) Apply(log *raft.Log) interface{} {
	//TODO: Do
	return nil
}

func (f *raftFsm) Restore(readCloser io.ReadCloser) error {
	//TODO: Do
	return nil
}

func (f *raftFsm) Snapshot() (raft.FSMSnapshot, error) {
	//TODO: Do
	return nil, nil
}
