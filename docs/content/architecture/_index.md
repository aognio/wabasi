---
title: Architecture
weight: 6
---

# Architecture

## Overview

Wabasi follows clean/hexagonal architecture with ports and adapters:

```
internal/
    domain/          # Domain types, events, value objects
    ports/           # Interfaces (repositories, publishers)
    application/     # Application services
    adapters/
        http/        # Fiber HTTP, admin UI, reports
        sqlite/      # GORM repos, event storage
        mcp/         # MCP server
        webhook/     # Delivery worker
cmd/
    wabasi/          # Main executable
```

**Dependencies point inward.** HTTP, SQLite, MCP, and the webhook worker are adapters around the Application and Domain layers. The domain knows nothing about Fiber, GORM, SQLite, HTML, or MCP.

## Domain Layer

The domain layer contains pure business types and rules:

- **Value Objects**: MessageID, PhoneNumber, WhatsAppID, Timestamp
- **Entities**: Message, Conversation, VirtualIdentity, Business, Application
- **Events**: DomainEvent with trace correlation
- **Errors**: Meta-compatible error envelope helpers

No external dependencies.

## Ports Layer

Interfaces that the application layer depends on:

```go
type MessageRepository interface {
    CreateMessage(ctx, msg) error
    GetMessageByID(ctx, id) (*Message, error)
    // ...
}

type EventPublisher interface {
    Publish(ctx, event) error
}

type EventReader interface {
    GetByID(ctx, id) (*DomainEvent, error)
    List(ctx, filter) ([]*DomainEvent, error)
    // ...
}
```

Implementations live in the adapters layer.

## Application Layer

Services that orchestrate domain logic:

- **MessagingService** — send/receive messages, update status
- **AdminService** — CRUD for business entities, credentials, webhooks
- **PhoneRegistry** — global phone number allocation
- **StatusSimulator** — automatic status progression
- **ObservabilityService** — event queries, traces, metrics

## Adapter Layer

### HTTP Adapter (Fiber)

Two API surfaces:

1. **Meta-compatible interface** — `/v18.0/{phone_number_id}/messages`
2. **Wabasi control interface** — `/admin/`, `/api/identities`, `/device/`, `/media/`

Bearer token auth middleware validates tokens against the credential chain.

### SQLite Adapter (GORM)

Repository implementations for all domain entities. Auto-migration on startup.

### MCP Adapter

MCP server exposing resources, tools, and prompts over stdio. Uses the `mcp-go` SDK.

### Webhook Adapter

Background delivery worker with:
- X-Hub-Signature-256 signing
- Exponential backoff retries
- Configurable max attempts

## Event Flow

```
Operation occurs
      ↓
DomainEvent created with TraceID, Source, correlation IDs
      ↓
EventPublisher persists to SQLite
      ↓
EventReader can query for traces, timelines, metrics
      ↓
ObservabilityService derives summaries
      ↓
Admin Reports and MCP expose the same data
```

## Message Lifecycle

```
SendMessage called
      ↓
Message created (status: pending)
      ↓
MessageStore persisted
      ↓
Event published (outbound_message_accepted)
      ↓
StatusSimulator advances: sent → delivered → read
      ↓
Events published at each transition
      ↓
Webhooks queued for business traffic
      ↓
WebhookDeliveryWorker sends with HMAC signing
```

## Phone Number Uniqueness

```
Global Phone Registry
        │
   ┌────┴────┐
   ↓         ↓
Consumer   Business
Numbers    Numbers
   │         │
   └────┬────┘
        ↓
  E.164 Normalized
  Uniqueness Enforced
```

A phone number belongs to exactly one participant across the entire system.

## Media Storage

```
Upload → SHA-256 stream → Deduplication → Blob (immutable)
                                              ↓
                                        MediaObject (logical)
                                              ↓
                                    WhatsApp-compatible API
```

CAS is internal implementation. Integrations see only Media IDs.
