package domain

import "time"

type EventID string
type EventType string

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
)

type DomainEvent struct {
	ID        EventID   `json:"id"`
	Type      EventType `json:"type"`
	Payload   any       `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

func NewDomainEvent(eventType EventType, payload any) *DomainEvent {
	return &DomainEvent{
		ID:        EventID(GenerateMessageID()),
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}
