package sqlite

import (
	"context"
	"encoding/json"
	"time"

	"github.com/wabasi/wabasi/internal/domain"
	"github.com/wabasi/wabasi/internal/ports"
	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

var _ ports.EventReader = (*EventRepository)(nil)

func (r *EventRepository) GetByID(ctx context.Context, id domain.EventID) (*domain.DomainEvent, error) {
	var record EventRecord
	if err := r.db.WithContext(ctx).Where("external_id = ?", string(id)).First(&record).Error; err != nil {
		return nil, err
	}
	return r.toDomain(&record)
}

func (r *EventRepository) List(ctx context.Context, filter ports.EventFilter) ([]*domain.DomainEvent, error) {
	query := r.db.WithContext(ctx).Model(&EventRecord{})
	query = r.applyFilter(query, filter)

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	query = query.Order("created_at DESC")

	var records []EventRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}

	events := make([]*domain.DomainEvent, 0, len(records))
	for i := range records {
		ev, err := r.toDomain(&records[i])
		if err != nil {
			continue
		}
		events = append(events, ev)
	}
	return events, nil
}

func (r *EventRepository) Count(ctx context.Context, filter ports.EventFilter) (int64, error) {
	query := r.db.WithContext(ctx).Model(&EventRecord{})
	query = r.applyFilter(query, filter)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *EventRepository) GetByTraceID(ctx context.Context, traceID domain.TraceID) ([]*domain.DomainEvent, error) {
	var records []EventRecord
	if err := r.db.WithContext(ctx).Where("trace_id = ?", string(traceID)).Order("created_at ASC").Find(&records).Error; err != nil {
		return nil, err
	}

	events := make([]*domain.DomainEvent, 0, len(records))
	for i := range records {
		ev, err := r.toDomain(&records[i])
		if err != nil {
			continue
		}
		events = append(events, ev)
	}
	return events, nil
}

func (r *EventRepository) GetByMessageID(ctx context.Context, messageID string) ([]*domain.DomainEvent, error) {
	var records []EventRecord
	if err := r.db.WithContext(ctx).Where("message_id = ?", messageID).Order("created_at ASC").Find(&records).Error; err != nil {
		return nil, err
	}

	events := make([]*domain.DomainEvent, 0, len(records))
	for i := range records {
		ev, err := r.toDomain(&records[i])
		if err != nil {
			continue
		}
		events = append(events, ev)
	}
	return events, nil
}

func (r *EventRepository) GetByConversationID(ctx context.Context, conversationID string, limit int) ([]*domain.DomainEvent, error) {
	query := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var records []EventRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}

	events := make([]*domain.DomainEvent, 0, len(records))
	for i := range records {
		ev, err := r.toDomain(&records[i])
		if err != nil {
			continue
		}
		events = append(events, ev)
	}
	return events, nil
}

func (r *EventRepository) GetByApplicationID(ctx context.Context, applicationID string, since *time.Time, limit int) ([]*domain.DomainEvent, error) {
	query := r.db.WithContext(ctx).Where("application_id = ?", applicationID)
	if since != nil {
		query = query.Where("created_at >= ?", *since)
	}
	query = query.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var records []EventRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}

	events := make([]*domain.DomainEvent, 0, len(records))
	for i := range records {
		ev, err := r.toDomain(&records[i])
		if err != nil {
			continue
		}
		events = append(events, ev)
	}
	return events, nil
}

func (r *EventRepository) CountByType(ctx context.Context, since *time.Time) (map[domain.EventType]int64, error) {
	query := r.db.WithContext(ctx).Model(&EventRecord{})
	if since != nil {
		query = query.Where("created_at >= ?", *since)
	}

	var results []struct {
		Type  string
		Count int64
	}
	if err := query.Select("type, count(*) as count").Group("type").Find(&results).Error; err != nil {
		return nil, err
	}

	counts := make(map[domain.EventType]int64, len(results))
	for _, r := range results {
		counts[domain.EventType(r.Type)] = r.Count
	}
	return counts, nil
}

func (r *EventRepository) CountByOutcome(ctx context.Context, since *time.Time) (map[domain.EventOutcome]int64, error) {
	query := r.db.WithContext(ctx).Model(&EventRecord{})
	if since != nil {
		query = query.Where("created_at >= ?", *since)
	}

	var results []struct {
		Outcome string
		Count   int64
	}
	if err := query.Select("outcome, count(*) as count").Group("outcome").Find(&results).Error; err != nil {
		return nil, err
	}

	counts := make(map[domain.EventOutcome]int64, len(results))
	for _, r := range results {
		counts[domain.EventOutcome(r.Outcome)] = r.Count
	}
	return counts, nil
}

func (r *EventRepository) applyFilter(query *gorm.DB, filter ports.EventFilter) *gorm.DB {
	if len(filter.Types) > 0 {
		types := make([]string, len(filter.Types))
		for i, t := range filter.Types {
			types[i] = string(t)
		}
		query = query.Where("type IN ?", types)
	}
	if filter.Source != "" {
		query = query.Where("source = ?", string(filter.Source))
	}
	if filter.ConversationID != "" {
		query = query.Where("conversation_id = ?", filter.ConversationID)
	}
	if filter.MessageID != "" {
		query = query.Where("message_id = ?", filter.MessageID)
	}
	if filter.ApplicationID != "" {
		query = query.Where("application_id = ?", filter.ApplicationID)
	}
	if filter.ActorID != "" {
		query = query.Where("actor_id = ?", filter.ActorID)
	}
	if filter.TraceID != "" {
		query = query.Where("trace_id = ?", string(filter.TraceID))
	}
	if filter.Outcome != "" {
		query = query.Where("outcome = ?", string(filter.Outcome))
	}
	if filter.Since != nil {
		query = query.Where("created_at >= ?", *filter.Since)
	}
	if filter.Until != nil {
		query = query.Where("created_at <= ?", *filter.Until)
	}
	return query
}

func (r *EventRepository) toDomain(record *EventRecord) (*domain.DomainEvent, error) {
	var payload any
	if record.Payload != "" {
		if err := json.Unmarshal([]byte(record.Payload), &payload); err != nil {
			payload = record.Payload
		}
	}

	createdAt := record.CreatedAt

	return &domain.DomainEvent{
		ID:             domain.EventID(record.ExternalID),
		Type:           domain.EventType(record.Type),
		Payload:        payload,
		Timestamp:      createdAt,
		TraceID:        domain.TraceID(record.TraceID),
		Source:         domain.EventSource(record.Source),
		ActorID:        record.ActorID,
		ConversationID: record.ConversationID,
		MessageID:      record.MessageID,
		ApplicationID:  record.ApplicationID,
		Outcome:        domain.EventOutcome(record.Outcome),
		DurationMs:     record.DurationMs,
	}, nil
}
