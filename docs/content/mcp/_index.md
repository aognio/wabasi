---
title: MCP (Model Context Protocol)
weight: 5
---

# MCP for Agentic Investigation

Wakaba exposes its world through MCP for agentic debugging, tracing, analytics, and deep investigation.

> **MCP is primarily a privileged observability, tracing, analytics and diagnostics interface for developers and agentic coding/investigation tools.**

## What MCP Is Not

MCP is NOT:

- A CRM integration API (use the Meta-compatible API)
- A replacement for the Business API
- A Virtual Client protocol
- Required for normal operation

CRM applications use the Meta-compatible API. Human users use the Virtual Client. MCP is for engineering analysis.

## Starting MCP Mode

```bash
# Start Wakaba in MCP stdio mode
./var/wakaba --mcp
```

MCP communicates over stdin/stdout using the JSON-RPC protocol.

## Architecture

```
Observability / Analysis Layer
        │
   ┌────┴────┐
   ↓         ↓
Reports     MCP
             ↓
        Coding Agent
```

MCP and web reports consume the same analytical services. There is one definition of every metric.

## Resources

Stable, readable representations of Wakaba state:

| URI | Description |
|-----|-------------|
| `wakaba://world/summary` | Consumer count, event totals, failure rate |
| `wakaba://health` | Server health status |
| `wakaba://metrics/current` | 24h event distribution and outcomes |

## Tools

All tools are **read-only** unless marked otherwise.

### Tracing

| Tool | Description |
|------|-------------|
| `trace_message` | Reconstruct full message lifecycle |
| `trace_conversation` | Conversation event timeline |
| `inspect_message` | Full message details |

### Search

| Tool | Description |
|------|-------------|
| `search_events` | Filtered event search (type, source, outcome, time range) |
| `search_failures` | Find failed events |
| `list_participants` | All virtual identities |

### Analysis

| Tool | Description |
|------|-------------|
| `analyze_application` | Application performance summary |
| `compare_applications` | Side-by-side comparison with sample counts |
| `analyze_webhook_health` | Webhook delivery analysis |

### Integrity

| Tool | Description |
|------|-------------|
| `audit_world_integrity` | Run integrity checks, return findings |

## Prompts

Guided investigative workflows:

| Prompt | Purpose |
|--------|---------|
| `investigate_message_failure` | Step-by-step failure investigation |
| `investigate_conversation` | Conversation lifecycle analysis |
| `compare_applications` | Application comparison methodology |
| `audit_world` | Comprehensive world audit |
| `explain_incident` | Time-range incident reconstruction |

## Investigation Workflow

```
Agent changes code
      ↓
runs tests
      ↓
test fails
      ↓
reads world summary (resource)
      ↓
searches failures (tool)
      ↓
identifies message msg_123
      ↓
trace_message(msg_123) (tool)
      ↓
discovers webhook attempt returned 500
      ↓
inspects webhook attempt
      ↓
fixes implementation
      ↓
reruns test
      ↓
trace now succeeds
```

## Evidence-Driven Investigation

MCP returns structured evidence, not prose. The agent reasons over:

1. **Observed fact** — directly from events/traces
2. **Inference** — logically derived from facts
3. **Hypothesis** — possible explanations not yet confirmed

This separation is critical for reliable debugging.

## Security

MCP is a privileged diagnostic interface. It does NOT expose:

- Bearer tokens
- Password hashes
- App Secrets
- Verify Tokens
- Session cookies

Message content may be visible for debugging purposes. Media bytes are never exposed — only identifiers and metadata.

## Context Efficiency

MCP results favor:

- Summaries and aggregates
- Top anomalies
- Stable IDs for drill-down
- Bounded result sets (default 50, max 200)

Rather than dumping complete logs into context. This is part of the agent-first design.
