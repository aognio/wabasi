package sqlite

import (
	"time"

	"gorm.io/gorm"
)

type MessageRecord struct {
	gorm.Model
	WAMessageID    string    `gorm:"index"`
	ExternalID     string    `gorm:"uniqueIndex"`
	ConversationID string    `gorm:"index"`
	FromPhone      string    `gorm:"index"`
	ToPhone        string    `gorm:"index"`
	Type           string    `gorm:"size:32"`
	Status         string    `gorm:"size:32;index"`
	Direction      string    `gorm:"size:16"`
	Body           string    `gorm:"type:text"`
	MediaID        string    `gorm:"index"`
	MediaMimeType  string    `gorm:"size:128"`
	SentAt         *time.Time
	DeliveredAt    *time.Time
	ReadAt         *time.Time
}

func (MessageRecord) TableName() string {
	return "messages"
}

type ConversationRecord struct {
	gorm.Model
	ExternalID       string    `gorm:"uniqueIndex"`
	PhoneNumber      string    `gorm:"index"`
	BusinessPhoneID  string    `gorm:"index"`
	LastMessageAt    *time.Time
	UnreadCount      int       `gorm:"default:0"`
}

func (ConversationRecord) TableName() string {
	return "conversations"
}

type VirtualIdentityRecord struct {
	gorm.Model
	ExternalID  string `gorm:"uniqueIndex"`
	Phone       string `gorm:"uniqueIndex"`
	DisplayName string
}

func (VirtualIdentityRecord) TableName() string {
	return "virtual_identities"
}

type PhoneNumberRecord struct {
	gorm.Model
	PhoneNumber string `gorm:"uniqueIndex"`
	OwnerType   string `gorm:"size:32"`
	OwnerID     string `gorm:"index"`
}

func (PhoneNumberRecord) TableName() string {
	return "phone_numbers"
}

type EventRecord struct {
	gorm.Model
	ExternalID string `gorm:"uniqueIndex"`
	Type       string `gorm:"index;size:64"`
	Payload    string `gorm:"type:text"`
}

func (EventRecord) TableName() string {
	return "events"
}

type WebhookDeliveryRecord struct {
	gorm.Model
	ExternalID      string `gorm:"uniqueIndex"`
	ApplicationID   string `gorm:"index"`
	CallbackURL     string
	EventType       string `gorm:"index;size:64"`
	Payload         string `gorm:"type:text"`
	Status          string `gorm:"size:32;index"`
	Attempts        int    `gorm:"default:0"`
	MaxAttempts     int    `gorm:"default:5"`
	NextRetryAt     *time.Time
	LastStatusCode  int
	LastError       string `gorm:"type:text"`
	Signature       string
}

func (WebhookDeliveryRecord) TableName() string {
	return "webhook_deliveries"
}
