package consumer

import (
	"FinanceTracker/notification/pkg/events"
	"context"
)

type controller struct {
	consumers []Consumer
}

func New(brokers []string, groupID string, svc NotificationService) *controller {
	factory := NewConsumerFactory(brokers, groupID)
	handler := NewHandler(svc)

	consumers := []Consumer{
		factory.Create(events.TopicOTPGenerated, handler.OTPGenerated),
		factory.Create(events.TopicRegistered, handler.UserRegistered),
	}

	return &controller{consumers}
}

func (c *controller) Start(ctx context.Context) {
	for _, consumer := range c.consumers {
		go consumer.Consume(ctx)
	}
}

func (c *controller) Stop() {
	for _, consumer := range c.consumers {
		consumer.Close()
	}
}
