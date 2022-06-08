package node

import (
	"go.uber.org/zap"
	"sync"
)

type ActiveNodes struct {
	logger       *zap.Logger
	mutex        *sync.Mutex
	nodes        map[string]bool
	indexedNodes []string
}

func NewNodes(log *zap.Logger) *ActiveNodes {
	return &ActiveNodes{logger: log, mutex: &sync.Mutex{}, nodes: map[string]bool{}, indexedNodes: []string{}}
}

func (sn *ActiveNodes) addNode(nodeAddress string) {
	sn.logger.Debug("New storage node registered", zap.String("NodeAddress", nodeAddress))
	sn.mutex.Lock()

	if sn.nodes[nodeAddress] == false {
		sn.nodes[nodeAddress] = true
		sn.indexedNodes = append(sn.indexedNodes, nodeAddress)
	}

	sn.mutex.Unlock()
}

func (sn *ActiveNodes) deleteNode(nodeAddress string) {
	sn.logger.Debug("Storage node unregistered", zap.String("NodeAddress", nodeAddress))
	sn.mutex.Lock()

	if _, ok := sn.nodes[nodeAddress]; ok {
		delete(sn.nodes, nodeAddress)
		var idx int

		for i := range sn.indexedNodes {
			if sn.indexedNodes[i] == nodeAddress {
				idx = i
				break
			}
		}

		sn.indexedNodes[idx] = sn.indexedNodes[len(sn.indexedNodes)-1]
		sn.indexedNodes = sn.indexedNodes[:len(sn.indexedNodes)-1]
	}

	sn.logger.Debug("Nodes after removing", zap.Any("Nodes", sn.indexedNodes))

	sn.mutex.Unlock()
}

func (sn *ActiveNodes) GetNodes() *[]string {
	return &sn.indexedNodes
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
