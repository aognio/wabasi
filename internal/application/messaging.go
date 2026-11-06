package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wabasi/wabasi/internal/domain"
	"github.com/wabasi/wabasi/internal/ports"
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
	PhoneNumberID string
	Body          string
	PreviewURL    bool
	MediaID       string
	Caption       string
	Filename      string
	Location      *LocationRequest
	Contacts      []ContactRequest
	Template      *TemplateRequest
	Interactive   *InteractiveRequest
	ReactionMessageID string
	ReactionEmoji     string
}

type LocationRequest struct {
	Latitude  float64
	Longitude float64
	Name      string
	Address   string
}

type ContactRequest struct {
	Name   *ContactNameRequest
	Phones []ContactPhoneRequest
}

type ContactNameRequest struct {
	FirstName     string
	LastName      string
	FormattedName string
}

type ContactPhoneRequest struct {
	Phone string
	Type  string
	WAID  string
}

type TemplateRequest struct {
	Name       string
	Language   string
	Components []TemplateComponentRequest
}

type TemplateComponentRequest struct {
	Type       string
	SubType    string
	Index      int
	Parameters []TemplateParameterRequest
}

type TemplateParameterRequest struct {
	Type string
	Text string
}

type InteractiveRequest struct {
	Type     string
	Header   *InteractiveHeaderRequest
	Body     string
	Foot     string
	Buttons  []InteractiveButtonRequest
	Sections []InteractiveSectionRequest
	Button   string
}

type InteractiveHeaderRequest struct {
	Type string
	Text string
}

type InteractiveButtonRequest struct {
	Type string
	Text string
	ID   string
}

type InteractiveSectionRequest struct {
	Title string
	Rows  []InteractiveRowRequest
}

type InteractiveRowRequest struct {
	ID          string
	Title       string
	Description string
}

type SendMessageResponse struct {
	MessageID  string
	WAMessageID string
}

func (s *MessagingService) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	traceID := domain.GenerateTraceID()
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
		Timestamp:      time.Now(),
	}

	switch req.Type {
	case domain.MessageTypeText:
		msg.Text = &domain.TextContent{Body: req.Body, PreviewURL: req.PreviewURL}

	case domain.MessageTypeImage:
		msg.Image = &domain.MediaContent{ID: req.MediaID, Caption: req.Caption}
		msg.MediaID = req.MediaID

	case domain.MessageTypeVideo:
		msg.Video = &domain.MediaContent{ID: req.MediaID, Caption: req.Caption}
		msg.MediaID = req.MediaID

	case domain.MessageTypeAudio:
		msg.Audio = &domain.MediaContent{ID: req.MediaID}
		msg.MediaID = req.MediaID

	case domain.MessageTypeDocument:
		msg.Document = &domain.MediaContent{ID: req.MediaID, Caption: req.Caption, Filename: req.Filename}
		msg.MediaID = req.MediaID

	case domain.MessageTypeLocation:
		msg.Location = &domain.LocationContent{
			Latitude:  req.Location.Latitude,
			Longitude: req.Location.Longitude,
			Name:      req.Location.Name,
			Address:   req.Location.Address,
		}

	case domain.MessageTypeContacts:
		for _, c := range req.Contacts {
			cc := domain.ContactContent{}
			if c.Name != nil {
				cc.Name = &domain.ContactName{
					FirstName:     c.Name.FirstName,
					LastName:      c.Name.LastName,
					FormattedName: c.Name.FormattedName,
				}
			}
			for _, p := range c.Phones {
				cc.Phones = append(cc.Phones, domain.ContactPhone{Phone: p.Phone, Type: p.Type, WAID: p.WAID})
			}
			msg.Contacts = append(msg.Contacts, cc)
		}

	case domain.MessageTypeTemplate:
		tc := &domain.TemplateContent{
			Name:     req.Template.Name,
			Language: domain.TemplateLanguage{Code: req.Template.Language},
		}
		for _, comp := range req.Template.Components {
			cr := domain.TemplateComponent{Type: comp.Type, SubType: comp.SubType, Index: comp.Index}
			for _, p := range comp.Parameters {
				cr.Parameters = append(cr.Parameters, domain.TemplateParameter{Type: p.Type, Text: p.Text})
			}
			tc.Components = append(tc.Components, cr)
		}
		msg.Template = tc

	case domain.MessageTypeInteractive:
		ic := &domain.InteractiveContent{
			Type: req.Interactive.Type,
			Body: domain.InteractiveBody{Text: req.Interactive.Body},
		}
		if req.Interactive.Header != nil {
			ic.Header = &domain.InteractiveHeader{Type: req.Interactive.Header.Type, Text: req.Interactive.Header.Text}
		}
		if req.Interactive.Foot != "" {
			ic.Foot = &domain.InteractiveFoot{Text: req.Interactive.Foot}
		}
		ic.Action.Button = req.Interactive.Button
		for _, btn := range req.Interactive.Buttons {
			ic.Action.Buttons = append(ic.Action.Buttons, domain.InteractiveButton{Type: btn.Type, Text: btn.Text, ID: btn.ID})
		}
		for _, sec := range req.Interactive.Sections {
			sr := domain.InteractiveSection{Title: sec.Title}
			for _, row := range sec.Rows {
				sr.Rows = append(sr.Rows, domain.InteractiveRow{ID: row.ID, Title: row.Title, Description: row.Description})
			}
			ic.Action.Sections = append(ic.Action.Sections, sr)
		}
		msg.Interactive = ic

	case domain.MessageTypeReaction:
		msg.Reaction = &domain.ReactionContent{
			MessageID: req.ReactionMessageID,
			Emoji:     req.ReactionEmoji,
		}
	}

	if err := s.messages.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("store message: %w", err)
	}

	waMsgID := string(msg.ID)
	msg.WAMessageID = waMsgID
	if err := s.messages.UpdateMessageStatus(ctx, msg.ID, domain.MessageStatusAccepted); err != nil {
		s.logger.Error("failed to update message status", "error", err)
	}

	s.events.Publish(ctx, domain.NewEvent(domain.EventOutboundMessageAccepted).
		WithTraceID(traceID).
		WithSource(domain.SourceMessaging).
		WithConversationID(msg.ConversationID).
		WithMessageID(string(msg.ID)).
		WithPayload(msg).
		Build())

	s.logger.Info("message accepted",
		"message_id", msg.ID,
		"type", msg.Type,
		"from", fromID,
		"to", toID,
		"trace_id", traceID,
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
	MediaID         string
	Caption         string
	Filename        string
	MediaMimeType   string
	Location        *LocationRequest
	Contacts        []ContactRequest
	InteractiveType     string
	InteractiveButtonID string
	ReactionMessageID   string
	ReactionEmoji       string
}

type ReceiveMessageResponse struct {
	MessageID string
}

func (s *MessagingService) ReceiveMessage(ctx context.Context, req *ReceiveMessageRequest) (*ReceiveMessageResponse, error) {
	traceID := domain.GenerateTraceID()
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
		Timestamp:      time.Now(),
	}

	switch req.Type {
	case domain.MessageTypeText:
		msg.Text = &domain.TextContent{Body: req.Body}
	case domain.MessageTypeImage:
		msg.Image = &domain.MediaContent{ID: req.MediaID, Caption: req.Caption, MimeType: req.MediaMimeType}
		msg.MediaID = req.MediaID
		msg.MediaMimeType = req.MediaMimeType
	case domain.MessageTypeVideo:
		msg.Video = &domain.MediaContent{ID: req.MediaID, Caption: req.Caption, MimeType: req.MediaMimeType}
		msg.MediaID = req.MediaID
		msg.MediaMimeType = req.MediaMimeType
	case domain.MessageTypeAudio:
		msg.Audio = &domain.MediaContent{ID: req.MediaID, MimeType: req.MediaMimeType}
		msg.MediaID = req.MediaID
		msg.MediaMimeType = req.MediaMimeType
	case domain.MessageTypeDocument:
		msg.Document = &domain.MediaContent{ID: req.MediaID, Caption: req.Caption, Filename: req.Filename, MimeType: req.MediaMimeType}
		msg.MediaID = req.MediaID
		msg.MediaMimeType = req.MediaMimeType
	case domain.MessageTypeLocation:
		if req.Location != nil {
			msg.Location = &domain.LocationContent{
				Latitude:  req.Location.Latitude,
				Longitude: req.Location.Longitude,
				Name:      req.Location.Name,
				Address:   req.Location.Address,
			}
		}
	case domain.MessageTypeContacts:
		for _, c := range req.Contacts {
			cc := domain.ContactContent{}
			if c.Name != nil {
				cc.Name = &domain.ContactName{
					FirstName:     c.Name.FirstName,
					LastName:      c.Name.LastName,
					FormattedName: c.Name.FormattedName,
				}
			}
			for _, p := range c.Phones {
				cc.Phones = append(cc.Phones, domain.ContactPhone{Phone: p.Phone, Type: p.Type, WAID: p.WAID})
			}
			msg.Contacts = append(msg.Contacts, cc)
		}
	case domain.MessageTypeInteractive:
		msg.Interactive = &domain.InteractiveContent{
			Type: req.InteractiveType,
			Body: domain.InteractiveBody{Text: req.InteractiveButtonID},
		}
	case domain.MessageTypeReaction:
		msg.Reaction = &domain.ReactionContent{
			MessageID: req.ReactionMessageID,
			Emoji:     req.ReactionEmoji,
		}
	}

	if err := s.messages.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("store message: %w", err)
	}

	now := time.Now()
	msg.SentAt = &now
	msg.DeliveredAt = &now

	s.conversations.UpdateLastMessage(ctx, conv.ID)
	s.conversations.IncrementUnread(ctx, conv.ID)

	s.events.Publish(ctx, domain.NewEvent(domain.EventInboundMessageReceived).
		WithTraceID(traceID).
		WithSource(domain.SourceMessaging).
		WithConversationID(msg.ConversationID).
		WithMessageID(string(msg.ID)).
		WithPayload(msg).
		Build())
	s.webhooks.Publish(ctx, domain.NewEvent(domain.EventWebhookQueued).
		WithTraceID(traceID).
		WithSource(domain.SourceWebhook).
		WithConversationID(msg.ConversationID).
		WithMessageID(string(msg.ID)).
		WithPayload(msg).
		Build())

	s.logger.Info("inbound message received",
		"message_id", msg.ID,
		"type", msg.Type,
		"from", fromID,
		"to", toID,
		"trace_id", traceID,
	)

	return &ReceiveMessageResponse{
		MessageID: string(msg.ID),
	}, nil
}

func (s *MessagingService) UpdateStatus(ctx context.Context, messageID domain.MessageID, status domain.MessageStatus) error {
	traceID := domain.GenerateTraceID()

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

	outcome := domain.EventOutcomeSuccess
	if status == domain.MessageStatusFailed {
		outcome = domain.EventOutcomeFailure
	}

	s.events.Publish(ctx, domain.NewEvent(eventType).
		WithTraceID(traceID).
		WithSource(domain.SourceStatusSim).
		WithConversationID(msg.ConversationID).
		WithMessageID(string(msg.ID)).
		WithOutcome(outcome).
		WithPayload(msg).
		Build())
	s.webhooks.Publish(ctx, domain.NewEvent(domain.EventWebhookQueued).
		WithTraceID(traceID).
		WithSource(domain.SourceWebhook).
		WithConversationID(msg.ConversationID).
		WithMessageID(string(msg.ID)).
		WithPayload(msg).
		Build())

	return nil
}

func (s *MessagingService) GetMessage(ctx context.Context, id domain.MessageID) (*domain.Message, error) {
	return s.messages.GetMessageByID(ctx, id)
}

func (s *MessagingService) GetMessages(ctx context.Context, conversationID domain.ConversationID, limit int, before *domain.MessageID) ([]*domain.Message, error) {
	return s.messages.GetMessagesByConversation(ctx, conversationID, limit, before)
}
