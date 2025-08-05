# Start maydinhed
This example can be used to bootstrap the cluster with the leader, as well as add followers.

Note that the port numbers and node names used here are just examples.

Also note that the follower must use the "raftAddress" value from the leader for the "leaderAddress" parameter.

### Start the cluster
```
$ go run main.go -httpAddress 127.0.0.1:11000 -raftAddress 127.0.0.1:12000 -nodeId leader
```

### Start the follower
```
$ go run main.go -httpAddress 127.0.0.1:11001 -raftAddress 127.0.0.1:12001 -nodeId follower -leaderAddress 127.0.0.1:11000
```
