package sender

import (
	"context"

	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/rmq"
	"github.com/sirupsen/logrus"
)

type Sender struct {
	consumer rmq.Consumer
	logger   *logrus.Logger
	queue    string
}

func New(consumer rmq.Consumer, logger *logrus.Logger, queue string) *Sender {
	return &Sender{
		consumer: consumer,
		logger:   logger,
		queue:    queue,
	}
}

func (s *Sender) Run(ctx context.Context) error {
	handler := func(msg rmq.Message) error {
		s.logger.Infof("Sending notification: %+v", msg)
		return nil
	}
	return s.consumer.Consume(ctx, s.queue, handler)
}
