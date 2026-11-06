---
title: Introduction
weight: 1
---

# Wabasi

**WhatsApp Business API Simulator** — a locally controllable simulated messaging world designed for WhatsApp Business Platform development, testing, and investigation.

## What Wabasi Provides

- **Simulated messaging world** with arbitrary consumer identities
- **Consumer-to-consumer communication** (Alice ↔ Bob)
- **Meta-compatible Business Platform** API for CRM integration
- **Media/file exchange** with content-addressable storage
- **Deterministic scenarios** for repeatable testing
- **Deep observability** with causal reconstruction
- **Agentic debugging** through MCP

## Why Wabasi?

Unlike Meta's official test facilities, Wabasi gives you **complete control and visibility** over the simulated world:

- Inspect every message, delivery transition, and webhook
- Trace causal paths across components
- Compare connected applications
- Inject and observe failures
- All locally, all deterministic, all observable

## Glass Box

Wabasi is a glass-box messaging laboratory. You can inspect:

- Participants and their relationships
- Messages and delivery state transitions
- Webhook delivery attempts and outcomes
- API requests and responses
- Media storage and deduplication
- Causal relationships across all components

This makes Wabasi valuable for:

- Automated testing and CI
- Agentic coding and debugging
- Protocol investigation
- Integration development
