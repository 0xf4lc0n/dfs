package services

import (
	"fmt"
	"os"
	"strconv"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RpcServer struct {
	logger     *zap.Logger
	connection *amqp.Connection
}

func NewRpcServer(logger *zap.Logger) *RpcServer {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(logger, err, "Failed to connect to RabbitMQ")
	return &RpcServer{logger: logger, connection: conn}
}

func (rpc *RpcServer) RegisterCreateHomeDirectory() {
	ch, err := rpc.connection.Channel()
	rpc.failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Storage service RPC queue aka RPC Server
	rpcQueue, err := ch.QueueDeclare(
		"rpc_storage_create_home_dir_queue", // name
		false,                               // durable
		false,                               // delete when unused
		false,                               // exclusive
		false,                               // no-wait
		nil,                                 // arguments
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

	forever := make(chan bool)

	go func() {
		// Listen and preocess each RPC request
		for msg := range messages {
			directoryName := string(msg.Body)

			rpc.logger.Debug("[<--]", zap.String("HomeDirectory", directoryName))

			var response bool

			const STORAGE_PATH = `C:\Users\Falcon\Desktop\Files\`

			directoryPath := fmt.Sprintf("%s%s", STORAGE_PATH, directoryName)

			err := os.Mkdir(directoryPath, 0755)

			if err != nil {
				rpc.logger.Warn("Cannot create home directory", zap.Error(err))
				response = false
			} else {
				response = true
			}

			rpc.logger.Debug("[-->]", zap.Bool("IsHomeDirectoryCreated", response))

			// Send message to client callback queue
			err = ch.Publish(
				"",          // exchange
				msg.ReplyTo, // routing key
				false,       // mandatory
				false,       // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: msg.CorrelationId,
					Body:          []byte(strconv.FormatBool(response)),
				})

			rpc.failOnError(err, "Failed to publish a message")

			// Send manual acknowledgement
			msg.Ack(false)
		}
	}()

	rpc.logger.Info("[*] Awaiting 'CreateHomeDirectory' RPC requests")
	<-forever
}

func (rpc *RpcServer) Close() {
	rpc.connection.Close()
}

func (rpc *RpcServer) failOnError(err error, msg string) {
	failOnError(rpc.logger, err, msg)
}

func failOnError(logger *zap.Logger, err error, msg string) {
	if err != nil {
		logger.Fatal("RPC", zap.String("Msg", msg), zap.Error(err))
	}
}
