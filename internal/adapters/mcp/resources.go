package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *MCPServer) registerResources() {
	s.mcpServer.AddResource(
		mcp.Resource{
			URI:         "wakaba://world/summary",
			Name:        "world_summary",
			Description: "Current Wakaba world snapshot: consumers, events, failures",
			MIMEType:    "application/json",
		},
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			summary, err := s.obs.GetWorldSummary(ctx)
			if err != nil {
				return nil, fmt.Errorf("get world summary: %w", err)
			}
			b, _ := json.MarshalIndent(summary, "", "  ")
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(b),
				},
			}, nil
		},
	)

	s.mcpServer.AddResource(
		mcp.Resource{
			URI:         "wakaba://health",
			Name:        "health",
			Description: "Wakaba server health status",
			MIMEType:    "application/json",
		},
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			health := map[string]any{
				"status":    "healthy",
				"timestamp": time.Now().UTC(),
				"version":   "1.0.0",
			}
			b, _ := json.MarshalIndent(health, "", "  ")
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(b),
				},
			}, nil
		},
	)

	s.mcpServer.AddResource(
		mcp.Resource{
			URI:         "wakaba://metrics/current",
			Name:        "current_metrics",
			Description: "Current metrics summary (last 24h): event distribution, failure rate",
			MIMEType:    "application/json",
		},
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			summary, err := s.obs.GetMetricsSummary(ctx, 24*time.Hour)
			if err != nil {
				return nil, fmt.Errorf("get metrics: %w", err)
			}
			b, _ := json.MarshalIndent(summary, "", "  ")
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(b),
				},
			}, nil
		},
	)
}
