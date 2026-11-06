package sqlite

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/wabasi/wabasi/internal/domain"
	"gorm.io/gorm"
)

type EventPublisher struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewEventPublisher(db *gorm.DB, logger *slog.Logger) *EventPublisher {
	return &EventPublisher{db: db, logger: logger}
}

func (p *EventPublisher) Publish(ctx context.Context, event *domain.DomainEvent) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	record := EventRecord{
		ExternalID:     string(event.ID),
		Type:           string(event.Type),
		Payload:        string(payload),
		TraceID:        string(event.TraceID),
		Source:         string(event.Source),
		ActorID:        event.ActorID,
		ConversationID: event.ConversationID,
		MessageID:      event.MessageID,
		ApplicationID:  event.ApplicationID,
		Outcome:        string(event.Outcome),
		DurationMs:     event.DurationMs,
	}

	if err := p.db.Create(&record).Error; err != nil {
		p.logger.Error("failed to store event", "error", err, "event_type", event.Type)
		return err
	}

	p.logger.Info("event published",
		"event_id", event.ID,
		"type", event.Type,
		"source", event.Source,
		"trace_id", event.TraceID,
		"timestamp", event.Timestamp,
	)

	return nil
}

type WebhookDispatcher struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewWebhookDispatcher(db *gorm.DB, logger *slog.Logger) *WebhookDispatcher {
	return &WebhookDispatcher{db: db, logger: logger}
}

func (d *WebhookDispatcher) Publish(ctx context.Context, event *domain.DomainEvent) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	record := WebhookDeliveryRecord{
		ExternalID:  string(event.ID),
		EventType:   string(event.Type),
		Payload:     string(payload),
		Status:      "pending",
		MaxAttempts: 5,
		NextRetryAt: func() *time.Time { t := time.Now(); return &t }(),
	}

	if err := d.db.Create(&record).Error; err != nil {
		d.logger.Error("failed to queue webhook", "error", err, "event_type", event.Type)
		return err
	}

	d.logger.Info("webhook queued",
		"event_id", event.ID,
		"type", event.Type,
	)

	return nil
}
