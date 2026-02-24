package rmq

import "context"

type Message struct {
	EventID string `json:"event_id"`
	Title   string `json:"title"`
	UserID  string `json:"user_id"`
	At      string `json:"at"` // ISO8601
}

type Publisher interface {
	Publish(queue string, msg Message) error
	Close() error
}

type Consumer interface {
	Consume(ctx context.Context, queue string, handler func(Message) error) error
	Close() error
}
