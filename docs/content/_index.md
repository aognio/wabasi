---
title: Introduction
weight: 1
---

# Wakaba

Wakaba is a locally controllable simulated messaging world designed for WhatsApp Business Platform development, testing, and investigation.

## What Wakaba Provides

- **Simulated messaging world** with arbitrary consumer identities
- **Consumer-to-consumer communication** (Alice ↔ Bob)
- **Meta-compatible Business Platform** API for CRM integration
- **Media/file exchange** with content-addressable storage
- **Deterministic scenarios** for repeatable testing
- **Deep observability** with causal reconstruction
- **Agentic debugging** through MCP

## Why Wakaba?

Unlike Meta's official test facilities, Wakaba gives you **complete control and visibility** over the simulated world:

- Inspect every message, delivery transition, and webhook
- Trace causal paths across components
- Compare connected applications
- Inject and observe failures
- All locally, all deterministic, all observable

## Glass Box

Wakaba is a glass-box messaging laboratory. You can inspect:

- Participants and their relationships
- Messages and delivery state transitions
- Webhook delivery attempts and outcomes
- API requests and responses
- Media storage and deduplication
- Causal relationships across all components

This makes Wakaba valuable for:

- Automated testing and CI
- Agentic coding and debugging
- Protocol investigation
- Integration development
