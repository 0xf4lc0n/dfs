package services

import (
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
