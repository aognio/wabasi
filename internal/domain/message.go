package domain

import "time"

type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeVideo    MessageType = "video"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeDocument MessageType = "document"
)

type MessageStatus string

const (
	MessageStatusPending     MessageStatus = "pending"
	MessageStatusAccepted    MessageStatus = "accepted"
	MessageStatusSent        MessageStatus = "sent"
	MessageStatusDelivered   MessageStatus = "delivered"
	MessageStatusRead        MessageStatus = "read"
	MessageStatusFailed      MessageStatus = "failed"
)

type MessageDirection string

const (
	MessageDirectionInbound  MessageDirection = "inbound"
	MessageDirectionOutbound MessageDirection = "outbound"
)

type TextContent struct {
	Body       string `json:"body"`
	PreviewURL bool   `json:"preview_url,omitempty"`
}

type ImageContent struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

type Message struct {
	ID            MessageID       `json:"id"`
	WAMessageID   string          `json:"wamessage_id,omitempty"`
	ConversationID string        `json:"conversation_id"`
	From          WhatsAppID      `json:"from"`
	To            WhatsAppID      `json:"to"`
	Type          MessageType     `json:"type"`
	Status        MessageStatus   `json:"status"`
	Direction     MessageDirection `json:"direction"`
	Text          *TextContent    `json:"text,omitempty"`
	Image         *ImageContent   `json:"image,omitempty"`
	Timestamp     time.Time       `json:"timestamp"`
	SentAt        *time.Time      `json:"sent_at,omitempty"`
	DeliveredAt   *time.Time      `json:"delivered_at,omitempty"`
	ReadAt        *time.Time      `json:"read_at,omitempty"`
	MediaID       string          `json:"media_id,omitempty"`
	MediaMimeType string          `json:"media_mime_type,omitempty"`
}

func NewInboundTextMessage(from, to WhatsAppID, body string, conversationID string) *Message {
	return &Message{
		ID:             GenerateMessageID(),
		ConversationID: conversationID,
		From:           from,
		To:             to,
		Type:           MessageTypeText,
		Status:         MessageStatusDelivered,
		Direction:      MessageDirectionInbound,
		Text:           &TextContent{Body: body},
		Timestamp:      time.Now(),
	}
}

func NewOutboundTextMessage(from, to WhatsAppID, body string, conversationID string) *Message {
	now := time.Now()
	return &Message{
		ID:             GenerateMessageID(),
		ConversationID: conversationID,
		From:           from,
		To:             to,
		Type:           MessageTypeText,
		Status:         MessageStatusPending,
		Direction:      MessageDirectionOutbound,
		Text:           &TextContent{Body: body},
		Timestamp:      now,
		SentAt:         &now,
	}
}
