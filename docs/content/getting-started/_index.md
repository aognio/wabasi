---
title: Getting Started
weight: 2
---

# Getting Started

## Installation

```bash
git clone https://github.com/wabasi/wabasi.git
cd wabasi
make build
```

## Start the Server

```bash
./bin/wabasi
# Or: make run
```

Wabasi starts at `http://localhost:8080`.

## First Login

Open `http://localhost:8080/admin` in your browser. You'll see the admin dashboard with counts for Businesses, WABAs, Phone Numbers, and Applications.

## Path A: Alice and Bob

The simplest demonstration — two consumers exchanging messages.

### 1. Create Alice

```bash
curl -X POST http://localhost:8080/api/identities \
  -H "Content-Type: application/json" \
  -d '{"phone": "+51987654321", "display_name": "Alice"}'
```

Note the `device_token` in the response.

### 2. Create Bob

```bash
curl -X POST http://localhost:8080/api/identities \
  -H "Content-Type: application/json" \
  -d '{"phone": "+14155550100", "display_name": "Bob"}'
```

### 3. Open Alice's Device

Navigate to `http://localhost:8080/device/{alice_device_token}`.

### 4. Start a Conversation

Select Bob from the participant list and send a message.

### 5. Open Bob's Device

Navigate to `http://localhost:8080/device/{bob_device_token}`. Bob sees Alice's message and can reply.

**No business entities, API tokens, or webhooks required.**

## Path B: Business Integration

Connect an external CRM to Wabasi.

### 1. Create Business Entities

Via the admin UI at `/admin`:

1. Create a Business (e.g., "ACME Corp")
2. Create a WABA under that business
3. Provision a phone number (e.g., "+5115550100")

### 2. Create Application and Credentials

1. Create an Application under the business
2. Issue a credential — **save the token**, it's shown only once

### 3. Configure Webhook

Set your application's webhook URL to `http://localhost:8080/webhook`.

### 4. Send a Message

```bash
curl -X POST http://localhost:8080/v18.0/{phone_number_id}/messages \
  -H "Authorization: Bearer {your_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "messaging_product": "whatsapp",
    "to": "+51987654321",
    "type": "text",
    "text": {"body": "Hello from ACME!"}
  }'
```

### 5. Observe

- Check `/admin/reports/events` for the event log
- Check `/admin/reports/overview` for system metrics
- If a CRM is connected, it receives the webhook

## Path C: Investigation

After generating traffic, investigate what happened:

1. Open `/admin/reports` for system overview
2. Check the event log for message lifecycle
3. Use MCP for deep analysis (see [MCP documentation]({{< relref "mcp" >}}))
