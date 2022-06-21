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
