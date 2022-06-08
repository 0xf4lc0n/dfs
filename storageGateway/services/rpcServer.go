package services

import (
	"dfs/storageGateway/node"
	"encoding/json"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RpcServer struct {
	logger     *zap.Logger
	connection *amqp.Connection
	nodes      *node.ActiveNodes
}

func NewRpcServer(logger *zap.Logger, nodesStorage *node.ActiveNodes) *RpcServer {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(logger, err, "Failed to connect to RabbitMQ")
	return &RpcServer{logger: logger, connection: conn, nodes: nodesStorage}
}

func (rpc *RpcServer) RegisterNodeMessages() {
	ch, _, messages := rpc.createQueue("rpc_gateway_node_messages")
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		// Listen and process  each RPC request
		for msg := range messages {
			rpc.logger.Debug("[<--]", zap.ByteString("NodeMessage", msg.Body))

			var nodeMessage node.Message

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

func (rpc *RpcServer) failOnError(err error, msg string) {
	failOnError(rpc.logger, err, msg)
}

func failOnError(logger *zap.Logger, err error, msg string) {
	if err != nil {
		logger.Fatal("RPC", zap.String("Msg", msg), zap.Error(err))
	}
}
