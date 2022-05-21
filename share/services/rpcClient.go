package services

import (
	"dfs/share/dtos"
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

func (rpc *RpcClient) GetUserDataByJwt(jwt string) *dtos.UserDto {
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

			var userDto *dtos.UserDto = nil
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

func (rpc *RpcClient) GetOwnedFile(shareDto *dtos.OwnedFileDto) *dtos.FileDto {
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

	serializedOwnedFileDto, err := json.Marshal(shareDto)

	if err != nil {
		rpc.logger.Fatal("Cannot serialize ShareDto", zap.Error(err))
	}

	rpc.logger.Debug("[-->]", zap.ByteString("SerializedShareDto", serializedOwnedFileDto))

	// Invoke RPC
	err = ch.Publish(
		"",
		"rpc_storage_get_owned_file_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       callbackQueue.Name,
			Body:          serializedOwnedFileDto,
		},
	)

	if err != nil {
		rpc.logger.Error("Failed to publish a message", zap.Error(err))
		return nil
	}
	// Listen for RPC responses
	for msg := range messages {
		if corrId == msg.CorrelationId {
			rpc.logger.Debug("[<--]", zap.ByteString("SerializedFileDto", msg.Body))

			var fileDto *dtos.FileDto = nil
			err := json.Unmarshal(msg.Body, &fileDto)

			if err != nil {
				rpc.logger.Error("Cannot deserialize data to UserDto", zap.Error(err))
				return nil
			}

			return fileDto
		}
	}

	return nil
}

func (rpc *RpcClient) GetFileById(fileId uint) *dtos.FileDto {
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

	serializedFileId, err := json.Marshal(fileId)

	if err != nil {
		rpc.logger.Fatal("Cannot serialize ShareDto", zap.Error(err))
	}

	rpc.logger.Debug("[-->]", zap.ByteString("SerializedFileId", serializedFileId))

	// Invoke RPC
	err = ch.Publish(
		"",
		"rpc_storage_get_file_by_id_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       callbackQueue.Name,
			Body:          serializedFileId,
		},
	)

	if err != nil {
		rpc.logger.Error("Failed to publish a message", zap.Error(err))
		return nil
	}
	// Listen for RPC responses
	for msg := range messages {
		if corrId == msg.CorrelationId {
			rpc.logger.Debug("[<--]", zap.ByteString("SerializedFileDto", msg.Body))

			var fileDto *dtos.FileDto = nil
			err := json.Unmarshal(msg.Body, &fileDto)

			if err != nil {
				rpc.logger.Error("Cannot deserialize data to UserDto", zap.Error(err))
				return nil
			}

			return fileDto
		}
	}

	return nil
}

func (rpc *RpcClient) GetUserDataById(userId uint) *dtos.UserDto {
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

	serializedId, err := json.Marshal(userId)

	if err != nil {
		rpc.logger.Fatal("Cannot serialize UserId", zap.Error(err))
	}

	rpc.logger.Debug("[-->]", zap.ByteString("UserId", serializedId))

	// Invoke RPC
	err = ch.Publish(
		"",
		"rpc_auth_get_user_data_by_id_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       callbackQueue.Name,
			Body:          serializedId,
		},
	)

	if err != nil {
		rpc.logger.Error("Failed to publish a message", zap.Error(err))
		return nil
	}
	// Listen for RPC responses
	for msg := range messages {
		if corrId == msg.CorrelationId {
			rpc.logger.Debug("[<--]", zap.ByteString("SerializedUserDto", msg.Body))

			var userDto *dtos.UserDto = nil
			err := json.Unmarshal(msg.Body, &userDto)

			if err != nil {
				rpc.logger.Fatal("Cannot deserialize data to UserDto", zap.Error(err))
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
