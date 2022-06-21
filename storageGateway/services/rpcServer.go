package services

import (
	"dfs/proto"
	"dfs/storageGateway/dtos"
	"dfs/storageGateway/node"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"strconv"
)

type RpcServer struct {
	logger     *zap.Logger
	connection *amqp.Connection
	nodes      *NodeService
	grpcClient *GrpcStorageClient
}

func NewRpcServer(logger *zap.Logger, nodesStorage *NodeService, grpc *GrpcStorageClient) *RpcServer {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(logger, err, "Failed to connect to RabbitMQ")
	return &RpcServer{logger: logger, connection: conn, nodes: nodesStorage, grpcClient: grpc}
}

func (rpc *RpcServer) RegisterNodeMessages() {
	ch, _, messages := rpc.createQueue("rpc_gateway_node_messages")
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		// Listen and process  each RPC request
		for msg := range messages {
			rpc.logger.Debug("[<--]", zap.ByteString("LifeCycleMessage", msg.Body))

			var nodeMessage node.LifeCycleMessage

			err := json.Unmarshal(msg.Body, &nodeMessage)

			rpc.nodes.ProcessNodeMessage(nodeMessage)

			if err != nil {
				rpc.failOnError(err, "Failed to deserialize nodeMessage")
			}

			// Send manual acknowledgement
			msg.Ack(false)
		}
	}()

	rpc.logger.Info("[*] Awaiting 'NodeLifecycle' messages")
	<-forever
}

func (rpc *RpcServer) RegisterCreateHomeDirectory() {
	ch, _, messages := rpc.createQueue("rpc_storage_create_home_dir_queue")
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		// Listen and process  each RPC request
		for msg := range messages {
			directoryName := string(msg.Body)

			rpc.logger.Debug("[<--]", zap.String("HomeDirectory", directoryName))

			pickedNode := rpc.nodes.Next()
			grpcClient := NewGrpcStorageClient(rpc.logger)

			if err := grpcClient.Connect(fmt.Sprintf("%s:%d", pickedNode.IpAddress, pickedNode.GrpcPort)); err != nil {
				rpc.logger.Error("Cannot connect to Grpc server", zap.Error(err))
			}

			homeDir := &proto.HomeDir{Name: directoryName}

			isCreated := grpcClient.CreateHomeDirectory(homeDir)

			if isCreated {
				go rpc.nodes.SyncHomeDirectory(pickedNode, homeDir)
			}

			rpc.logger.Debug("[-->]", zap.Bool("IsHomeDirectoryCreated", isCreated))

			rpc.publishAndAck(ch, msg, []byte(strconv.FormatBool(isCreated)), "text/plain")

			grpcClient.Disconnect()
		}
	}()

	rpc.logger.Info("[*] Awaiting 'CreateHomeDirectory' RPC requests")
	<-forever
}

func (rpc *RpcServer) RegisterGetOwnedFile() {
	ch, _, messages := rpc.createQueue("rpc_storage_get_owned_file_queue")
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		// Listen and process  each RPC request
		for msg := range messages {
			rpc.logger.Debug("[<--]", zap.String("SerializedOwnedFileDto", string(msg.Body)))

			var ownedFileDto dtos.OwnedFileDto

			err := json.Unmarshal(msg.Body, &ownedFileDto)

			rpc.failOnError(err, "Cannot deserialize object to shareDto")

			pickedNode := rpc.nodes.Next()
			grpcClient := NewGrpcStorageClient(rpc.logger)

			if err := grpcClient.Connect(fmt.Sprintf("%s:%d", pickedNode.IpAddress, pickedNode.GrpcPort)); err != nil {
				rpc.logger.Error("Cannot connect to Grpc server", zap.Error(err))
			}

			fileEntry := grpcClient.GetOwnedFile(&proto.OwnedFileRequest{FileId: ownedFileDto.FileId,
				OwnerId: ownedFileDto.OwnerId})

			serializedFileEntry, err := json.Marshal(fileEntry)

			rpc.failOnError(err, "Cannot serialize file entry")

			rpc.logger.Debug("[-->]", zap.String("FileEntry", string(serializedFileEntry)))

			rpc.publishAndAck(ch, msg, serializedFileEntry, "application/json")

			grpcClient.Disconnect()
		}
	}()

	rpc.logger.Info("[*] Awaiting 'GetOwnedFile' RPC requests")
	<-forever
}

func (rpc *RpcServer) RegisterGetFileById() {
	ch, _, messages := rpc.createQueue("rpc_storage_get_file_by_id_queue")
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		// Listen and process each RPC request
		for msg := range messages {
			rpc.logger.Debug("[<--]", zap.String("SerializedFileId", string(msg.Body)))

			var fileId uint64

			err := json.Unmarshal(msg.Body, &fileId)

			rpc.failOnError(err, "Cannot deserialize object to uint64")

			pickedNode := rpc.nodes.Next()
			grpcClient := NewGrpcStorageClient(rpc.logger)

			if err := grpcClient.Connect(fmt.Sprintf("%s:%d", pickedNode.IpAddress, pickedNode.GrpcPort)); err != nil {
				rpc.logger.Error("Cannot connect to Grpc server", zap.Error(err))
			}

			fileEntry := grpcClient.GetFileById(&proto.GetFileByIdRequest{FileId: fileId})

			serializedFileEntry, err := json.Marshal(fileEntry)

			rpc.failOnError(err, "Cannot serialize file entry")

			rpc.logger.Debug("[-->]", zap.String("FileEntry", string(serializedFileEntry)))

			rpc.publishAndAck(ch, msg, serializedFileEntry, "application/json")

			grpcClient.Disconnect()
		}
	}()

	rpc.logger.Info("[*] Awaiting 'GetFileById' RPC requests")
	<-forever
}

func (rpc *RpcServer) RegisterGetFileByUniqueName() {
	ch, _, messages := rpc.createQueue("rpc_storage_get_file_by_unique_name_queue")
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		// Listen and process each RPC request
		for msg := range messages {
			rpc.logger.Debug("[<--]", zap.ByteString("SerializedFileUniqueName", msg.Body))

			var fileUniqueName string

			err := json.Unmarshal(msg.Body, &fileUniqueName)

			rpc.failOnError(err, "Cannot deserialize object to shareDto")

			pickedNode := rpc.nodes.Next()
			grpcClient := NewGrpcStorageClient(rpc.logger)

			if err := grpcClient.Connect(fmt.Sprintf("%s:%d", pickedNode.IpAddress, pickedNode.GrpcPort)); err != nil {
				rpc.logger.Error("Cannot connect to Grpc server", zap.Error(err))
			}

			fileEntry := grpcClient.GetFileByUniqueName(&proto.FileUniqueName{Name: fileUniqueName})

			serializedFileEntry, err := json.Marshal(fileEntry)

			rpc.failOnError(err, "Cannot serialize file entry")

			rpc.logger.Debug("[-->]", zap.String("FileEntry", string(serializedFileEntry)))

			rpc.publishAndAck(ch, msg, serializedFileEntry, "application/json")

			grpcClient.Disconnect()
		}
	}()

	rpc.logger.Info("[*] Awaiting 'GetFileByUniqueName' RPC requests")
	<-forever
}

func (rpc *RpcServer) RegisterSaveFileOnDisk() {
	ch, _, messages := rpc.createQueue("rpc_storage_save_file")
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		// Listen and process each RPC request
		for msg := range messages {
			rpc.logger.Debug("[<--]", zap.String("SerializedSaveFileDto", string(msg.Body)))

			var saveFileDto dtos.SaveFileDto
			err := json.Unmarshal(msg.Body, &saveFileDto)

			rpc.failOnError(err, "Cannot deserialize object to SaveFileDto")

			pickedNode := rpc.nodes.Next()
			grpcClient := NewGrpcStorageClient(rpc.logger)

			if err := grpcClient.Connect(fmt.Sprintf("%s:%d", pickedNode.IpAddress, pickedNode.GrpcPort)); err != nil {
				rpc.logger.Error("Cannot connect to Grpc server", zap.Error(err))
			}

			req := &proto.SaveFileRequest{SavePath: saveFileDto.SavePath, EncryptionKey: saveFileDto.EncryptionKey,
				Content: saveFileDto.Content}

			isSaved := grpcClient.SaveFileOnDisk(req)

			if isSaved {
				go rpc.nodes.SyncSaveFile(pickedNode, req)
			}

			rpc.logger.Debug("[-->]", zap.Bool("IsFileSaved", isSaved))

			rpc.publishAndAck(ch, msg, []byte(strconv.FormatBool(isSaved)), "")

			grpcClient.Disconnect()
		}
	}()

	rpc.logger.Info("[*] Awaiting 'SaveFileOnDisk' RPC requests")
	<-forever
}

func (rpc *RpcServer) RegisterDeleteFileFromDisk() {
	ch, _, messages := rpc.createQueue("rpc_storage_delete_file")
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		// Listen and process each RPC request
		for msg := range messages {
			rpc.logger.Debug("[<--]", zap.String("SerializedDeleteFileDto", string(msg.Body)))

			var deleteFileDto *dtos.DeleteFileDto = nil
			err := json.Unmarshal(msg.Body, &deleteFileDto)

			rpc.failOnError(err, "Cannot deserialize object to DeleteFileDto")

			pickedNode := rpc.nodes.Next()
			grpcClient := NewGrpcStorageClient(rpc.logger)

			if err := grpcClient.Connect(fmt.Sprintf("%s:%d", pickedNode.IpAddress, pickedNode.GrpcPort)); err != nil {
				rpc.logger.Error("Cannot connect to Grpc server", zap.Error(err))
			}

			req := &proto.DeleteFileRequest{FilePath: deleteFileDto.FilePath}

			isDeleted := grpcClient.DeleteFileFromDisk(req)

			if isDeleted {
				go rpc.nodes.SyncDeleteFile(pickedNode, req)
			}

			rpc.logger.Debug("[-->]", zap.Bool("IsFileDeleted", isDeleted))

			rpc.publishAndAck(ch, msg, []byte(strconv.FormatBool(isDeleted)), "")

			grpcClient.Disconnect()
		}
	}()

	rpc.logger.Info("[*] Awaiting 'DeleteFileFromDisk' RPC requests")
	<-forever
}

func (rpc *RpcServer) RegisterGetFileContentFromDisk() {
	ch, _, messages := rpc.createQueue("rpc_storage_get_file_content")
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		// Listen and process each RPC request
		for msg := range messages {
			rpc.logger.Debug("[<--]", zap.ByteString("SerializedReadFileDto", msg.Body))

			var readFileDto *dtos.ReadFileDto = nil

			err := json.Unmarshal(msg.Body, &readFileDto)

			rpc.failOnError(err, "Cannot deserialize object to ReadFileDto")

			pickedNode := rpc.nodes.Next()
			grpcClient := NewGrpcStorageClient(rpc.logger)

			if err := grpcClient.Connect(fmt.Sprintf("%s:%d", pickedNode.IpAddress, pickedNode.GrpcPort)); err != nil {
				rpc.logger.Error("Cannot connect to Grpc server", zap.Error(err))
			}

			fileContent := grpcClient.GetFileContentFromDisk(&proto.ReadFileRequest{ReadPath: readFileDto.ReadPath,
				DecryptionKey: readFileDto.DecryptionKey})

			rpc.logger.Debug("[-->]", zap.ByteString("FileContent", fileContent))

			rpc.publishAndAck(ch, msg, fileContent, "text/plain")

			grpcClient.Disconnect()
		}
	}()

	rpc.logger.Info("[*] Awaiting 'ReadFileFromDisk' RPC requests")
	<-forever
}

func (rpc *RpcServer) Close() {
	if err := rpc.connection.Close(); err != nil {
		rpc.logger.Error("Cannot close RabbitMq connection", zap.Error(err))
	}
}

func (rpc *RpcServer) createQueue(queueName string) (*amqp.Channel, amqp.Queue, <-chan amqp.Delivery) {
	ch, err := rpc.connection.Channel()
	rpc.failOnError(err, "Failed to open a channel")

	// Storage service RPC queue aka RPC Server
	rpcQueue, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	rpc.failOnError(err, "Failed to declare a queue")

	// Get server messages channel
	messages, err := ch.Consume(
		rpcQueue.Name, // queue
		"",            // consumer
		false,         // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)

	rpc.failOnError(err, "Failed to register a consumer")

	return ch, rpcQueue, messages
}

func (rpc *RpcServer) publishAndAck(ch *amqp.Channel, msg amqp.Delivery, data []byte, contentType string) {
	// Send message to client callback queue
	err := ch.Publish(
		"",          // exchange
		msg.ReplyTo, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:   contentType,
			CorrelationId: msg.CorrelationId,
			Body:          data,
		})

	rpc.failOnError(err, "Failed to publish a message")

	// Send manual acknowledgement
	msg.Ack(false)
}

func (rpc *RpcServer) failOnError(err error, msg string) {
	failOnError(rpc.logger, err, msg)
}

func failOnError(logger *zap.Logger, err error, msg string) {
	if err != nil {
		logger.Fatal("RPC", zap.String("Msg", msg), zap.Error(err))
	}
}
