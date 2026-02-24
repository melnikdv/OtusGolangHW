package rmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/streadway/amqp"
)

type AMQP struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewAMQP(url string) (*AMQP, error) {
	var conn *amqp.Connection
	var err error

	op := func() error {
		conn, err = amqp.Dial(url)
		return err
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = 10 * time.Second
	bo.MaxElapsedTime = 30 * time.Second

	if err := backoff.Retry(op, bo); err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after retries: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &AMQP{conn: conn, ch: ch}, nil
}

func (a *AMQP) Publish(queue string, msg Message) error {
	body, _ := json.Marshal(msg)
	return a.ch.Publish("", queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (a *AMQP) Consume(ctx context.Context, queue string, handler func(Message) error) error {
	q, err := a.ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	msgs, err := a.ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case amqpMsg, ok := <-msgs:
			if !ok {
				return nil
			}
			var msg Message
			_ = json.Unmarshal(amqpMsg.Body, &msg) // игнорируем ошибку парсинга
			if err := handler(msg); err != nil {
				_ = handler(msg)
			}
		}
	}
}

func (a *AMQP) Close() error {
	_ = a.ch.Close()
	return a.conn.Close()
}
