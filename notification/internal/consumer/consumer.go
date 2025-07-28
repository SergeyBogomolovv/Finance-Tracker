package consumer

import (
	"FinanceTracker/notification/pkg/logger"
	"context"
	"errors"
	"io"

	kafka "github.com/segmentio/kafka-go"
)

type consumerFactory struct {
	brokers []string
	groupID string
}

func NewConsumerFactory(brokers []string, groupID string) *consumerFactory {
	return &consumerFactory{
		brokers: brokers,
		groupID: groupID,
	}
}

type MessageHandler func(context.Context, kafka.Message) error

type consumer struct {
	handler func(context.Context, kafka.Message) error
	reader  *kafka.Reader
}

type Consumer interface {
	Consume(ctx context.Context)
	Close() error
}

func (f *consumerFactory) Create(topic string, handler MessageHandler) Consumer {
	return &consumer{
		handler: handler,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: f.brokers,
			GroupID: f.groupID,
			Topic:   topic,
		}),
	}
}

func (r *consumer) Consume(ctx context.Context) {
	for {
		m, err := r.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				break
			} else {
				logger.Error(ctx, "failed to fetch message", "err", err)
				continue
			}
		}
		if err := r.handler(ctx, m); err != nil {
			logger.Error(ctx, "failed to handle message", "err", err)
			continue
		}
		if err := r.reader.CommitMessages(ctx, m); err != nil {
			logger.Error(ctx, "failed to commit message", "err", err)
		}
		logger.Debug(ctx, "message consumed", "topic", m.Topic, "partition", m.Partition, "offset", m.Offset)
	}
}

func (r *consumer) Close() error {
	return r.reader.Close()
}
