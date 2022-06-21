package node

import "github.com/google/uuid"

type RegisterNodeMessage struct {
	Uuid      uuid.UUID
	IpAddress string
	Port      uint64
	GrpcPort  uint64
}

type DeregisterNodeMessage struct {
	Uuid uuid.UUID
}
