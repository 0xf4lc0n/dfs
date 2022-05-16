package services

import (
	"dfs/auth/models"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RpcServer struct {
	logger     *zap.Logger
	connection *amqp.Connection
	db         *gorm.DB
}

func NewRpcServer(logger *zap.Logger, db *gorm.DB) *RpcServer {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(logger, err, "Failed to connect to RabbitMQ")
	return &RpcServer{logger: logger, db: db, connection: conn}
}

func (rpc *RpcServer) RegisterValidateJwt() {
	ch, err := rpc.connection.Channel()
	rpc.failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Storage service RPC queue aka RPC Server
	rpcQueue, err := ch.QueueDeclare(
		"rpc_auth_validate_jwt_queue", // name
		false,                         // durable
		false,                         // delete when unused
		false,                         // exclusive
		false,                         // no-wait
		nil,                           // arguments
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
			rawToken := string(msg.Body)

			rpc.logger.Debug("[<--]", zap.String("Jwt", rawToken))

			const SecretKey = "secret"
			token, err := jwt.ParseWithClaims(rawToken, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(SecretKey), nil
			})

			var response bool

			if err != nil {
				rpc.logger.Debug("Cannot parse token", zap.Error(err))
				response = false
			} else {
				response = token.Valid
			}

			rpc.logger.Debug("[-->]", zap.Bool("Authenticated", response))

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

	rpc.logger.Info("[*] Awaiting 'ValidateJwt' RPC requests")
	<-forever
}

func (rpc *RpcServer) RegisterGetUserHomeDirectory() {
	ch, err := rpc.connection.Channel()
	rpc.failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Storage service RPC queue aka RPC Server
	rpcQueue, err := ch.QueueDeclare(
		"rpc_auth_get_user_home_dir_queue", // name
		false,                              // durable
		false,                              // delete when unused
		false,                              // exclusive
		false,                              // no-wait
		nil,                                // arguments
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
			rawToken := string(msg.Body)

			rpc.logger.Debug("[<--]", zap.String("Jwt", rawToken))

			const SecretKey = "secret"
			token, err := jwt.ParseWithClaims(rawToken, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(SecretKey), nil
			})

			homeDir := ""

			if err != nil {
				rpc.logger.Warn("Cannot parse token", zap.Error(err))
			} else {
				claims := token.Claims.(*jwt.StandardClaims)

				var user models.User

				if rpc.db.Where("id = ?", claims.Issuer).First(&user).Error == nil {
					homeDir = user.HomeDirectory
				}
			}

			rpc.logger.Debug("[-->]", zap.String("HomeDirectory", homeDir))

			// Send message to client callback queue
			err = ch.Publish(
				"",          // exchange
				msg.ReplyTo, // routing key
				false,       // mandatory
				false,       // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: msg.CorrelationId,
					Body:          []byte(homeDir),
				})

			rpc.failOnError(err, "Failed to publish a message")

			// Send manual acknowledgement
			msg.Ack(false)
		}
	}()

	rpc.logger.Info("[*] Awaiting 'GetUserHomeDirectory' RPC requests")
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
