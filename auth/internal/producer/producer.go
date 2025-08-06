package producer

import (
	"FinanceTracker/auth/pkg/events"
	"context"
	"encoding/json"
	"fmt"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type producer struct {
	userRegisteredWriter *kafka.Writer
	otpGeneratedWriter   *kafka.Writer
}

func New(brokers []string, batchTimeout time.Duration) *producer {
	addr := kafka.TCP(brokers...)

	return &producer{
		userRegisteredWriter: &kafka.Writer{
			Addr:                   addr,
			Topic:                  events.TopicRegistered,
			AllowAutoTopicCreation: true,
			BatchTimeout:           batchTimeout,
		},
		otpGeneratedWriter: &kafka.Writer{
			Addr:                   addr,
			Topic:                  events.TopicOTPGenerated,
			AllowAutoTopicCreation: true,
			BatchTimeout:           batchTimeout,
		},
	}
}

func (p *producer) PublishUserRegistered(ctx context.Context, event events.EventUserRegistered) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return p.userRegisteredWriter.WriteMessages(ctx, kafka.Message{Value: data})
}

func (p *producer) PublishOTPGenerated(ctx context.Context, event events.EventOTPGenerated) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return p.otpGeneratedWriter.WriteMessages(ctx, kafka.Message{Value: data})
}

func (p *producer) Close() error {
	if err := p.userRegisteredWriter.Close(); err != nil {
		return fmt.Errorf("failed to close user registered writer: %w", err)
	}
	if err := p.otpGeneratedWriter.Close(); err != nil {
		return fmt.Errorf("failed to close otp generated writer: %w", err)
	}
	return nil
}
