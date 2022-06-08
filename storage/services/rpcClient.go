package services

import (
	"dfs/storage/dtos"
	"dfs/storage/node"
	"encoding/json"
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

func (rpc *RpcClient) GetUserDataByJwt(jwt string) *dtos.User {
	ch, callbackQueue, messages := rpc.createCallbackQueue()
	defer ch.Close()

	// Generate correlation ID for RPC
	corrId := uuid.New().String()

	rpc.logger.Debug("[-->]", zap.String("Jwt", jwt))

	// Invoke RPC
	err := ch.Publish(
		"",
		"rpc_auth_get_user_data_by_jwt_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       callbackQueue.Name,
			Body:          []byte(jwt),
		},
	)

	if err != nil {
		rpc.logger.Error("Failed to publish a message", zap.Error(err))
		return nil
	}
	// Listen for RPC responses
	for msg := range messages {
		if corrId == msg.CorrelationId {
			rpc.logger.Debug("[<--]", zap.String("UserData", string(msg.Body)))

			var userDto *dtos.User = nil
			err := json.Unmarshal(msg.Body, &userDto)

			if err != nil {
				rpc.logger.Error("Cannot deserialize data to UserDto", zap.Error(err))
				return nil
			}

			return userDto
		}
	}

	return nil
}

func (rpc *RpcClient) SendNodeMessage(node *node.Message) {
	ch, err := rpc.connection.Channel()
	rpc.failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rpc_gateway_node_messages", // name
		false,                       // durable
		false,                       // delete when unused
		false,                       // exclusive
		false,                       // no-wait
		nil,                         // arguments
	)

	rpc.failOnError(err, "Failed to declare a queue")

	serializedNodeMessage, err := json.Marshal(node)

	if err != nil {
		rpc.failOnError(err, "Cannot serialize ActionType dto")
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        serializedNodeMessage,
		})

	rpc.failOnError(err, "Failed to publish a message")

	rpc.logger.Debug("[-->]", zap.ByteString("SerializedNodeDto", serializedNodeMessage))
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
