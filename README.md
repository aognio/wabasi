# Wabasi

**WhatsApp Business API Simulator** — a locally controllable simulated messaging world for WhatsApp Business Platform development, testing, and investigation.

## What is Wabasi?

Wabasi is an open-source simulator and compatibility implementation of the WhatsApp Business Platform / Cloud API. It creates an isolated, deterministic messaging world where:

- **Consumer identities** communicate directly with each other
- **Business participants** receive messages through a Meta-compatible API
- **External applications** (CRMs, bots) connect via webhooks
- **Media** flows through content-addressable storage
- **Everything** is observable, traceable, and diagnosable

A WhatsApp-aware application switches from Meta's real API to Wabasi by changing only the API base URL and credentials.

## Quick Start

```bash
# Build and run
make build
./bin/wabasi

# Or with make
make run
```

Wabasi starts at `http://localhost:8080`.

### Path A: Consumer Messaging

1. Open `http://localhost:8080/admin` and create a Virtual Identity (e.g., Alice, +51987654321)
2. Create another identity (e.g., Bob, +14155550100)
3. Open Alice's device at `/device/{alice_token}`
4. Start a conversation with Bob and exchange messages

No business setup required.

### Path B: Business Integration

1. Create a Business, WABA, and provision a phone number via `/admin`
2. Create an Application and issue credentials
3. Configure a webhook endpoint
4. Connect your CRM using the Meta-compatible API:
   ```bash
   curl -X POST http://localhost:8080/v18.0/{phone_number_id}/messages \
     -H "Authorization: Bearer {token}" \
     -H "Content-Type: application/json" \
     -d '{"messaging_product":"whatsapp","to":"+51987654321","type":"text","text":{"body":"Hello!"}}'
   ```

## Architecture

```
                    Wabasi World

     Alice <──────────────> Bob
       │
       ↓
   ACME Support
       │
  Business Platform
       │
    Tiny CRM


  Observability Layer
       │
   ┌───┴────┐
 Reports    MCP
             │
        Coding Agent
```

### Key Concepts

| Concept | Description |
|---------|-------------|
| **VirtualIdentity** | Simulated consumer participant (Alice, Bob) |
| **BusinessPhoneNumber** | Business participant (ACME Support) |
| **Application** | External CRM authorized to use the API |
| **Conversation** | Message thread between two participants |
| **Message** | Individual message with status progression |
| **MediaObject** | Logical media resource backed by CAS |
| **Trace** | Causal reconstruction of a message lifecycle |

### Package Structure

```
internal/
    domain/          # Domain types, events, value objects
    ports/           # Interfaces (repositories, event publishers)
    application/     # Application services
    adapters/
        http/        # Fiber HTTP handlers, admin UI, reports
        sqlite/      # GORM repos, event storage
        mcp/         # MCP server for agentic investigation
        webhook/     # Webhook delivery worker
cmd/
    wabasi/          # Main executable
```

## MCP (Model Context Protocol)

Wabasi exposes its world through MCP for agentic debugging, tracing, and analysis:

```bash
# Start in MCP mode (stdio transport)
./bin/wabasi --mcp
```

### Resources
- `wabasi://world/summary` — world snapshot
- `wabasi://health` — server health
- `wabasi://metrics/current` — 24h metrics

### Tools (read-only)
- `trace_message` — reconstruct message lifecycle
- `trace_conversation` — conversation event timeline
- `search_events` — filtered event search
- `search_failures` — find failed events
- `analyze_application` — application performance summary
- `compare_applications` — side-by-side comparison
- `analyze_webhook_health` — webhook delivery analysis
- `audit_world_integrity` — world health check
- `inspect_message` — full message details
- `list_participants` — all identities

### Prompts
- `investigate_message_failure` — guided failure investigation
- `investigate_conversation` — conversation analysis
- `compare_applications` — application comparison
- `audit_world` — comprehensive audit
- `explain_incident` — time-range reconstruction

## Observability

Wabasi records structured operational evidence for every interaction:

- **Events** — discrete facts (message created, webhook delivered)
- **Traces** — causal relationships across components
- **Metrics** — aggregated measurements (failure rates, event distributions)

Admin reports at `/admin/reports`:
- System Overview — world summary with event counts
- Event Log — recent events with filtering
- Metrics — event distribution and outcomes

## API Compatibility

### Meta-Compatible Interface

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v18.0/{phone_number_id}/messages` | POST | Send message |
| `/webhook` | POST | Receive inbound messages/statuses |
| `/webhook` | GET | Webhook verification |

### Wabasi Control Interface

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/admin` | GET | Admin dashboard |
| `/admin/reports` | GET | Observability reports |
| `/api/identities` | POST | Create virtual identity |
| `/device/:token` | GET | Virtual device UI |
| `/media/` | POST | Upload media |
| `/media/:id` | GET | Download media |

## Configuration

| Flag/Env | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `--mcp` | false | Start in MCP stdio mode |
| (arg 1) | `var/wabasi.db` | SQLite database path |

## Development

```bash
make build     # Build binary
make test      # Run tests
make lint      # Run linter
make fmt       # Format code
make vet       # Run go vet
make clean     # Remove build artifacts
make help      # Show all targets
```

## License

MIT License - Copyright (c) 2026 Antonio Ognio

Made with ❤️ from 🇵🇪. El Perú es clave 🔑.
