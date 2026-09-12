package sqlite

import (
	"context"
	"fmt"

	"github.com/wakaba/wakaba/internal/domain"
	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) CreateMessage(ctx context.Context, msg *domain.Message) error {
	record := MessageRecord{
		ExternalID:     string(msg.ID),
		WAMessageID:    msg.WAMessageID,
		ConversationID: msg.ConversationID,
		FromPhone:      string(msg.From),
		ToPhone:        string(msg.To),
		Type:           string(msg.Type),
		Status:         string(msg.Status),
		Direction:      string(msg.Direction),
		SentAt:         msg.SentAt,
		DeliveredAt:    msg.DeliveredAt,
		ReadAt:         msg.ReadAt,
		MediaID:        msg.MediaID,
		MediaMimeType:  msg.MediaMimeType,
	}

	if msg.Text != nil {
		record.Body = msg.Text.Body
	}

	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("create message: %w", err)
	}
	return nil
}

func (r *MessageRepository) GetMessageByID(ctx context.Context, id domain.MessageID) (*domain.Message, error) {
	var record MessageRecord
	if err := r.db.WithContext(ctx).Where("external_id = ?", string(id)).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get message: %w", err)
	}
	return r.toDomain(&record), nil
}

func (r *MessageRepository) GetMessagesByConversation(ctx context.Context, conversationID domain.ConversationID, limit int, before *domain.MessageID) ([]*domain.Message, error) {
	query := r.db.WithContext(ctx).Where("conversation_id = ?", string(conversationID))
	if before != nil {
		query = query.Where("external_id < ?", string(*before))
	}
	query = query.Order("external_id DESC").Limit(limit)

	var records []MessageRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	result := make([]*domain.Message, len(records))
	for i, rec := range records {
		result[i] = r.toDomain(&rec)
	}
	return result, nil
}

func (r *MessageRepository) UpdateMessageStatus(ctx context.Context, id domain.MessageID, status domain.MessageStatus) error {
	updates := map[string]interface{}{
		"status": string(status),
	}
	result := r.db.WithContext(ctx).Model(&MessageRecord{}).Where("external_id = ?", string(id)).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update status: %w", result.Error)
	}
	return nil
}

func (r *MessageRepository) toDomain(rec *MessageRecord) *domain.Message {
	msg := &domain.Message{
		ID:             domain.MessageID(rec.ExternalID),
		WAMessageID:    rec.WAMessageID,
		ConversationID: rec.ConversationID,
		From:           domain.WhatsAppID(rec.FromPhone),
		To:             domain.WhatsAppID(rec.ToPhone),
		Type:           domain.MessageType(rec.Type),
		Status:         domain.MessageStatus(rec.Status),
		Direction:      domain.MessageDirection(rec.Direction),
		Timestamp:      rec.CreatedAt,
		SentAt:         rec.SentAt,
		DeliveredAt:    rec.DeliveredAt,
		ReadAt:         rec.ReadAt,
		MediaID:        rec.MediaID,
		MediaMimeType:  rec.MediaMimeType,
	}

	if rec.Body != "" {
		msg.Text = &domain.TextContent{Body: rec.Body}
	}

	return msg
}
