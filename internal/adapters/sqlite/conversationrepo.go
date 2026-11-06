package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/wabasi/wabasi/internal/domain"
	"gorm.io/gorm"
)

type ConversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) CreateConversation(ctx context.Context, conv *domain.Conversation) error {
	record := ConversationRecord{
		ExternalID:      string(conv.ID),
		PhoneNumber:     string(conv.PhoneNumber),
		BusinessPhoneID: conv.BusinessPhoneID,
		UnreadCount:     conv.UnreadCount,
	}

	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("create conversation: %w", err)
	}
	return nil
}

func (r *ConversationRepository) GetConversationByID(ctx context.Context, id domain.ConversationID) (*domain.Conversation, error) {
	var record ConversationRecord
	if err := r.db.WithContext(ctx).Where("external_id = ?", string(id)).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get conversation: %w", err)
	}
	return r.toDomain(&record), nil
}

func (r *ConversationRepository) GetConversationByPhone(ctx context.Context, phone domain.PhoneNumber, businessPhoneID string) (*domain.Conversation, error) {
	var record ConversationRecord
	if err := r.db.WithContext(ctx).
		Where("phone_number = ? AND business_phone_id = ?", string(phone), businessPhoneID).
		First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get conversation by phone: %w", err)
	}
	return r.toDomain(&record), nil
}

func (r *ConversationRepository) GetConversationsByPhone(ctx context.Context, phone domain.PhoneNumber) ([]*domain.Conversation, error) {
	var records []ConversationRecord
	if err := r.db.WithContext(ctx).Where("phone_number = ?", string(phone)).Order("updated_at DESC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("get conversations: %w", err)
	}

	result := make([]*domain.Conversation, len(records))
	for i, rec := range records {
		result[i] = r.toDomain(&rec)
	}
	return result, nil
}

func (r *ConversationRepository) UpdateLastMessage(ctx context.Context, id domain.ConversationID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&ConversationRecord{}).
		Where("external_id = ?", string(id)).
		Update("last_message_at", now)
	if result.Error != nil {
		return fmt.Errorf("update last message: %w", result.Error)
	}
	return nil
}

func (r *ConversationRepository) IncrementUnread(ctx context.Context, id domain.ConversationID) error {
	result := r.db.WithContext(ctx).Model(&ConversationRecord{}).
		Where("external_id = ?", string(id)).
		Update("unread_count", gorm.Expr("unread_count + 1"))
	if result.Error != nil {
		return fmt.Errorf("increment unread: %w", result.Error)
	}
	return nil
}

func (r *ConversationRepository) ResetUnread(ctx context.Context, id domain.ConversationID) error {
	result := r.db.WithContext(ctx).Model(&ConversationRecord{}).
		Where("external_id = ?", string(id)).
		Update("unread_count", 0)
	if result.Error != nil {
		return fmt.Errorf("reset unread: %w", result.Error)
	}
	return nil
}

func (r *ConversationRepository) toDomain(rec *ConversationRecord) *domain.Conversation {
	return &domain.Conversation{
		ID:              domain.ConversationID(rec.ExternalID),
		PhoneNumber:     domain.PhoneNumber(rec.PhoneNumber),
		BusinessPhoneID: rec.BusinessPhoneID,
		LastMessageAt:   rec.LastMessageAt,
		UnreadCount:     rec.UnreadCount,
		CreatedAt:       rec.CreatedAt,
	}
}
