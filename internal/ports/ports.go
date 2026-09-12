package ports

import (
	"context"
	"time"

	"github.com/wakaba/wakaba/internal/domain"
)

type MessageRepository interface {
	CreateMessage(ctx context.Context, msg *domain.Message) error
	GetMessageByID(ctx context.Context, id domain.MessageID) (*domain.Message, error)
	GetMessagesByConversation(ctx context.Context, conversationID domain.ConversationID, limit int, before *domain.MessageID) ([]*domain.Message, error)
	UpdateMessageStatus(ctx context.Context, id domain.MessageID, status domain.MessageStatus) error
	GetMessagesNeedingStatusAdvance(ctx context.Context) ([]*domain.Message, error)
}

type ConversationRepository interface {
	CreateConversation(ctx context.Context, conv *domain.Conversation) error
	GetConversationByID(ctx context.Context, id domain.ConversationID) (*domain.Conversation, error)
	GetConversationByPhone(ctx context.Context, phone domain.PhoneNumber, businessPhoneID string) (*domain.Conversation, error)
	GetConversationsByPhone(ctx context.Context, phone domain.PhoneNumber) ([]*domain.Conversation, error)
	UpdateLastMessage(ctx context.Context, id domain.ConversationID) error
	IncrementUnread(ctx context.Context, id domain.ConversationID) error
	ResetUnread(ctx context.Context, id domain.ConversationID) error
}

type VirtualIdentityRepository interface {
	CreateIdentity(ctx context.Context, identity *domain.VirtualIdentity) error
	GetIdentityByID(ctx context.Context, id domain.VirtualIdentityID) (*domain.VirtualIdentity, error)
	GetIdentityByPhone(ctx context.Context, phone domain.PhoneNumber) (*domain.VirtualIdentity, error)
	ListIdentities(ctx context.Context) ([]*domain.VirtualIdentity, error)
}

type PhoneNumberRepository interface {
	Register(ctx context.Context, phone domain.PhoneNumber, ownerType string, ownerID string) error
	IsAvailable(ctx context.Context, phone domain.PhoneNumber) (bool, error)
	GetOwner(ctx context.Context, phone domain.PhoneNumber) (string, string, error)
}

type BusinessRepository interface {
	Create(ctx context.Context, b *domain.Business) error
	GetByID(ctx context.Context, id domain.BusinessID) (*domain.Business, error)
	List(ctx context.Context) ([]*domain.Business, error)
}

type WABARepository interface {
	Create(ctx context.Context, w *domain.WABA) error
	GetByID(ctx context.Context, id domain.WABAID) (*domain.WABA, error)
	GetByBusiness(ctx context.Context, businessID domain.BusinessID) ([]*domain.WABA, error)
	List(ctx context.Context) ([]*domain.WABA, error)
}

type BusinessPhoneNumberRepository interface {
	Create(ctx context.Context, p *domain.BusinessPhoneNumber) error
	GetByID(ctx context.Context, id domain.BusinessPhoneNumberID) (*domain.BusinessPhoneNumber, error)
	GetByPhone(ctx context.Context, phone domain.PhoneNumber) (*domain.BusinessPhoneNumber, error)
	GetByWABA(ctx context.Context, wabaID domain.WABAID) ([]*domain.BusinessPhoneNumber, error)
	List(ctx context.Context) ([]*domain.BusinessPhoneNumber, error)
}

type ApplicationRepository interface {
	Create(ctx context.Context, a *domain.Application) error
	GetByID(ctx context.Context, id domain.ApplicationID) (*domain.Application, error)
	GetByBusiness(ctx context.Context, businessID domain.BusinessID) ([]*domain.Application, error)
	List(ctx context.Context) ([]*domain.Application, error)
}

type CredentialRepository interface {
	Create(ctx context.Context, c *domain.Credential) error
	GetByID(ctx context.Context, id domain.CredentialID) (*domain.Credential, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Credential, error)
	GetByApplication(ctx context.Context, appID domain.ApplicationID) ([]*domain.Credential, error)
	Revoke(ctx context.Context, id domain.CredentialID) error
}

type WebhookConfigRepository interface {
	Create(ctx context.Context, w *domain.WebhookConfig) error
	GetByApplication(ctx context.Context, appID domain.ApplicationID) (*domain.WebhookConfig, error)
	Update(ctx context.Context, w *domain.WebhookConfig) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event *domain.DomainEvent) error
}

type EventFilter struct {
	Types           []domain.EventType
	Source          domain.EventSource
	ConversationID  string
	MessageID       string
	ApplicationID   string
	ActorID         string
	TraceID         domain.TraceID
	Outcome         domain.EventOutcome
	Since           *time.Time
	Until           *time.Time
	Limit           int
	Offset          int
}

type EventReader interface {
	GetByID(ctx context.Context, id domain.EventID) (*domain.DomainEvent, error)
	List(ctx context.Context, filter EventFilter) ([]*domain.DomainEvent, error)
	Count(ctx context.Context, filter EventFilter) (int64, error)
	GetByTraceID(ctx context.Context, traceID domain.TraceID) ([]*domain.DomainEvent, error)
	GetByMessageID(ctx context.Context, messageID string) ([]*domain.DomainEvent, error)
	GetByConversationID(ctx context.Context, conversationID string, limit int) ([]*domain.DomainEvent, error)
	GetByApplicationID(ctx context.Context, applicationID string, since *time.Time, limit int) ([]*domain.DomainEvent, error)
	CountByType(ctx context.Context, since *time.Time) (map[domain.EventType]int64, error)
	CountByOutcome(ctx context.Context, since *time.Time) (map[domain.EventOutcome]int64, error)
}

type MediaStore interface {
	Store(ctx context.Context, filename string, data []byte, mimeType string) (string, error)
	Get(ctx context.Context, id string) ([]byte, string, error)
	GetURL(ctx context.Context, id string) (string, error)
}
