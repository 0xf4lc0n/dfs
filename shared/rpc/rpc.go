package rpc

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

func IsAuthenticated(jwt string, logger *zap.Logger) bool {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", zap.Error(err))
		return false
	}

	defer conn.Close()

	ch, err := conn.Channel()

	if err != nil {
		logger.Error("Failed to open a channel", zap.Error(err))
		return false
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)

	if err != nil {
		logger.Error("Failed to declare a queue", zap.Error(err))
		return false
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		logger.Error("Failed to register a consumer", zap.Error(err))
		return false
	}

	corrId := uuid.New().String()

	err = ch.Publish(
		"",
		"rpc_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          []byte(jwt),
		},
	)

	logger.Info("Published: {Jwt}", zap.String("Jwt", jwt))

	if err != nil {
		logger.Error("Failed to publish a message", zap.Error(err))
		return false
	}

	for d := range msgs {
		if corrId == d.CorrelationId {
			res, err := strconv.ParseBool(string(d.Body))

			if err != nil {
				logger.Error("Failed to convert body to bool", zap.Error(err))
				return false
			} else {
				return res
			}
		}
	}

	return false

}
