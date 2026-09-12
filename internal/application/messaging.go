package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wakaba/wakaba/internal/domain"
	"github.com/wakaba/wakaba/internal/ports"
)

type MessagingService struct {
	messages    ports.MessageRepository
	conversations ports.ConversationRepository
	identities  ports.VirtualIdentityRepository
	phoneReg    ports.PhoneNumberRepository
	events      ports.EventPublisher
	webhooks    ports.EventPublisher
	logger      *slog.Logger
}

func NewMessagingService(
	messages ports.MessageRepository,
	conversations ports.ConversationRepository,
	identities ports.VirtualIdentityRepository,
	phoneReg ports.PhoneNumberRepository,
	events ports.EventPublisher,
	webhooks ports.EventPublisher,
	logger *slog.Logger,
) *MessagingService {
	return &MessagingService{
		messages:      messages,
		conversations: conversations,
		identities:    identities,
		phoneReg:      phoneReg,
		events:        events,
		webhooks:      webhooks,
		logger:        logger,
	}
}

type SendMessageRequest struct {
	From          string
	To            string
	Type          domain.MessageType
	Body          string
	PhoneNumberID string
}

type SendMessageResponse struct {
	MessageID  string
	WAMessageID string
}

func (s *MessagingService) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	fromPhone := domain.NormalizePhoneNumber(req.From)
	toPhone := domain.NormalizePhoneNumber(req.To)

	fromID := domain.PhoneNumberToWhatsAppID(fromPhone)
	toID := domain.PhoneNumberToWhatsAppID(toPhone)

	conv, err := s.conversations.GetConversationByPhone(ctx, toPhone, req.PhoneNumberID)
	if err != nil {
		return nil, fmt.Errorf("find conversation: %w", err)
	}
	if conv == nil {
		conv = domain.NewConversation(toPhone, req.PhoneNumberID)
		if err := s.conversations.CreateConversation(ctx, conv); err != nil {
			return nil, fmt.Errorf("create conversation: %w", err)
		}
	}

	msg := &domain.Message{
		ID:             domain.GenerateMessageID(),
		ConversationID: string(conv.ID),
		From:           fromID,
		To:             toID,
		Type:           req.Type,
		Status:         domain.MessageStatusPending,
		Direction:      domain.MessageDirectionOutbound,
		Text:           &domain.TextContent{Body: req.Body},
		Timestamp:      time.Now(),
	}

	if err := s.messages.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("store message: %w", err)
	}

	waMsgID := string(msg.ID)
	msg.WAMessageID = waMsgID
	if err := s.messages.UpdateMessageStatus(ctx, msg.ID, domain.MessageStatusAccepted); err != nil {
		s.logger.Error("failed to update message status", "error", err)
	}

	s.events.Publish(ctx, domain.NewDomainEvent(domain.EventOutboundMessageAccepted, msg))

	s.logger.Info("message accepted",
		"message_id", msg.ID,
		"from", fromID,
		"to", toID,
	)

	return &SendMessageResponse{
		MessageID:  string(msg.ID),
		WAMessageID: waMsgID,
	}, nil
}

type ReceiveMessageRequest struct {
	From            string
	To              string
	Type            domain.MessageType
	Body            string
	PhoneNumberID   string
	WAMessageID     string
}

type ReceiveMessageResponse struct {
	MessageID string
}

func (s *MessagingService) ReceiveMessage(ctx context.Context, req *ReceiveMessageRequest) (*ReceiveMessageResponse, error) {
	fromPhone := domain.NormalizePhoneNumber(req.From)
	toPhone := domain.NormalizePhoneNumber(req.To)

	fromID := domain.PhoneNumberToWhatsAppID(fromPhone)
	toID := domain.PhoneNumberToWhatsAppID(toPhone)

	conv, err := s.conversations.GetConversationByPhone(ctx, fromPhone, req.PhoneNumberID)
	if err != nil {
		return nil, fmt.Errorf("find conversation: %w", err)
	}
	if conv == nil {
		conv = domain.NewConversation(fromPhone, req.PhoneNumberID)
		if err := s.conversations.CreateConversation(ctx, conv); err != nil {
			return nil, fmt.Errorf("create conversation: %w", err)
		}
	}

	msg := &domain.Message{
		ID:             domain.GenerateMessageID(),
		WAMessageID:    req.WAMessageID,
		ConversationID: string(conv.ID),
		From:           fromID,
		To:             toID,
		Type:           req.Type,
		Status:         domain.MessageStatusDelivered,
		Direction:      domain.MessageDirectionInbound,
		Text:           &domain.TextContent{Body: req.Body},
		Timestamp:      time.Now(),
	}

	if err := s.messages.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("store message: %w", err)
	}

	now := time.Now()
	msg.SentAt = &now
	msg.DeliveredAt = &now

	s.conversations.UpdateLastMessage(ctx, conv.ID)
	s.conversations.IncrementUnread(ctx, conv.ID)

	s.events.Publish(ctx, domain.NewDomainEvent(domain.EventInboundMessageReceived, msg))
	s.webhooks.Publish(ctx, domain.NewDomainEvent(domain.EventWebhookQueued, msg))

	s.logger.Info("inbound message received",
		"message_id", msg.ID,
		"from", fromID,
		"to", toID,
	)

	return &ReceiveMessageResponse{
		MessageID: string(msg.ID),
	}, nil
}

func (s *MessagingService) UpdateStatus(ctx context.Context, messageID domain.MessageID, status domain.MessageStatus) error {
	if err := s.messages.UpdateMessageStatus(ctx, messageID, status); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	msg, err := s.messages.GetMessageByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("get message: %w", err)
	}

	now := time.Now()
	switch status {
	case domain.MessageStatusSent:
		msg.SentAt = &now
	case domain.MessageStatusDelivered:
		msg.DeliveredAt = &now
	case domain.MessageStatusRead:
		msg.ReadAt = &now
	}

	var eventType domain.EventType
	switch status {
	case domain.MessageStatusSent:
		eventType = domain.EventMessageSent
	case domain.MessageStatusDelivered:
		eventType = domain.EventMessageDelivered
	case domain.MessageStatusRead:
		eventType = domain.EventMessageRead
	case domain.MessageStatusFailed:
		eventType = domain.EventMessageFailed
	default:
		return nil
	}

	s.events.Publish(ctx, domain.NewDomainEvent(eventType, msg))
	s.webhooks.Publish(ctx, domain.NewDomainEvent(domain.EventWebhookQueued, msg))

	return nil
}

func (s *MessagingService) GetMessage(ctx context.Context, id domain.MessageID) (*domain.Message, error) {
	return s.messages.GetMessageByID(ctx, id)
}

func (s *MessagingService) GetMessages(ctx context.Context, conversationID domain.ConversationID, limit int, before *domain.MessageID) ([]*domain.Message, error) {
	return s.messages.GetMessagesByConversation(ctx, conversationID, limit, before)
}
