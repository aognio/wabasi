---
title: Observability
weight: 4
---

# Observability and Diagnostics

Wakaba deliberately records enough structured operational evidence to reconstruct simulated interactions. This is not a side effect of logging — it is a first-class architectural subsystem.

## Three Complementary Concepts

### Events

Events record meaningful facts that occurred:

- MessageCreated
- MessageDelivered
- MessageRead
- WebhookQueued
- WebhookDelivered
- MediaUploaded

Events answer: **What happened?**

### Traces

Traces reconstruct causal paths across Wakaba components:

```
Alice sends message
      ↓
message created
      ↓
ACME resolved
      ↓
webhook queued
      ↓
Tiny CRM receives webhook
      ↓
Tiny CRM POST /messages
      ↓
reply delivered to Alice
      ↓
Alice reads reply
```

Traces answer: **How did this happen?**

### Metrics

Metrics derive aggregate measurements from operational facts:

- Message counts
- Delivery latency
- First response latency
- Webhook failure rate
- CAS deduplication ratio

Metrics answer: **How often, how much, how fast?**

## Event Model

Every event carries correlation fields:

| Field | Purpose |
|-------|---------|
| TraceID | Groups related events into a causal chain |
| Source | Which component generated the event |
| ConversationID | Links to the conversation |
| MessageID | Links to the message |
| ApplicationID | Links to the business application |
| Outcome | success, failure, or pending |
| DurationMs | How long the operation took |

## Event Sources

- `messaging` — message creation, delivery
- `admin` — business entity management
- `webhook` — webhook generation and delivery
- `status_simulator` — automatic status transitions
- `media` — media upload/download
- `http` — HTTP request handling

## Diagnostic Findings

Diagnostics represent analytical conclusions supported by evidence:

- **PASS** — check passed
- **WARN** — potential issue detected
- **ERROR** — integrity problem found

Examples:

```
PASS 17 virtual identities registered
WARN 3 failures in last 24h
FAIL MediaObject references missing Blob
```

## Admin Reports

Access reports at `/admin/reports`:

- **System Overview** — world summary, event counts, failure rate
- **Event Log** — recent events with filtering by type, source, outcome
- **Metrics** — event distribution and outcome percentages

## Causal Timeline Example

```
12:00:00.114 Alice submitted message
12:00:00.118 Message msg_123 created
12:00:00.121 Conversation conv_9 resolved
12:00:00.125 Destination ACME Support resolved
12:00:00.132 Webhook event wh_77 queued
12:00:00.184 POST sent to Tiny CRM
12:00:00.371 Tiny CRM returned HTTP 200
12:00:01.901 Tiny CRM POST /messages
12:00:01.928 Reply msg_124 created
12:00:01.945 Reply delivered to Alice
12:00:02.630 Alice read reply
```

From these facts, Wakaba derives:

- Webhook transport latency
- CRM processing/response latency
- Business first-response latency
- End-to-end latency
