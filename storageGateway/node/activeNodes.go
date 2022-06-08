package node

import (
	"go.uber.org/zap"
	"sync"
)

type ActiveNodes struct {
	logger *zap.Logger
	mutex  *sync.Mutex
	nodes  map[string]bool
}

func NewNodes(log *zap.Logger) *ActiveNodes {
	return &ActiveNodes{logger: log, mutex: &sync.Mutex{}, nodes: map[string]bool{}}
}

func (sn *ActiveNodes) addNode(nodeAddress string) {
	sn.logger.Debug("New storage node registered", zap.String("NodeAddress", nodeAddress))
	sn.mutex.Lock()
	sn.nodes[nodeAddress] = true
	sn.mutex.Unlock()
}

func (sn *ActiveNodes) deleteNode(nodeAddress string) {
	sn.logger.Debug("Storage node unregistered", zap.String("NodeAddress", nodeAddress))
	sn.mutex.Lock()
	delete(sn.nodes, nodeAddress)
	sn.mutex.Unlock()
}

func (sn *ActiveNodes) GetNodes() map[string]bool {
	return sn.nodes
}

func (sn *ActiveNodes) ProcessNodeMessage(message Message) {
	switch message.Action {
	case Add:
		sn.addNode(message.NodeAddress)
		break
	case Delete:
		sn.deleteNode(message.NodeAddress)
		break
	}
}
