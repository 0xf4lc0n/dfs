package node

import "github.com/google/uuid"

type ActionType uint

const (
	Add    ActionType = iota
	Delete ActionType = iota
)

type Node struct {
	Uuid      uuid.UUID
	IpAddress string
	Port      uint64
	GrpcPort  uint64
}

type LifeCycleMessage struct {
	Node   Node
	Action ActionType
}

func CreateRegisterNodeMessage(node *Node) *LifeCycleMessage {
	return &LifeCycleMessage{Node: *node, Action: Add}
}

func CreateDeregisterNodeMessage(node *Node) *LifeCycleMessage {
	return &LifeCycleMessage{Node: *node, Action: Delete}
}
