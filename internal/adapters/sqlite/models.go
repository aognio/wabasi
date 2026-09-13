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
	ExternalID      string    `gorm:"uniqueIndex"`
	PhoneNumber     string    `gorm:"index"`
	BusinessPhoneID string    `gorm:"index"`
	LastMessageAt   *time.Time
	UnreadCount     int `gorm:"default:0"`
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
	ExternalID     string `gorm:"uniqueIndex"`
	Type           string `gorm:"index;size:64"`
	Payload        string `gorm:"type:text"`
	TraceID        string `gorm:"index"`
	Source         string `gorm:"index;size:32"`
	ActorID        string `gorm:"index"`
	ConversationID string `gorm:"index"`
	MessageID      string `gorm:"index"`
	ApplicationID  string `gorm:"index"`
	Outcome        string `gorm:"index;size:16"`
	DurationMs     int64
}

func (EventRecord) TableName() string {
	return "events"
}

type WebhookDeliveryRecord struct {
	gorm.Model
	ExternalID    string `gorm:"uniqueIndex"`
	ApplicationID string `gorm:"index"`
	CallbackURL   string
	EventType     string `gorm:"index;size:64"`
	Payload       string `gorm:"type:text"`
	Status        string `gorm:"size:32;index"`
	Attempts      int    `gorm:"default:0"`
	MaxAttempts   int    `gorm:"default:5"`
	NextRetryAt   *time.Time
	LastStatusCode int
	LastError     string `gorm:"type:text"`
	Signature     string
}

func (WebhookDeliveryRecord) TableName() string {
	return "webhook_deliveries"
}

type BusinessRecord struct {
	gorm.Model
	ExternalID  string `gorm:"uniqueIndex"`
	Name        string
	DisplayName string
	Country     string
	Status      string `gorm:"size:32"`
}

func (BusinessRecord) TableName() string {
	return "businesses"
}

type WABARecord struct {
	gorm.Model
	ExternalID string `gorm:"uniqueIndex"`
	BusinessID string `gorm:"index"`
	Name       string
	Timezone   string
	Currency   string
	Status     string `gorm:"size:32"`
}

func (WABARecord) TableName() string {
	return "wabas"
}

type BusinessPhoneNumberRecord struct {
	gorm.Model
	ExternalID  string `gorm:"uniqueIndex"`
	WABAID      string `gorm:"index"`
	Phone       string `gorm:"uniqueIndex"`
	DisplayName string
	Country     string
	Status      string `gorm:"size:32"`
}

func (BusinessPhoneNumberRecord) TableName() string {
	return "business_phone_numbers"
}

type ApplicationRecord struct {
	gorm.Model
	ExternalID string `gorm:"uniqueIndex"`
	BusinessID string `gorm:"index"`
	Name       string
	Status     string `gorm:"size:32"`
	AppSecret  string
}

func (ApplicationRecord) TableName() string {
	return "applications"
}

type CredentialRecord struct {
	gorm.Model
	ExternalID    string `gorm:"uniqueIndex"`
	ApplicationID string `gorm:"index"`
	TokenHash     string `gorm:"index"`
	Status        string `gorm:"size:32"`
	RevokedAt     *time.Time
}

func (CredentialRecord) TableName() string {
	return "credentials"
}

type WebhookConfigRecord struct {
	gorm.Model
	ApplicationID string `gorm:"uniqueIndex"`
	CallbackURL   string
	VerifyToken   string
	AppSecret     string
	Enabled       bool
}

func (WebhookConfigRecord) TableName() string {
	return "webhook_configs"
}

type APIKeyRecord struct {
	gorm.Model
	ExternalID string `gorm:"uniqueIndex"`
	Name       string
	KeyHash    string `gorm:"index"`
	KeyPrefix  string `gorm:"size:16"`
	Scopes     string `gorm:"type:text"`
	Status     string `gorm:"size:32"`
	RevokedAt  *time.Time
}

func (APIKeyRecord) TableName() string {
	return "api_keys"
}
