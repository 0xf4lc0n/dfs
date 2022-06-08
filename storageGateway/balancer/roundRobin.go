package balancer

import (
	"dfs/storageGateway/node"
	"sync/atomic"
)

type RoundRobin struct {
	nodes *node.ActiveNodes
	next  uint32
}

func New(nodes *node.ActiveNodes) *RoundRobin {
	return &RoundRobin{nodes: nodes}
}

func (r *RoundRobin) Next() string {
	n := atomic.AddUint32(&r.next, 1)
	activeNodes := r.nodes.GetNodes()
	return (*activeNodes)[(int(n)-1)%len(*activeNodes)]
}
