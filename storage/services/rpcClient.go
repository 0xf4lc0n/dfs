package services

import (
	"dfs/storage/dtos"
	"encoding/json"
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

func (rpc *RpcClient) GetHomeDirectory(jwt string) string {
	ch, err := rpc.connection.Channel()

	if err != nil {
		rpc.logger.Error("Failed to open a channel", zap.Error(err))
		return ""
	}

	defer ch.Close()

	// Anonymous exclusive callback queue
	callbackQueue, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)

	if err != nil {
		rpc.logger.Error("Failed to declare a queue", zap.Error(err))
		return ""
	}

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

	if err != nil {
		rpc.logger.Error("Failed to register a consumer", zap.Error(err))
		return ""
	}

	// Generate correlation ID for RPC
	corrId := uuid.New().String()

	rpc.logger.Debug("[-->]", zap.String("Jwt", jwt))

	// Invoke RPC
	err = ch.Publish(
		"",
		"rpc_auth_get_user_home_dir_queue",
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
		return ""
	}

	// Listen for RPC responses
	for msg := range messages {
		if corrId == msg.CorrelationId {
			response := string(msg.Body)
			rpc.logger.Debug("[<--]", zap.String("HomeDirectory", response))
			return response
		}
	}

	return ""
}

func (rpc *RpcClient) GetUserDataByJwt(jwt string) *dtos.User {
	ch, err := rpc.connection.Channel()

	if err != nil {
		rpc.logger.Error("Failed to open a channel", zap.Error(err))
		return nil
	}

	defer ch.Close()

	// Anonymous exclusive callback queue
	callbackQueue, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)

	if err != nil {
		rpc.logger.Error("Failed to declare a queue", zap.Error(err))
		return nil
	}

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

	if err != nil {
		rpc.logger.Error("Failed to register a consumer", zap.Error(err))
		return nil
	}

	// Generate correlation ID for RPC
	corrId := uuid.New().String()

	rpc.logger.Debug("[-->]", zap.String("Jwt", jwt))

	// Invoke RPC
	err = ch.Publish(
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

func (rpc *RpcClient) IsAuthenticated(jwt string) bool {
	ch, err := rpc.connection.Channel()

	if err != nil {
		rpc.logger.Error("Failed to open a channel", zap.Error(err))
		return false
	}

	defer ch.Close()

	// Anonymous exclusive callback queue
	callbackQueue, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)

	if err != nil {
		rpc.logger.Error("Failed to declare a queue", zap.Error(err))
		return false
	}

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

	if err != nil {
		rpc.logger.Error("Failed to register a consumer", zap.Error(err))
		return false
	}

	// Generate correlation ID for RPC
	corrId := uuid.New().String()

	rpc.logger.Debug("[-->]", zap.String("Jwt", jwt))

	// Invoke RPC
	err = ch.Publish(
		"",
		"rpc_auth_validate_jwt_queue",
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
		return false
	}

	// Listen for RPC responses
	for msg := range messages {
		if corrId == msg.CorrelationId {
			response, err := strconv.ParseBool(string(msg.Body))

			if err != nil {
				rpc.logger.Error("Failed to convert body to bool", zap.Error(err))
				return false
			} else {
				rpc.logger.Debug("[<--]", zap.Bool("IsAuthenticated", response))
				return response
			}
		}
	}

	return false
}

func (rpc *RpcClient) Close() {
	rpc.connection.Close()
}
