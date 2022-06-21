package services

import (
	"dfs/proto"
	"dfs/storageGateway/node"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type NodeService struct {
	logger       *zap.Logger
	mutex        *sync.Mutex
	nodes        map[uuid.UUID]*node.Node
	indexedNodes []*node.Node
	next         uint32
}

func NewNodeService(log *zap.Logger) *NodeService {
	return &NodeService{logger: log, mutex: &sync.Mutex{}, nodes: map[uuid.UUID]*node.Node{}, indexedNodes: []*node.Node{}}
}

func (sn *NodeService) addNode(node *node.Node) {
	sn.logger.Debug("New storage node registered", zap.String("NodeAddress", node.IpAddress),
		zap.Uint64("NodePort", node.Port))
	sn.mutex.Lock()

	if _, ok := sn.nodes[node.Uuid]; ok == false {
		sn.nodes[node.Uuid] = node
		sn.indexedNodes = append(sn.indexedNodes, node)
	}

	sn.mutex.Unlock()
}

func (sn *NodeService) deleteNode(nodeUuid uuid.UUID) {
	sn.logger.Debug("Storage node unregistered", zap.String("NodeUuid", nodeUuid.String()))
	sn.mutex.Lock()

	if _, ok := sn.nodes[nodeUuid]; ok {
		delete(sn.nodes, nodeUuid)
		var idx int

		for i := range sn.indexedNodes {
			if sn.indexedNodes[i].Uuid == nodeUuid {
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

func (sn *NodeService) GetNodes() *[]*node.Node {
	return &sn.indexedNodes
}

func (sn *NodeService) ProcessNodeMessage(message node.LifeCycleMessage) {
	switch message.Action {
	case node.Add:
		if len(sn.indexedNodes) == 0 {
			sn.addNode(&message.Node)
		} else {
			if sn.SyncNode(&message.Node) {
				sn.addNode(&message.Node)
			}
		}
		break
	case node.Delete:
		sn.deleteNode(message.Node.Uuid)
		break
	}
}

func (sn *NodeService) SyncNode(newNode *node.Node) bool {
	sn.logger.Debug("Syncing node", zap.String("IpAddress", newNode.IpAddress), zap.Uint64("Port", newNode.Port))
	rand.Seed(time.Now().UnixNano())
	idx := rand.Int() % len(sn.indexedNodes)
	n := sn.indexedNodes[idx]

	grpcMasterNodeClient := NewGrpcStorageClient(sn.logger)

	if err := grpcMasterNodeClient.Connect(fmt.Sprintf("%s:%d", n.IpAddress, n.GrpcPort)); err != nil {
		sn.logger.Error("Cannot connect to Grpc server", zap.Error(err))
	}

	storedFiles := grpcMasterNodeClient.GetStoredFiles()

	grpcMasterNodeClient.Disconnect()

	grpcNewNodeClient := NewGrpcStorageClient(sn.logger)

	if err := grpcNewNodeClient.Connect(fmt.Sprintf("%s:%d", newNode.IpAddress, newNode.GrpcPort)); err != nil {
		sn.logger.Error("Cannot connect to Grpc server", zap.Error(err))
	}

	isSync := grpcNewNodeClient.SyncStoredFiles(storedFiles)

	grpcNewNodeClient.Disconnect()

	return isSync
}

func (sn *NodeService) Next() *node.Node {
	n := atomic.AddUint32(&sn.next, 1)
	activeNodes := sn.indexedNodes

	sn.logger.Debug("Active nodes", zap.Int("ActiveNodesLen", len(activeNodes)))

	return (activeNodes)[(int(n)-1)%len(activeNodes)]
}

func (sn *NodeService) SyncHomeDirectory(masterNode *node.Node, homeDir *proto.HomeDir) {
	sn.mutex.Lock()
	sn.mutex.Unlock()

	for _, n := range sn.indexedNodes {
		if n != masterNode {
			grpcClient := NewGrpcStorageClient(sn.logger)

			if err := grpcClient.Connect(fmt.Sprintf("%s:%d", n.IpAddress, n.GrpcPort)); err != nil {
				sn.logger.Error("Cannot connect to Grpc server", zap.Error(err))
			}

			if grpcClient.CreateHomeDirectory(homeDir) == false {
				sn.logger.Error("Cannot create home directory on node",
					zap.String("NodeAddress", n.IpAddress), zap.Uint64("NodePort", n.Port))
			}

			grpcClient.Disconnect()
		}
	}
}

func (sn *NodeService) SyncSaveFile(masterNode *node.Node, req *proto.SaveFileRequest) {
	for _, n := range sn.nodes {
		if n != masterNode {
			grpcClient := NewGrpcStorageClient(sn.logger)

			if err := grpcClient.Connect(fmt.Sprintf("%s:%d", n.IpAddress, n.GrpcPort)); err != nil {
				sn.logger.Error("Cannot connect to Grpc server", zap.Error(err))
			}

			if grpcClient.SaveFileOnDisk(req) == false {
				sn.logger.Error("Cannot save file on node",
					zap.String("NodeAddress", n.IpAddress), zap.Uint64("NodePort", n.Port))
			}

			grpcClient.Disconnect()
		}
	}
}

func (sn *NodeService) SyncDeleteFile(masterNode *node.Node, req *proto.DeleteFileRequest) {
	for _, n := range sn.nodes {
		if n != masterNode {
			grpcClient := NewGrpcStorageClient(sn.logger)

			if err := grpcClient.Connect(fmt.Sprintf("%s:%d", n.IpAddress, n.GrpcPort)); err != nil {
				sn.logger.Error("Cannot connect to Grpc server", zap.Error(err))
			}

			if grpcClient.DeleteFileFromDisk(req) == false {
				sn.logger.Error("Cannot delete file from node",
					zap.String("NodeAddress", n.IpAddress), zap.Uint64("NodePort", n.Port))
			}

			grpcClient.Disconnect()
		}
	}
}
