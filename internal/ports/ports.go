package ports

import (
	"context"

	"github.com/wakaba/wakaba/internal/domain"
)

type MessageRepository interface {
	CreateMessage(ctx context.Context, msg *domain.Message) error
	GetMessageByID(ctx context.Context, id domain.MessageID) (*domain.Message, error)
	GetMessagesByConversation(ctx context.Context, conversationID domain.ConversationID, limit int, before *domain.MessageID) ([]*domain.Message, error)
	UpdateMessageStatus(ctx context.Context, id domain.MessageID, status domain.MessageStatus) error
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

type EventPublisher interface {
	Publish(ctx context.Context, event *domain.DomainEvent) error
}

type MediaStore interface {
	Store(ctx context.Context, filename string, data []byte, mimeType string) (string, error)
	Get(ctx context.Context, id string) ([]byte, string, error)
	GetURL(ctx context.Context, id string) (string, error)
}
