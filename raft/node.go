package raft

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"github.com/ristryder/maydinhed/messaging"
	"github.com/ristryder/maydinhed/stores"
)

const (
	raftTimeout         = 10 * time.Second
	retainSnapshotCount = 2
)

type Node[K stores.StoreKey] struct {
	Id          string
	RaftAddress string

	inMemory              bool
	isLeader              bool
	messengerImpl         messaging.Messenger[K]
	raftDirectory         string
	raftNode              *raft.Raft
	raftObserver          *raft.Observer
	raftObserverCtx       context.Context
	raftObserverCtxCancel context.CancelFunc
	storeImpl             stores.Store[K]
}

type StoreNode[K stores.StoreKey] interface {
	stores.Store[K]
	AddNode(address, id string) error
	Shutdown() error
}

func (n *Node[K]) transitionToFollower() {
	n.isLeader = false

	if messengerStopErr := n.messengerImpl.StopListening(); messengerStopErr != nil {
		log.Println("raft state changed, we are currently not leader but failed to stop messaging consumer")
	} else {
		log.Println("raft state changed, we are currently not leader, messaging consumer stopped")
	}
}

func (n *Node[K]) transitionToLeader() {
	n.isLeader = true

	if messengerStartErr := n.messengerImpl.StartListening(); messengerStartErr != nil {
		log.Println("raft state changed, we are currently leader but failed to start messaging consumer")
	} else {
		log.Println("raft state changed, we are currently leader, messaging consumer started")
	}
}

func (n *Node[K]) watchRaftState() {
	observationChannel := make(chan raft.Observation)

	go func() {
		for {
			select {
			case <-n.raftObserverCtx.Done():
				log.Println("stopping raft observer")
				return
			case <-observationChannel:
				if n.raftNode.State() == raft.Leader {
					n.transitionToLeader()
				} else {
					n.transitionToFollower()
				}
			case <-time.After(time.Second):
				if n.raftNode.State() == raft.Leader && !n.isLeader {
					n.transitionToLeader()
				} else if n.raftNode.State() != raft.Leader && n.isLeader {
					n.transitionToFollower()
				}
			}
		}
	}()

	n.raftObserver = raft.NewObserver(observationChannel, true, func(o *raft.Observation) bool {
		switch o.Data.(type) {
		case raft.LeaderObservation:
			return true
		default:
			return false
		}
	})

	n.raftObserverCtx, n.raftObserverCtxCancel = context.WithCancel(context.Background())

	n.raftNode.RegisterObserver(n.raftObserver)
}

func NewNode[K stores.StoreKey](id string, inMemory bool, messengerImpl messaging.Messenger[K], raftAddress, raftDirectory string, storeImpl stores.Store[K]) *Node[K] {
	return &Node[K]{
		Id:            id,
		inMemory:      inMemory,
		messengerImpl: messengerImpl,
		RaftAddress:   raftAddress,
		raftDirectory: raftDirectory,
		storeImpl:     storeImpl,
	}
}

func (n *Node[K]) AddNode(address, nodeId string) error {
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

func (n *Node[K]) Delete(key K) error {
	deleteCommand := &stores.Command[K]{
		Key:       key,
		Operation: "delete",
	}
	deleteCommandBytes, marshalErr := json.Marshal(deleteCommand)
	if marshalErr != nil {
		return errors.Wrap(marshalErr, "failed to marshal delete command")
	}

	return n.raftNode.Apply(deleteCommandBytes, raftTimeout).Error()
}

func (n *Node[K]) Get(key K) (stores.Location, error) {
	return n.storeImpl.Get(key)
}

func (n *Node[K]) Open(bootstrap bool, localNodeId string) error {
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

	raftNode, raftErr := raft.NewRaft(raftConfig, (*RaftFsm[K])(n), logStore, stableStore, snapshots, transport)
	if raftErr != nil {
		return errors.Wrap(raftErr, "failed to initialize Raft node")
	}
	n.raftNode = raftNode

	n.watchRaftState()

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

func (n *Node[K]) Set(key K, value stores.Location) error {
	setCommand := &stores.Command[K]{
		Key:       key,
		Operation: "set",
		Value:     value,
	}
	setCommandBytes, marshalErr := json.Marshal(setCommand)
	if marshalErr != nil {
		return errors.Wrap(marshalErr, "failed to marshal set command")
	}

	if applyErr := n.raftNode.Apply(setCommandBytes, raftTimeout).Error(); applyErr != nil {
		return errors.Wrap(applyErr, "failed to apply set command")
	}

	if messengerErr := n.messengerImpl.SendLocationUpdate(key, value); messengerErr != nil {
		return errors.Wrap(messengerErr, "failed to send location update message")
	}

	return nil
}

func (n *Node[K]) SetAndForget(key K, value stores.Location) {
	go func() {
		setCommand := &stores.Command[K]{
			Key:       key,
			Operation: "set",
			Value:     value,
		}
		setCommandBytes, marshalErr := json.Marshal(setCommand)
		if marshalErr != nil {
			log.Println("failed to marshal set command: ", marshalErr)

			return
		}

		if applyErr := n.raftNode.Apply(setCommandBytes, raftTimeout).Error(); applyErr != nil {
			log.Println("failed to apply set command: ", applyErr)
		}
	}()

	go func() {
		if messengerErr := n.messengerImpl.SendLocationUpdate(key, value); messengerErr != nil {
			log.Println("failed to send location update message: ", messengerErr)
		}
	}()
}

func (n *Node[K]) Shutdown() error {
	n.raftObserverCtxCancel()

	n.raftNode.DeregisterObserver(n.raftObserver)

	return n.raftNode.Shutdown().Error()
}
