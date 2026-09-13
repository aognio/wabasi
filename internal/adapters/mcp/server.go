package mcp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/wakaba/wakaba/internal/application"
	"github.com/wakaba/wakaba/internal/ports"
)

type MCPServer struct {
	mcpServer *server.MCPServer
	obs       *application.ObservabilityService
	messages  ports.MessageRepository
	convs     ports.ConversationRepository
	identities ports.VirtualIdentityRepository
	events    ports.EventReader
}

func NewMCPServer(
	obs *application.ObservabilityService,
	messages ports.MessageRepository,
	convs ports.ConversationRepository,
	identities ports.VirtualIdentityRepository,
	events ports.EventReader,
) *MCPServer {
	s := &MCPServer{
		obs:        obs,
		messages:   messages,
		convs:      convs,
		identities: identities,
		events:     events,
	}

	mcpSrv := server.NewMCPServer(
		"wakaba",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
		server.WithPromptCapabilities(true),
	)

	s.mcpServer = mcpSrv
	s.registerResources()
	s.registerTools()
	s.registerPrompts()

	return s
}

func (s *MCPServer) Server() *server.MCPServer {
	return s.mcpServer
}

func (s *MCPServer) Serve_stdio() error {
	return server.ServeStdio(s.mcpServer)
}

func jsonResult(data any) *mcp.CallToolResult {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("%v", data)}},
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(b)}},
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: text}},
	}
}

func errResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: msg}},
		IsError: true,
	}
}

func parseTimeArg(val any, defaultDuration time.Duration) *time.Time {
	if val == nil {
		t := time.Now().Add(-defaultDuration)
		return &t
	}
	if s, ok := val.(string); ok {
		if s == "" {
			t := time.Now().Add(-defaultDuration)
			return &t
		}
		if d, err := time.ParseDuration(s); err == nil {
			t := time.Now().Add(-d)
			return &t
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return &t
		}
	}
	t := time.Now().Add(-defaultDuration)
	return &t
}

func stringArg(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
