package services

import (
	"dfs/share/dtos"
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

func (rpc *RpcClient) GetUserDataByJwt(jwt string) *dtos.UserDto {
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
	ch, callbackQueue, messages := rpc.createCallbackQueue()
	defer ch.Close()

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
				rpc.logger.Error("Cannot deserialize data to FileDto", zap.Error(err))
				return nil
			}

			return fileDto
		}
	}

	return nil
}

func (rpc *RpcClient) GetFileById(fileId uint) *dtos.FileDto {
	ch, callbackQueue, messages := rpc.createCallbackQueue()
	defer ch.Close()

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

func (rpc *RpcClient) GetFileByUniqueName(uniqueName string) *dtos.FileDto {
	ch, callbackQueue, messages := rpc.createCallbackQueue()
	defer ch.Close()

	// Generate correlation ID for RPC
	corrId := uuid.New().String()

	serializedFileUniqueName, err := json.Marshal(uniqueName)

	if err != nil {
		rpc.logger.Fatal("Cannot serialize ShareDto", zap.Error(err))
	}

	rpc.logger.Debug("[-->]", zap.ByteString("SerializedFileUniqueName", serializedFileUniqueName))

	// Invoke RPC
	err = ch.Publish(
		"",
		"rpc_storage_get_file_by_unique_name_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       callbackQueue.Name,
			Body:          serializedFileUniqueName,
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
	ch, callbackQueue, messages := rpc.createCallbackQueue()
	defer ch.Close()

	// Generate correlation ID for RPC
	corrId := uuid.New().String()

	serializedId, err := json.Marshal(userId)

	if err != nil {
		rpc.logger.Fatal("Cannot serialize SharedToId", zap.Error(err))
	}

	rpc.logger.Debug("[-->]", zap.ByteString("SharedToId", serializedId))

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

func (rpc *RpcClient) ReadFileFromDisk(readFileDto dtos.ReadFileDto) []byte {
	ch, callbackQueue, messages := rpc.createCallbackQueue()
	defer ch.Close()

	// Generate correlation ID for RPC
	corrId := uuid.New().String()

	serializedReadFileDto, err := json.Marshal(readFileDto)

	rpc.failOnError(err, "Cannot serialize ReadFileDto")

	rpc.logger.Debug("[-->]", zap.ByteString("SerializedReadFileDto", serializedReadFileDto))

	// Invoke RPC
	err = ch.Publish(
		"",
		"rpc_storage_get_file_content",
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       callbackQueue.Name,
			Body:          serializedReadFileDto,
		},
	)

	rpc.failOnError(err, "Failed to publish message")

	// Listen for RPC responses
	for msg := range messages {
		if corrId == msg.CorrelationId {
			rpc.logger.Debug("[<--]", zap.ByteString("FileContent", msg.Body))
			return msg.Body
		}
	}

	return nil
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

func (rpc *RpcClient) Close() {
	rpc.connection.Close()
}

func (rpc *RpcClient) failOnError(err error, msg string) {
	if err != nil {
		rpc.logger.Fatal("RPC", zap.String("Msg", msg), zap.Error(err))
	}
}
