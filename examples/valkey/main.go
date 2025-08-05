package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/cockroachdb/errors"
	"github.com/ristryder/maydinhed/raft"
	"github.com/ristryder/maydinhed/stores"
	mvalkey "github.com/ristryder/maydinhed/stores/valkey"
	"github.com/valkey-io/valkey-go"
)

func joinCluster[K stores.StoreKey](leaderAddress string, node *raft.Node[K]) error {
	requestBytes, marshalErr := json.Marshal(map[string]string{"address": node.RaftAddress, "id": node.Id})
	if marshalErr != nil {
		return errors.Wrap(marshalErr, "failed to serialize join request")
	}

	response, responseErr := http.Post(fmt.Sprintf("http://%s/join", leaderAddress), "application/json", bytes.NewReader(requestBytes))
	if responseErr != nil {
		return errors.Wrap(responseErr, "received error response from join cluster request")
	}

	defer response.Body.Close()

	return nil
}

func main() {
	var httpAddress string
	var leaderAddress string
	var nodeId string
	var raftAddress string
	var raftDirectory string
	flag.StringVar(&httpAddress, "httpAddress", "", "HTTP listen address")
	flag.StringVar(&leaderAddress, "leaderAddress", "", "Address of the leader to join")
	flag.StringVar(&nodeId, "nodeId", "", "Unique id of this node")
	flag.StringVar(&raftAddress, "raftAddress", "", "Raft listen address")
	flag.StringVar(&raftDirectory, "raftDirectory", "", "Directory to store Raft data")
	flag.Parse()

	valkeyStore, valkeyStoreErr := mvalkey.New[string](valkey.ClientOption{
		EnableReplicaAZInfo: true,
		InitAddress:         []string{"address.example.com:6379"},
	})
	if valkeyStoreErr != nil {
		log.Fatal(valkeyStoreErr)
	}
	node := raft.NewNode(nodeId, true, raftAddress, raftDirectory, valkeyStore)

	isLeader := leaderAddress == ""

	if isLeader {
		log.Println("starting leader...")
	} else {
		log.Println("starting follower...")
	}

	openErr := node.Open(isLeader, nodeId)
	if openErr != nil {
		log.Fatal(openErr)
	}

	listener := raft.NewHttpListener(httpAddress, node)
	if listenerStartErr := listener.Start(); listenerStartErr != nil {
		log.Fatal(listenerStartErr)
	}

	defer listener.Close()

	if !isLeader {
		log.Println("joining cluster...")

		joinCluster(leaderAddress, node)
	}

	log.Println("waiting...")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("finished")
}
