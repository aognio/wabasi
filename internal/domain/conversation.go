package domain

import "time"

type ConversationID string

type Conversation struct {
	ID                ConversationID `json:"id"`
	PhoneNumber       PhoneNumber    `json:"phone_number"`
	BusinessPhoneID   string         `json:"business_phone_id,omitempty"`
	PhoneNumberID     string         `json:"phone_number_id,omitempty"`
	LastMessageAt     *time.Time     `json:"last_message_at,omitempty"`
	UnreadCount       int            `json:"unread_count"`
	CreatedAt         time.Time      `json:"created_at"`
}

func NewConversation(phone PhoneNumber, businessPhoneID string) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:              ConversationID(GenerateMessageID()),
		PhoneNumber:     phone,
		BusinessPhoneID: businessPhoneID,
		CreatedAt:       now,
	}
}
