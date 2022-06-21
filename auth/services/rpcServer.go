package services

import (
	"dfs/auth/dtos"
	"dfs/auth/models"
	"encoding/json"
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

func (rpc *RpcServer) RegisterGetUserDataByJwt() {
	ch, _, messages := rpc.createQueue("rpc_auth_get_user_data_by_jwt_queue")
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		// Listen and process each RPC request
		for msg := range messages {
			rawToken := string(msg.Body)

			rpc.logger.Debug("[<--]", zap.String("Jwt", rawToken))

			const SecretKey = "secret"
			token, err := jwt.ParseWithClaims(rawToken, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(SecretKey), nil
			})

			var user models.User
			var userDto *dtos.User = nil

			if err != nil {
				rpc.logger.Warn("Cannot parse token", zap.Error(err))
			} else {
				claims := token.Claims.(*jwt.StandardClaims)

				if err := rpc.db.Where("id = ?", claims.Issuer).First(&user).Error; err != nil {
					rpc.logger.Debug("Cannot find user", zap.Error(err))
				} else {
					userDto = &dtos.User{
						Id:            user.Id,
						Name:          user.Name,
						Email:         user.Email,
						Verified:      user.Verified,
						HomeDirectory: user.HomeDirectory,
						CryptKey:      user.CryptKey,
					}
				}
			}

			serializedUser, err := json.Marshal(userDto)

			if err != nil {
				rpc.logger.Error("Cannot serialize userDto to JSON", zap.Error(err))
			}

			rpc.logger.Debug("[-->]", zap.String("UserData", string(serializedUser)))

			rpc.publishAndAck(ch, msg, serializedUser, "application/json")
		}
	}()

	rpc.logger.Info("[*] Awaiting 'GetUserDataByJwt' RPC requests")
	<-forever
}

func (rpc *RpcServer) RegisterGetUserDataById() {
	ch, _, messages := rpc.createQueue("rpc_auth_get_user_data_by_id_queue")
	defer ch.Close()

	forever := make(chan bool)

	go func() {
		// Listen and process each RPC request
		for msg := range messages {
			rpc.logger.Debug("[<--]", zap.ByteString("SharedToId", msg.Body))

			var userId uint

			err := json.Unmarshal(msg.Body, &userId)

			if err != nil {
				rpc.logger.Error("Cannot deserialize SharedToId to uint", zap.Error(err))
			}

			// rpc.failOnError(err, "Cannot parse SharedToId to uint")

			var user models.User
			var userDto *dtos.User = nil

			if err := rpc.db.Where("id = ?", userId).First(&user).Error; err != nil {
				rpc.logger.Debug("Cannot find user with id", zap.Uint("UserID", userId))
			} else {
				userDto = &dtos.User{
					Id:            user.Id,
					Name:          user.Name,
					Email:         user.Email,
					Verified:      user.Verified,
					HomeDirectory: user.HomeDirectory,
					CryptKey:      user.CryptKey,
				}
			}

			serializedUser, err := json.Marshal(userDto)

			if err != nil {
				rpc.logger.Error("Cannot serialize userDto to JSON", zap.Error(err))
			}

			rpc.logger.Debug("[-->]", zap.String("UserData", string(serializedUser)))

			rpc.publishAndAck(ch, msg, serializedUser, "application/json")
		}
	}()

	rpc.logger.Info("[*] Awaiting 'GetUserDataById' RPC requests")
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
