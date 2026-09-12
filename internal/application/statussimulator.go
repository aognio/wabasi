package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wakaba/wakaba/internal/domain"
	"github.com/wakaba/wakaba/internal/ports"
)

type StatusSimulator struct {
	messages   ports.MessageRepository
	conversations ports.ConversationRepository
	events     ports.EventPublisher
	webhooks   ports.EventPublisher
	logger     *slog.Logger
	running    bool
	stopCh     chan struct{}
}

func NewStatusSimulator(
	messages ports.MessageRepository,
	conversations ports.ConversationRepository,
	events ports.EventPublisher,
	webhooks ports.EventPublisher,
	logger *slog.Logger,
) *StatusSimulator {
	return &StatusSimulator{
		messages:      messages,
		conversations: conversations,
		events:        events,
		webhooks:      webhooks,
		logger:        logger,
		stopCh:        make(chan struct{}),
	}
}

func (s *StatusSimulator) Start(ctx context.Context) {
	s.running = true
	s.logger.Info("status simulator started")
	go s.run(ctx)
}

func (s *StatusSimulator) Stop() {
	s.running = false
	close(s.stopCh)
	s.logger.Info("status simulator stopped")
}

func (s *StatusSimulator) run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			if !s.running {
				return
			}
			s.processOutbound(ctx)
		}
	}
}

func (s *StatusSimulator) processOutbound(ctx context.Context) {
	msgs, err := s.messages.GetMessagesNeedingStatusAdvance(ctx)
	if err != nil {
		s.logger.Error("failed to fetch messages for status advance", "error", err)
		return
	}

	for _, msg := range msgs {
		s.advanceStatus(ctx, msg)
	}
}

func (s *StatusSimulator) advanceStatus(ctx context.Context, msg *domain.Message) {
	now := time.Now()
	var nextStatus domain.MessageStatus
	var delay time.Duration

	switch msg.Status {
	case domain.MessageStatusAccepted:
		nextStatus = domain.MessageStatusSent
		delay = 1 * time.Second
	case domain.MessageStatusSent:
		nextStatus = domain.MessageStatusDelivered
		delay = 2 * time.Second
	case domain.MessageStatusDelivered:
		nextStatus = domain.MessageStatusRead
		delay = 5 * time.Second
	default:
		return
	}

	// Only advance if enough time has passed since the last status change
	var lastChange time.Time
	switch msg.Status {
	case domain.MessageStatusAccepted:
		lastChange = msg.Timestamp
	case domain.MessageStatusSent:
		if msg.SentAt != nil {
			lastChange = *msg.SentAt
		} else {
			lastChange = msg.Timestamp
		}
	case domain.MessageStatusDelivered:
		if msg.DeliveredAt != nil {
			lastChange = *msg.DeliveredAt
		} else if msg.SentAt != nil {
			lastChange = *msg.SentAt
		} else {
			lastChange = msg.Timestamp
		}
	}

	if now.Sub(lastChange) < delay {
		return
	}

	s.logger.Info("advancing message status",
		"message_id", msg.ID,
		"from", msg.Status,
		"to", nextStatus,
	)

	if err := s.messages.UpdateMessageStatus(ctx, msg.ID, nextStatus); err != nil {
		s.logger.Error("failed to update message status", "error", err, "message_id", msg.ID)
		return
	}

	updatedMsg, _ := s.messages.GetMessageByID(ctx, msg.ID)

	switch nextStatus {
	case domain.MessageStatusSent:
		if updatedMsg != nil {
			updatedMsg.SentAt = &now
		}
	case domain.MessageStatusDelivered:
		if updatedMsg != nil {
			updatedMsg.DeliveredAt = &now
		}
	case domain.MessageStatusRead:
		if updatedMsg != nil {
			updatedMsg.ReadAt = &now
		}
	}

	var eventType domain.EventType
	switch nextStatus {
	case domain.MessageStatusSent:
		eventType = domain.EventMessageSent
	case domain.MessageStatusDelivered:
		eventType = domain.EventMessageDelivered
	case domain.MessageStatusRead:
		eventType = domain.EventMessageRead
	}

	if updatedMsg != nil {
		s.events.Publish(ctx, domain.NewDomainEvent(eventType, updatedMsg))
		s.webhooks.Publish(ctx, domain.NewDomainEvent(domain.EventWebhookQueued, updatedMsg))
	}
}
