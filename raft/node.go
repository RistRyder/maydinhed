package raft

import (
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

const (
	raftTimeout         = 10 * time.Second
	retainSnapshotCount = 2
)

type Node struct {
	Id          string
	RaftAddress string

	inMemory      bool
	raftDirectory string
	raftNode      *raft.Raft
}

func NewNode(id string, inMemory bool, raftAddress, raftDirectory string) *Node {
	return &Node{
		Id:            id,
		inMemory:      inMemory,
		RaftAddress:   raftAddress,
		raftDirectory: raftDirectory,
	}
}

func (n *Node) AddNode(address, nodeId string) error {
	raftConfig := n.raftNode.GetConfiguration()
	if raftConfigErr := raftConfig.Error(); raftConfigErr != nil {
		return errors.Wrap(raftConfigErr, "failed to retrieve Raft configuration")
	}

	for _, server := range raftConfig.Configuration().Servers {
		if server.ID == raft.ServerID(nodeId) || server.Address == raft.ServerAddress(address) {
			if server.ID == raft.ServerID(nodeId) && server.Address == raft.ServerAddress(address) {
				return nil
			}

			removeResult := n.raftNode.RemoveServer(server.ID, 0, 0)
			if removeResultErr := removeResult.Error(); removeResultErr != nil {
				return errors.Wrapf(removeResultErr, "failed to remove node %s at %s", nodeId, address)
			}
		}
	}

	addVoterResult := n.raftNode.AddVoter(raft.ServerID(nodeId), raft.ServerAddress(address), 0, raftTimeout)
	if addVoterResultErr := addVoterResult.Error(); addVoterResultErr != nil {
		return addVoterResultErr
	}

	return nil
}

func (n *Node) Delete(key StoreKey) error {
	//TODO: Do
	return nil
}

func (n *Node) Get(key StoreKey) (StoreValue, error) {
	//TODO: Do
	return nil, nil
}

func (n *Node) Open(bootstrap bool, localNodeId string) error {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(localNodeId)

	bindAddress, bindErr := net.ResolveTCPAddr("tcp", n.RaftAddress)
	if bindErr != nil {
		return errors.Wrap(bindErr, "failed to resolve Raft bind address")
	}
	transport, transportErr := raft.NewTCPTransport(n.RaftAddress, bindAddress, 3, 10*time.Second, os.Stderr)
	if transportErr != nil {
		return errors.Wrap(transportErr, "failed to create Raft transport")
	}
	snapshots, snapshotErr := raft.NewFileSnapshotStore(n.raftDirectory, retainSnapshotCount, os.Stderr)
	if snapshotErr != nil {
		return errors.Wrap(snapshotErr, "failed to create snapshot store")
	}

	var logStore raft.LogStore
	var stableStore raft.StableStore
	if n.inMemory {
		logStore = raft.NewInmemStore()
		stableStore = raft.NewInmemStore()
	} else {
		boltDb, boltDbErr := raftboltdb.New(raftboltdb.Options{
			Path: filepath.Join(n.raftDirectory, "raft.db"),
		})
		if boltDbErr != nil {
			return errors.Wrap(boltDbErr, "failed to create bbolt store")
		}

		logStore = boltDb
		stableStore = boltDb
	}

	raftNode, raftErr := raft.NewRaft(raftConfig, (*raftFsm)(n), logStore, stableStore, snapshots, transport)
	if raftErr != nil {
		return errors.Wrap(raftErr, "failed to initialize Raft node")
	}
	n.raftNode = raftNode

	if bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					Address: transport.LocalAddr(),
					ID:      raftConfig.LocalID,
				},
			},
		}

		bootstrapResult := raftNode.BootstrapCluster(configuration)
		if bootstrapResultErr := bootstrapResult.Error(); bootstrapResultErr != nil {
			return errors.Wrap(bootstrapResultErr, "failed to bootstrap cluster")
		}
	}

	return nil
}

func (n *Node) Set(key StoreKey, value StoreValue) error {
	//TODO: Do
	return nil
}
