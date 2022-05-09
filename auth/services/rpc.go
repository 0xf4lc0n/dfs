package services

import (
	"log"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RpcService struct {
	logger *zap.Logger
}

func NewRpcService(logger *zap.Logger) *RpcService {
	return &RpcService{logger: logger}
}

func (rpc *RpcService) RegisterValidateJwt() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	rpc.failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	rpc.failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rpc_queue", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)

	rpc.failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	rpc.failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	rpc.failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			tokener := string(d.Body)
			rpc.failOnError(err, "Failed to convert body to integer")

			rpc.logger.Info(" [.] jwt: {Jwt}", zap.String("Jwt", tokener))

			const SecretKey = "secret"
			token, err := jwt.ParseWithClaims(tokener, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(SecretKey), nil
			})

			var response bool

			if err != nil {
				rpc.logger.Warn("Cannot parse token", zap.Error(err))
				response = false
			} else {
				response = token.Valid
			}

			err = ch.Publish(
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(strconv.FormatBool(response)),
				})

			rpc.failOnError(err, "Failed to publish a message")

			d.Ack(false)
		}
	}()

	log.Printf(" [*] Awaiting RPC requests")
	<-forever
}

func (rpc *RpcService) failOnError(err error, msg string) {
	if err != nil {
		rpc.logger.Fatal("{Msg}", zap.String("Msg", msg), zap.Error(err))
	}
}
