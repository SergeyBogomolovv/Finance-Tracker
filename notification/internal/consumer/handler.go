package consumer

import (
	"FinanceTracker/notification/pkg/events"
	"context"
	"encoding/json"
	"fmt"

	kafka "github.com/segmentio/kafka-go"
)

type MailService interface {
	SendOTP(ctx context.Context, email, code string) error
	SendRegistered(ctx context.Context, email, name string) error
}

type handler struct {
	svc MailService
}

func NewHandler(svc MailService) *handler {
	return &handler{svc: svc}
}

func (h *handler) OTPGenerated(ctx context.Context, m kafka.Message) error {
	var event events.EventOTPGenerated
	if err := decodeMessage(m, &event); err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}

	return h.svc.SendOTP(ctx, event.Email, event.Code)
}

func (h *handler) UserRegistered(ctx context.Context, m kafka.Message) error {
	var event events.EventUserRegistered
	if err := decodeMessage(m, &event); err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}

	return h.svc.SendRegistered(ctx, event.Email, event.FullName)
}

func decodeMessage(m kafka.Message, event any) error {
	return json.Unmarshal(m.Value, event)
}
