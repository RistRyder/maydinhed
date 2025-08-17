package raft

type NodeConfiguration struct {
	EnableEgress           bool
	EnableIngress          bool
	ValidateLocationUpdate bool
}

func NewNodeConfiguration(enableEgress, enableIngress, validateLocationUpdate bool) *NodeConfiguration {
	return &NodeConfiguration{
		EnableEgress:           enableEgress,
		EnableIngress:          enableIngress,
		ValidateLocationUpdate: validateLocationUpdate,
	}
}
