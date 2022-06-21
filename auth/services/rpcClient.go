package services

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RpcClient struct {
	logger     *zap.Logger
	connection *amqp.Connection
}

func NewRpcClient(logger *zap.Logger) *RpcClient {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(logger, err, "Failed to connect to RabbitMQ")
	return &RpcClient{logger: logger, connection: conn}
}

func (rpc *RpcClient) CreateHomeDirectory(directoryName string) bool {
	ch, callbackQueue, messages := rpc.createCallbackQueue()
	defer ch.Close()

	// Generate correlation ID for RPC
	corrId := uuid.New().String()

	rpc.logger.Debug("[-->]", zap.String("HomeDirectory", directoryName))
	// Invoke RPC
	err := ch.Publish(
		"",
		"rpc_storage_create_home_dir_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       callbackQueue.Name,
			Body:          []byte(directoryName),
		},
	)

	rpc.logger.Info("Published", zap.String("DirectoryName", directoryName))

	if err != nil {
		rpc.logger.Error("Failed to publish a message", zap.Error(err))
		return false
	}

	// Listen for RPC responses
	for msg := range messages {
		if corrId == msg.CorrelationId {
			response, err := strconv.ParseBool(string(msg.Body))

			if err != nil {
				rpc.logger.Debug("Failed to convert body to bool", zap.Error(err))
				return false
			} else {
				rpc.logger.Debug("[<--]", zap.Bool("IsHomeDirectoryCreated", response))
				return response
			}
		}
	}

	return false
}

func (rpc *RpcClient) Close() {
	rpc.connection.Close()
}

func (rpc *RpcClient) createCallbackQueue() (*amqp.Channel, amqp.Queue, <-chan amqp.Delivery) {
	ch, err := rpc.connection.Channel()

	rpc.failOnError(err, "Failed to open a channel")

	// Anonymous exclusive callback queue
	callbackQueue, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)

	rpc.failOnError(err, "Failed to declare a callback queue")

	// Get callback messages channel
	messages, err := ch.Consume(
		callbackQueue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	rpc.failOnError(err, "Failed to register a consumer")

	return ch, callbackQueue, messages
}

func (rpc *RpcClient) failOnError(err error, msg string) {
	if err != nil {
		rpc.logger.Fatal("RPC", zap.String("Msg", msg), zap.Error(err))
	}
}
