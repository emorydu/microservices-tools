package rabbitmq

import (
	"context"
	"github.com/ahmetb/go-linq/v3"
	"github.com/emorydu/microservices-tools/pkg/logger"
	"github.com/emorydu/microservices-tools/pkg/otel"
	"github.com/iancoleman/strcase"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	uuid "github.com/satori/go.uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"time"

	//"github.com/meysamhadeli/shop-golang-microservices/internal/pkg/otel"
	"reflect"
)

type Publisher interface {
	PublishMessage(msg any) error
	IsPublished(msg any) bool
}

var publishedMessages []string

type publisher struct {
	cfg          *Config
	conn         *amqp.Connection
	log          logger.Logger
	jaegerTracer trace.Tracer
	ctx          context.Context
}

func (p publisher) PublishMessage(msg any) error {
	data, err := jsoniter.Marshal(msg)
	if err != nil {
		p.log.Error("Error in marshalling message to publish message")
		return err
	}

	typeName := reflect.TypeOf(msg).Elem().Name()
	snakeTypeName := strcase.ToSnake(typeName)

	ctx, span := p.jaegerTracer.Start(p.ctx, typeName)
	defer span.End()

	//Inject the context in the headers
	headers := otel.InjectAMQPHeaders(ctx)

	channel, err := p.conn.Channel()
	if err != nil {
		p.log.Error("Error in opening channel to consume message")
		return err
	}

	correlationId := ""
	if ctx.Value(echo.HeaderXCorrelationID) != nil {
		correlationId = ctx.Value(echo.HeaderXCorrelationID).(string)
	}

	publishingMsg := amqp.Publishing{
		Body:          data,
		ContentType:   "application/json",
		DeliveryMode:  amqp.Persistent,
		MessageId:     uuid.NewV4().String(),
		Timestamp:     time.Now(),
		CorrelationId: correlationId,
		Headers:       headers,
	}

	err = channel.Publish(snakeTypeName, snakeTypeName, false, false, publishingMsg)

	if err != nil {
		p.log.Error("Error in publishing message")
		return err
	}

	publishedMessages = append(publishedMessages, snakeTypeName)

	h, err := jsoniter.Marshal(headers)

	if err != nil {
		p.log.Error("Error in marshalling headers to publish message")
		return err
	}

	p.log.Infof("Published message: %s", publishingMsg.Body)
	span.SetAttributes(attribute.Key("message-id").String(publishingMsg.MessageId))
	span.SetAttributes(attribute.Key("correlation-id").String(publishingMsg.CorrelationId))
	span.SetAttributes(attribute.Key("exchange").String(snakeTypeName))
	span.SetAttributes(attribute.Key("kind").String(p.cfg.Kind))
	span.SetAttributes(attribute.Key("content-type").String("application/json"))
	span.SetAttributes(attribute.Key("timestamp").String(publishingMsg.Timestamp.String()))
	span.SetAttributes(attribute.Key("body").String(string(publishingMsg.Body)))
	span.SetAttributes(attribute.Key("headers").String(string(h)))

	return nil
}

func (p publisher) IsPublished(msg any) bool {
	typeName := reflect.TypeOf(msg).Elem().Name()
	snakeTypeName := strcase.ToSnake(typeName)

	return linq.From(publishedMessages).Contains(snakeTypeName)
}

func NewPublisher(ctx context.Context, cfg *Config, conn *amqp.Connection, log logger.Logger, jaegerTracer trace.Tracer) Publisher {
	return &publisher{
		cfg:          cfg,
		conn:         conn,
		log:          log,
		jaegerTracer: jaegerTracer,
		ctx:          ctx,
	}
}
