package consumer

import (
	"FinanceTracker/notification/pkg/events"
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/IBM/sarama"
)

type NotificationService interface {
	SendOTP(ctx context.Context, userID int, code string) error
}

type EventHandler func(ctx context.Context, msg *sarama.ConsumerMessage)

type consumer struct {
	master   sarama.Consumer
	svc      NotificationService
	handlers map[string]EventHandler
	logger   *slog.Logger
}

func MustNew(logger *slog.Logger, brokers []string, svc NotificationService) *consumer {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		logger.Error("failed to create consumer", "err", err)
		os.Exit(1)
	}

	consumer := &consumer{master: master, svc: svc, logger: logger.With(slog.String("layer", "consumer"))}

	consumer.handlers = map[string]EventHandler{
		events.AuthOTPGeneratedTopic: consumer.handleOTPGenerated,
	}

	return consumer
}

func (c *consumer) ConsumeTopic(ctx context.Context, topic string) {
	partitions, err := c.master.Partitions(topic)
	if err != nil {
		c.logger.Error("failed to get partitions", "err", err)
		return
	}

	for _, partition := range partitions {
		pc, err := c.master.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			c.logger.Error("failed to subscribe partition", "err", err)
			return
		}

		go func(pc sarama.PartitionConsumer) {
			defer pc.Close()
			for {
				select {
				case msg := <-pc.Messages():
					handler, ok := c.handlers[msg.Topic]
					if !ok {
						c.logger.Error("unknown topic", "topic", msg.Topic)
						continue
					}
					handler(ctx, msg)
				case err := <-pc.Errors():
					c.logger.Error("failed to consume message", "err", err)
				case <-ctx.Done():
					return
				}
			}
		}(pc)
	}
}

func (c *consumer) Close() error {
	return c.master.Close()
}

func (c *consumer) handleOTPGenerated(ctx context.Context, msg *sarama.ConsumerMessage) {
	var payload events.OTPGeneratedEvent
	if err := decodeMessage(msg, &payload); err != nil {
		c.logger.Error("failed to decode message", "err", err)
		return
	}

	if err := c.svc.SendOTP(ctx, payload.UserID, payload.Code); err != nil {
		c.logger.Error("failed to send otp", "err", err)
		return
	}

	c.logger.Info("sent otp", "user_id", payload.UserID)
}

func decodeMessage(msg *sarama.ConsumerMessage, dest any) error {
	return json.Unmarshal(msg.Value, dest)
}
