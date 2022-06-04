package services

import (
	"dfs/storage/config"
	"dfs/storage/dtos"
	"dfs/storage/models"
	"encoding/json"
	"strconv"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RpcServer struct {
	config      *config.Config
	logger      *zap.Logger
	connection  *amqp.Connection
	database    *gorm.DB
	fileService *FileService
}

func NewRpcServer(cfg *config.Config, logger *zap.Logger, database *gorm.DB, fileService *FileService) *RpcServer {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(logger, err, "Failed to connect to RabbitMQ")
	return &RpcServer{config: cfg, logger: logger, connection: conn, database: database, fileService: fileService}
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

			response := rpc.fileService.CreateDirectory(directoryName)

			rpc.logger.Debug("[-->]", zap.Bool("IsHomeDirectoryCreated", response))

			rpc.publishAndAck(ch, msg, []byte(strconv.FormatBool(response)), "text/plain")
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

			var fileEntry *models.File = nil
			var ownedFileDto *dtos.OwnedFileDto = nil

			err := json.Unmarshal(msg.Body, &ownedFileDto)

			rpc.failOnError(err, "Cannot deserialize object to shareDto")

			if rpc.database.Where("id = ? AND owner_id = ?", ownedFileDto.FileId, ownedFileDto.OwnerId).First(&fileEntry).Error != nil {
				fileEntry = nil
			}

			serializedFileEntry, err := json.Marshal(fileEntry)

			rpc.failOnError(err, "Cannot serialize file entry")

			rpc.logger.Debug("[-->]", zap.String("FileEntry", string(serializedFileEntry)))

			rpc.publishAndAck(ch, msg, serializedFileEntry, "application/json")
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

			var fileEntry *models.File = nil
			var fileId uint = 0

			err := json.Unmarshal(msg.Body, &fileId)

			rpc.failOnError(err, "Cannot deserialize object to shareDto")

			if rpc.database.Where("id = ?", fileId).FirstOrInit(&fileEntry).Error != nil {
				fileEntry = nil
			}

			serializedFileEntry, err := json.Marshal(fileEntry)

			rpc.failOnError(err, "Cannot serialize file entry")

			rpc.logger.Debug("[-->]", zap.String("FileEntry", string(serializedFileEntry)))

			rpc.publishAndAck(ch, msg, serializedFileEntry, "application/json")
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

			var fileEntry *models.File = nil
			var fileUniqueName string

			err := json.Unmarshal(msg.Body, &fileUniqueName)

			rpc.failOnError(err, "Cannot deserialize object to shareDto")

			if rpc.database.Where("unique_name = ?", fileUniqueName).FirstOrInit(&fileEntry).Error != nil {
				fileEntry = nil
			}

			serializedFileEntry, err := json.Marshal(fileEntry)

			rpc.failOnError(err, "Cannot serialize file entry")

			rpc.logger.Debug("[-->]", zap.String("FileEntry", string(serializedFileEntry)))

			rpc.publishAndAck(ch, msg, serializedFileEntry, "application/json")
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

			var saveFileDto *dtos.SaveFileDto = nil
			err := json.Unmarshal(msg.Body, &saveFileDto)

			rpc.failOnError(err, "Cannot deserialize object to SaveFileDto")

			isSaved := rpc.fileService.EncryptAndSaveFile(saveFileDto.SavePath, saveFileDto.Content,
				saveFileDto.EncryptionKey)

			rpc.logger.Debug("[-->]", zap.Bool("IsFileSaved", isSaved))

			rpc.publishAndAck(ch, msg, []byte(strconv.FormatBool(isSaved)), "")
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

			isDeleted := rpc.fileService.RemoveFileFromDisk(deleteFileDto.FilePath)

			rpc.logger.Debug("[-->]", zap.Bool("IsFileDeleted", isDeleted))

			rpc.publishAndAck(ch, msg, []byte(strconv.FormatBool(isDeleted)), "")
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

			fileContent := rpc.fileService.DecryptAndReadFileContent(readFileDto.ReadPath, readFileDto.DecryptionKey)

			rpc.logger.Debug("[-->]", zap.ByteString("FileContent", fileContent))

			rpc.publishAndAck(ch, msg, fileContent, "text/plain")
		}
	}()

	rpc.logger.Info("[*] Awaiting 'ReadFileFromDisk' RPC requests")
	<-forever
}

func (rpc *RpcServer) Close() {
	rpc.connection.Close()
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

	// Don't dispatch a new message to this worker until it has  processed and acknowledged the previous one
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	rpc.failOnError(err, "Failed to set QoS")

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
