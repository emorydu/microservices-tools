package rabbitmq

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"time"
)

type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	ExchangeName string
	Kind         string
}

// NewRabbitMQConn initialize new channel for rabbitmq.
func NewRabbitMQConn(ctx context.Context, cfg *Config) (*amqp.Connection, error) {
	addr := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port)

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 10 * time.Second // Maximum time to retry
	maxRetries := 5                      // Number of retries (including the initial attempt)

	var conn *amqp.Connection
	var err error

	err = backoff.Retry(func() error {
		conn, err = amqp.Dial(addr)
		if err != nil {
			log.Errorf("Failed to connect to RabbitMQ: %v. Connection information: %s", err, addr)
			return err
		}

		return nil
	}, backoff.WithMaxRetries(bo, uint64(maxRetries-1)))

	log.Info("Connected to RabbitMQ")

	go func() {
		select {
		case <-ctx.Done():
			err := conn.Close()
			if err != nil {
				log.Error("Failed to close RabbitMQ connection")
			}

			log.Info("RabbitMQ connection is closed")
		}
	}()

	return conn, nil
}
