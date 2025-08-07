package controller

import (
	"FinanceTracker/profile/pkg/events"
	"FinanceTracker/profile/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/segmentio/kafka-go"
)

type EventsService interface {
	InitializeUserProfile(ctx context.Context, data events.EventUserRegistered) error
}

type eventsController struct {
	svc    EventsService
	reader *kafka.Reader
}

func NewEventsController(brokers []string, groupID string, svc EventsService) *eventsController {
	return &eventsController{
		svc: svc,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   events.TopicRegistered,
		}),
	}
}

func (c *eventsController) Consume(ctx context.Context) {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				break
			} else {
				logger.Error(ctx, "failed to fetch message", "err", err)
				continue
			}
		}

		if err := c.handleUserRegistered(ctx, m); err != nil {
			logger.Error(ctx, "failed to handle message", "err", err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			logger.Error(ctx, "failed to commit message", "err", err)
		}
	}
}

func (c *eventsController) handleUserRegistered(ctx context.Context, m kafka.Message) error {
	var event events.EventUserRegistered
	if err := decodeMessage(m, &event); err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}

	return c.svc.InitializeUserProfile(ctx, event)
}

func (c *eventsController) Close() error {
	return c.reader.Close()
}

func decodeMessage(m kafka.Message, event any) error {
	return json.Unmarshal(m.Value, event)
}
