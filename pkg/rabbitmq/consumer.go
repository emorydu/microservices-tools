package rabbitmq

import (
	"context"
	"github.com/ahmetb/go-linq/v3"
	"github.com/emorydu/microservices-tools/pkg/logger"
	"github.com/iancoleman/strcase"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
	"reflect"
	"time"
)

type Consumer[T any] interface {
	ConsumeMessage(msg any, dependencies T) error
	IsConsumed(msg any) bool
}

var consumedMessages []string

type consumer[T any] struct {
	cfg          *Config
	conn         *amqp.Connection
	log          logger.Logger
	handler      func(queue string, msg amqp.Delivery, dependencies T) error
	jaegerTracer trace.Tracer
	ctx          context.Context
}

func (c consumer[T]) ConsumeMessage(msg any, dependencies T) error {
	return nil
}

func (c consumer[T]) IsConsumed(msg any) bool {
	timeoutTime := 20 * time.Second
	startTime := time.Now()
	timeoutExpired := false
	isConsumed := false

	for {
		if timeoutExpired {
			return false
		}
		if isConsumed {
			return true
		}
		time.Sleep(time.Second * 2)
		typeName := reflect.TypeOf(msg).Elem().Name()
		snakeTypeName := strcase.ToSnake(typeName)

		isConsumed = linq.From(consumedMessages).Contains(snakeTypeName)
		timeoutExpired = time.Now().Sub(startTime) > timeoutTime
	}
}
