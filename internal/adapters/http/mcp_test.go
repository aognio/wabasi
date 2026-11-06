package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	wabasiMCP "github.com/wabasi/wabasi/internal/adapters/mcp"
)

func TestMCPStreamableHTTPInitialize(t *testing.T) {
	server, _, cleanup := setupTestWithCredential(t)
	defer cleanup()

	mcpServer := wabasiMCP.NewMCPServer(nil, nil, nil, nil, nil)
	server.SetMCPTransport(mcpServer.StreamableHTTP())

	body := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"protocolVersion": "2025-03-26",
			"capabilities": {},
			"clientInfo": {"name": "wabasi-test", "version": "1.0.0"}
		}
	}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := server.App().Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["jsonrpc"] != "2.0" {
		t.Fatalf("expected JSON-RPC 2.0 response, got %#v", payload)
	}
	result, ok := payload["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected initialize result, got %#v", payload)
	}
	if result["protocolVersion"] == nil {
		t.Fatalf("expected protocolVersion in initialize result, got %#v", result)
	}
}
