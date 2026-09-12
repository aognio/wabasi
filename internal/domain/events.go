package domain

import "time"

type EventID string
type EventType string
type TraceID string
type EventSource string

const (
	EventInboundMessageReceived   EventType = "inbound_message_received"
	EventOutboundMessageAccepted  EventType = "outbound_message_accepted"
	EventMessageSent              EventType = "message_sent"
	EventMessageDelivered         EventType = "message_delivered"
	EventMessageRead              EventType = "message_read"
	EventMessageFailed            EventType = "message_failed"
	EventWebhookQueued            EventType = "webhook_queued"
	EventWebhookDelivered         EventType = "webhook_delivered"
	EventWebhookDeliveryFailed    EventType = "webhook_delivery_failed"
	EventVirtualIdentityCreated   EventType = "virtual_identity_created"
	EventBusinessCreated          EventType = "business_created"
	EventWABACreated              EventType = "waba_created"
	EventBusinessPhoneProvisioned EventType = "business_phone_provisioned"
	EventApplicationCreated       EventType = "application_created"
	EventCredentialIssued         EventType = "credential_issued"
	EventCredentialRevoked        EventType = "credential_revoked"
	EventWebhookConfigured        EventType = "webhook_configured"
	EventMediaUploaded            EventType = "media_uploaded"
	EventMediaDownloaded          EventType = "media_downloaded"

	SourceMessaging    EventSource = "messaging"
	SourceAdmin        EventSource = "admin"
	SourceWebhook      EventSource = "webhook"
	SourceStatusSim    EventSource = "status_simulator"
	SourceMedia        EventSource = "media"
	SourceHTTP         EventSource = "http"
)

type EventOutcome string

const (
	EventOutcomeSuccess EventOutcome = "success"
	EventOutcomeFailure EventOutcome = "failure"
	EventOutcomePending EventOutcome = "pending"
)

type DomainEvent struct {
	ID             EventID      `json:"id"`
	Type           EventType    `json:"type"`
	Payload        any          `json:"payload"`
	Timestamp      time.Time    `json:"timestamp"`
	TraceID        TraceID      `json:"trace_id,omitempty"`
	Source         EventSource  `json:"source,omitempty"`
	ActorID        string       `json:"actor_id,omitempty"`
	ConversationID string       `json:"conversation_id,omitempty"`
	MessageID      string       `json:"message_id,omitempty"`
	ApplicationID  string       `json:"application_id,omitempty"`
	Outcome        EventOutcome `json:"outcome,omitempty"`
	DurationMs     int64        `json:"duration_ms,omitempty"`
}

type EventBuilder struct {
	event *DomainEvent
}

func NewEvent(eventType EventType) *EventBuilder {
	return &EventBuilder{
		event: &DomainEvent{
			ID:        EventID(GenerateMessageID()),
			Type:      eventType,
			Timestamp: time.Now(),
			Outcome:   EventOutcomeSuccess,
		},
	}
}

func (b *EventBuilder) WithTraceID(traceID TraceID) *EventBuilder {
	b.event.TraceID = traceID
	return b
}

func (b *EventBuilder) WithSource(source EventSource) *EventBuilder {
	b.event.Source = source
	return b
}

func (b *EventBuilder) WithActorID(actorID string) *EventBuilder {
	b.event.ActorID = actorID
	return b
}

func (b *EventBuilder) WithConversationID(convID string) *EventBuilder {
	b.event.ConversationID = convID
	return b
}

func (b *EventBuilder) WithMessageID(msgID string) *EventBuilder {
	b.event.MessageID = msgID
	return b
}

func (b *EventBuilder) WithApplicationID(appID string) *EventBuilder {
	b.event.ApplicationID = appID
	return b
}

func (b *EventBuilder) WithOutcome(outcome EventOutcome) *EventBuilder {
	b.event.Outcome = outcome
	return b
}

func (b *EventBuilder) WithDuration(d time.Duration) *EventBuilder {
	b.event.DurationMs = d.Milliseconds()
	return b
}

func (b *EventBuilder) WithPayload(payload any) *EventBuilder {
	b.event.Payload = payload
	return b
}

func (b *EventBuilder) Build() *DomainEvent {
	return b.event
}

func NewDomainEvent(eventType EventType, payload any) *DomainEvent {
	return NewEvent(eventType).WithPayload(payload).Build()
}

func GenerateTraceID() TraceID {
	return TraceID(GenerateMessageID())
}
