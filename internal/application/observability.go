package application

import (
	"context"
	"time"

	"github.com/wakaba/wakaba/internal/domain"
	"github.com/wakaba/wakaba/internal/ports"
)

type ObservabilityService struct {
	events    ports.EventReader
	messages  ports.MessageRepository
	convs     ports.ConversationRepository
	identities ports.VirtualIdentityRepository
}

func NewObservabilityService(
	events ports.EventReader,
	messages ports.MessageRepository,
	convs ports.ConversationRepository,
	identities ports.VirtualIdentityRepository,
) *ObservabilityService {
	return &ObservabilityService{
		events:     events,
		messages:   messages,
		convs:      convs,
		identities: identities,
	}
}

type WorldSummary struct {
	ConsumerCount    int64                        `json:"consumer_count"`
	BusinessCount    int64                        `json:"business_count"`
	MessageCount     int64                        `json:"message_count"`
	ConversationCount int64                       `json:"conversation_count"`
	EventCount       int64                        `json:"event_count"`
	FailureCount     int64                        `json:"failure_count"`
	EventsByType     map[domain.EventType]int64   `json:"events_by_type"`
	EventsByOutcome  map[domain.EventOutcome]int64 `json:"events_by_outcome"`
}

func (s *ObservabilityService) GetWorldSummary(ctx context.Context) (*WorldSummary, error) {
	identities, err := s.identities.ListIdentities(ctx)
	if err != nil {
		return nil, err
	}

	since24h := time.Now().Add(-24 * time.Hour)

	eventsByType, err := s.events.CountByType(ctx, &since24h)
	if err != nil {
		return nil, err
	}

	eventsByOutcome, err := s.events.CountByOutcome(ctx, &since24h)
	if err != nil {
		return nil, err
	}

	totalEvents := int64(0)
	for _, c := range eventsByType {
		totalEvents += c
	}

	failureCount := eventsByOutcome[domain.EventOutcomeFailure]

	return &WorldSummary{
		ConsumerCount:     int64(len(identities)),
		EventCount:        totalEvents,
		FailureCount:      failureCount,
		EventsByType:      eventsByType,
		EventsByOutcome:   eventsByOutcome,
	}, nil
}

type TraceResult struct {
	TraceID domain.TraceID      `json:"trace_id"`
	Steps   []*domain.DomainEvent `json:"steps"`
}

func (s *ObservabilityService) GetTrace(ctx context.Context, traceID domain.TraceID) (*TraceResult, error) {
	events, err := s.events.GetByTraceID(ctx, traceID)
	if err != nil {
		return nil, err
	}
	return &TraceResult{
		TraceID: traceID,
		Steps:   events,
	}, nil
}

type MessageTimeline struct {
	MessageID string                `json:"message_id"`
	Events    []*domain.DomainEvent `json:"events"`
}

func (s *ObservabilityService) GetMessageTimeline(ctx context.Context, messageID string) (*MessageTimeline, error) {
	events, err := s.events.GetByMessageID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	return &MessageTimeline{
		MessageID: messageID,
		Events:    events,
	}, nil
}

type ConversationTimeline struct {
	ConversationID string                `json:"conversation_id"`
	Events         []*domain.DomainEvent `json:"events"`
}

func (s *ObservabilityService) GetConversationTimeline(ctx context.Context, conversationID string, limit int) (*ConversationTimeline, error) {
	if limit <= 0 {
		limit = 100
	}
	events, err := s.events.GetByConversationID(ctx, conversationID, limit)
	if err != nil {
		return nil, err
	}
	return &ConversationTimeline{
		ConversationID: conversationID,
		Events:         events,
	}, nil
}

type EventListResult struct {
	Events []*domain.DomainEvent `json:"events"`
	Total  int64                 `json:"total"`
}

func (s *ObservabilityService) ListEvents(ctx context.Context, filter ports.EventFilter) (*EventListResult, error) {
	events, err := s.events.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	total, err := s.events.Count(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &EventListResult{
		Events: events,
		Total:  total,
	}, nil
}

type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int64     `json:"count"`
}

type MetricsSummary struct {
	Period         string                      `json:"period"`
	EventsByType   map[domain.EventType]int64  `json:"events_by_type"`
	EventsByOutcome map[domain.EventOutcome]int64 `json:"events_by_outcome"`
	TotalEvents    int64                       `json:"total_events"`
	FailureRate    float64                     `json:"failure_rate"`
}

func (s *ObservabilityService) GetMetricsSummary(ctx context.Context, since time.Duration) (*MetricsSummary, error) {
	sinceTime := time.Now().Add(-since)

	eventsByType, err := s.events.CountByType(ctx, &sinceTime)
	if err != nil {
		return nil, err
	}

	eventsByOutcome, err := s.events.CountByOutcome(ctx, &sinceTime)
	if err != nil {
		return nil, err
	}

	total := int64(0)
	for _, c := range eventsByType {
		total += c
	}

	var failureRate float64
	if total > 0 {
		failureRate = float64(eventsByOutcome[domain.EventOutcomeFailure]) / float64(total) * 100
	}

	return &MetricsSummary{
		Period:          since.String(),
		EventsByType:    eventsByType,
		EventsByOutcome: eventsByOutcome,
		TotalEvents:     total,
		FailureRate:     failureRate,
	}, nil
}
