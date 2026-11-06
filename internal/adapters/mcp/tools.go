package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wabasi/wabasi/internal/domain"
	"github.com/wabasi/wabasi/internal/ports"
)

func (s *MCPServer) registerTools() {
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "trace_message",
		Description: "Reconstruct the full causal lifecycle of a message. READ ONLY.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"message_id": map[string]any{
					"type":        "string",
					"description": "The message ID to trace",
				},
			},
			Required: []string{"message_id"},
		},
	}, s.handleTraceMessage)

	s.mcpServer.AddTool(mcp.Tool{
		Name:        "trace_conversation",
		Description: "Reconstruct event timeline for a conversation. READ ONLY.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"conversation_id": map[string]any{
					"type":        "string",
					"description": "The conversation ID to trace",
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Max events to return (default 50)",
				},
			},
			Required: []string{"conversation_id"},
		},
	}, s.handleTraceConversation)

	s.mcpServer.AddTool(mcp.Tool{
		Name:        "search_events",
		Description: "Search events with filters. Returns bounded results. READ ONLY.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"event_type": map[string]any{
					"type":        "string",
					"description": "Filter by event type",
				},
				"source": map[string]any{
					"type":        "string",
					"description": "Filter by source (messaging, admin, webhook, status_simulator, media, http)",
				},
				"outcome": map[string]any{
					"type":        "string",
					"description": "Filter by outcome (success, failure, pending)",
				},
				"application_id": map[string]any{
					"type":        "string",
					"description": "Filter by application ID",
				},
				"conversation_id": map[string]any{
					"type":        "string",
					"description": "Filter by conversation ID",
				},
				"since": map[string]any{
					"type":        "string",
					"description": "Start time (RFC3339 or duration like '1h', '24h')",
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Max results (default 50, max 200)",
				},
			},
		},
	}, s.handleSearchEvents)

	s.mcpServer.AddTool(mcp.Tool{
		Name:        "search_failures",
		Description: "Search for failed events. Returns bounded results with context. READ ONLY.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"since": map[string]any{
					"type":        "string",
					"description": "Start time (RFC3339 or duration like '1h', '24h')",
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Max results (default 50)",
				},
			},
		},
	}, s.handleSearchFailures)

	s.mcpServer.AddTool(mcp.Tool{
		Name:        "analyze_application",
		Description: "Summarize an Application's activity over a time range: traffic, messages, events. READ ONLY.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"application_id": map[string]any{
					"type":        "string",
					"description": "The application ID to analyze",
				},
				"since": map[string]any{
					"type":        "string",
					"description": "Start time (RFC3339 or duration like '1h', '24h')",
				},
			},
			Required: []string{"application_id"},
		},
	}, s.handleAnalyzeApplication)

	s.mcpServer.AddTool(mcp.Tool{
		Name:        "compare_applications",
		Description: "Compare two Applications' performance metrics. Returns sample counts. READ ONLY.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"application_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "List of application IDs to compare (2+)",
				},
				"since": map[string]any{
					"type":        "string",
					"description": "Start time (RFC3339 or duration like '1h', '24h')",
				},
			},
			Required: []string{"application_ids"},
		},
	}, s.handleCompareApplications)

	s.mcpServer.AddTool(mcp.Tool{
		Name:        "analyze_webhook_health",
		Description: "Analyze webhook delivery health for an Application. READ ONLY.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"application_id": map[string]any{
					"type":        "string",
					"description": "The application ID to analyze",
				},
				"since": map[string]any{
					"type":        "string",
					"description": "Start time (RFC3339 or duration like '1h', '24h')",
				},
			},
			Required: []string{"application_id"},
		},
	}, s.handleAnalyzeWebhookHealth)

	s.mcpServer.AddTool(mcp.Tool{
		Name:        "audit_world_integrity",
		Description: "Run integrity checks across the Wabasi world: participants, conversations, messages, delivery states. Returns structured findings. READ ONLY.",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, s.handleAuditWorldIntegrity)

	s.mcpServer.AddTool(mcp.Tool{
		Name:        "inspect_message",
		Description: "Get full message details including participants, status, timestamps. READ ONLY.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"message_id": map[string]any{
					"type":        "string",
					"description": "The message ID to inspect",
				},
			},
			Required: []string{"message_id"},
		},
	}, s.handleInspectMessage)

	s.mcpServer.AddTool(mcp.Tool{
		Name:        "list_participants",
		Description: "List all virtual identities and business phone numbers. READ ONLY.",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, s.handleListParticipants)
}

func (s *MCPServer) handleTraceMessage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	messageID := stringArg(args, "message_id")
	if messageID == "" {
		return errResult("message_id is required"), nil
	}

	timeline, err := s.obs.GetMessageTimeline(ctx, messageID)
	if err != nil {
		return errResult(fmt.Sprintf("error: %v", err)), nil
	}

	return jsonResult(timeline), nil
}

func (s *MCPServer) handleTraceConversation(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	convID := stringArg(args, "conversation_id")
	if convID == "" {
		return errResult("conversation_id is required"), nil
	}
	limit := 50
	if v, ok := args["limit"].(float64); ok && v > 0 {
		limit = int(v)
		if limit > 200 {
			limit = 200
		}
	}

	timeline, err := s.obs.GetConversationTimeline(ctx, convID, limit)
	if err != nil {
		return errResult(fmt.Sprintf("error: %v", err)), nil
	}

	return jsonResult(timeline), nil
}

func (s *MCPServer) handleSearchEvents(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	filter := ports.EventFilter{
		Limit: 50,
	}

	if et := stringArg(args, "event_type"); et != "" {
		filter.Types = []domain.EventType{domain.EventType(et)}
	}
	if src := stringArg(args, "source"); src != "" {
		filter.Source = domain.EventSource(src)
	}
	if out := stringArg(args, "outcome"); out != "" {
		filter.Outcome = domain.EventOutcome(out)
	}
	if appID := stringArg(args, "application_id"); appID != "" {
		filter.ApplicationID = appID
	}
	if convID := stringArg(args, "conversation_id"); convID != "" {
		filter.ConversationID = convID
	}
	if since := parseTimeArg(args["since"], 24*time.Hour); since != nil {
		filter.Since = since
	}
	if v, ok := args["limit"].(float64); ok && v > 0 {
		filter.Limit = int(v)
		if filter.Limit > 200 {
			filter.Limit = 200
		}
	}

	result, err := s.obs.ListEvents(ctx, filter)
	if err != nil {
		return errResult(fmt.Sprintf("error: %v", err)), nil
	}

	return jsonResult(result), nil
}

func (s *MCPServer) handleSearchFailures(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	filter := ports.EventFilter{
		Outcome: domain.EventOutcomeFailure,
		Limit:   50,
	}

	if since := parseTimeArg(args["since"], 24*time.Hour); since != nil {
		filter.Since = since
	}
	if v, ok := args["limit"].(float64); ok && v > 0 {
		filter.Limit = int(v)
		if filter.Limit > 200 {
			filter.Limit = 200
		}
	}

	result, err := s.obs.ListEvents(ctx, filter)
	if err != nil {
		return errResult(fmt.Sprintf("error: %v", err)), nil
	}

	return jsonResult(result), nil
}

func (s *MCPServer) handleAnalyzeApplication(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	appID := stringArg(args, "application_id")
	if appID == "" {
		return errResult("application_id is required"), nil
	}

	since := parseTimeArg(args["since"], 24*time.Hour)

	events, err := s.events.GetByApplicationID(ctx, appID, since, 500)
	if err != nil {
		return errResult(fmt.Sprintf("error: %v", err)), nil
	}

	byType := make(map[domain.EventType]int)
	byOutcome := make(map[domain.EventOutcome]int)
	for _, ev := range events {
		byType[ev.Type]++
		byOutcome[ev.Outcome]++
	}

	total := len(events)
	failures := byOutcome[domain.EventOutcomeFailure]
	var failureRate float64
	if total > 0 {
		failureRate = float64(failures) / float64(total) * 100
	}

	analysis := map[string]any{
		"application_id":  appID,
		"period":          fmt.Sprintf("since %s", since.Format(time.RFC3339)),
		"total_events":    total,
		"failures":        failures,
		"failure_rate":    fmt.Sprintf("%.1f%%", failureRate),
		"events_by_type":  byType,
		"events_by_outcome": byOutcome,
	}

	return jsonResult(analysis), nil
}

func (s *MCPServer) handleCompareApplications(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	idsRaw, ok := args["application_ids"].([]any)
	if !ok || len(idsRaw) < 2 {
		return errResult("application_ids must contain at least 2 IDs"), nil
	}

	ids := make([]string, len(idsRaw))
	for i, v := range idsRaw {
		ids[i] = fmt.Sprintf("%v", v)
	}

	since := parseTimeArg(args["since"], 24*time.Hour)

	type appMetrics struct {
		ApplicationID string         `json:"application_id"`
		TotalEvents   int            `json:"total_events"`
		Failures      int            `json:"failures"`
		FailureRate   string         `json:"failure_rate"`
		EventsByType  map[string]int `json:"events_by_type"`
	}

	results := make([]appMetrics, 0, len(ids))
	for _, appID := range ids {
		events, err := s.events.GetByApplicationID(ctx, appID, since, 500)
		if err != nil {
			continue
		}

		byType := make(map[string]int)
		failures := 0
		for _, ev := range events {
			byType[string(ev.Type)]++
			if ev.Outcome == domain.EventOutcomeFailure {
				failures++
			}
		}

		total := len(events)
		var failureRate float64
		if total > 0 {
			failureRate = float64(failures) / float64(total) * 100
		}

		results = append(results, appMetrics{
			ApplicationID: appID,
			TotalEvents:   total,
			Failures:      failures,
			FailureRate:   fmt.Sprintf("%.1f%%", failureRate),
			EventsByType:  byType,
		})
	}

	return jsonResult(map[string]any{
		"period":        fmt.Sprintf("since %s", since.Format(time.RFC3339)),
		"comparisons":   results,
		"sample_note":   "Compare failure rates only with sufficient sample sizes.",
	}), nil
}

func (s *MCPServer) handleAnalyzeWebhookHealth(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	appID := stringArg(args, "application_id")
	if appID == "" {
		return errResult("application_id is required"), nil
	}

	since := parseTimeArg(args["since"], 24*time.Hour)

	webhookEvents, err := s.events.List(ctx, ports.EventFilter{
		Types:         []domain.EventType{domain.EventWebhookQueued, domain.EventWebhookDelivered, domain.EventWebhookDeliveryFailed},
		ApplicationID: appID,
		Since:         since,
		Limit:         500,
	})
	if err != nil {
		return errResult(fmt.Sprintf("error: %v", err)), nil
	}

	queued := 0
	delivered := 0
	failed := 0
	for _, ev := range webhookEvents {
		switch ev.Type {
		case domain.EventWebhookQueued:
			queued++
		case domain.EventWebhookDelivered:
			delivered++
		case domain.EventWebhookDeliveryFailed:
			failed++
		}
	}

	total := delivered + failed
	var successRate float64
	if total > 0 {
		successRate = float64(delivered) / float64(total) * 100
	}

	return jsonResult(map[string]any{
		"application_id":     appID,
		"period":             fmt.Sprintf("since %s", since.Format(time.RFC3339)),
		"webhooks_queued":    queued,
		"webhooks_delivered": delivered,
		"webhooks_failed":    failed,
		"success_rate":       fmt.Sprintf("%.1f%%", successRate),
	}), nil
}

func (s *MCPServer) handleAuditWorldIntegrity(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type finding struct {
		Severity string   `json:"severity"`
		Category string   `json:"category"`
		Summary  string   `json:"summary"`
	RelatedIDs []string `json:"related_ids,omitempty"`
	}

	var findings []finding

	identities, err := s.identities.ListIdentities(ctx)
	if err != nil {
		findings = append(findings, finding{
			Severity: "ERROR",
			Category: "participants",
			Summary:  fmt.Sprintf("Failed to list identities: %v", err),
		})
	} else {
		findings = append(findings, finding{
			Severity: "PASS",
			Category: "participants",
			Summary:  fmt.Sprintf("%d virtual identities registered", len(identities)),
		})
	}

	events24h, err := s.events.Count(ctx, ports.EventFilter{
		Since: func() *time.Time { t := time.Now().Add(-24 * time.Hour); return &t }(),
	})
	if err != nil {
		findings = append(findings, finding{
			Severity: "WARN",
			Category: "events",
			Summary:  fmt.Sprintf("Failed to count events: %v", err),
		})
	} else {
		findings = append(findings, finding{
			Severity: "PASS",
			Category: "events",
			Summary:  fmt.Sprintf("%d events in last 24h", events24h),
		})
	}

	failures, err := s.events.List(ctx, ports.EventFilter{
		Outcome: domain.EventOutcomeFailure,
		Since:   func() *time.Time { t := time.Now().Add(-24 * time.Hour); return &t }(),
		Limit:   10,
	})
	if err == nil && len(failures) > 0 {
		ids := make([]string, 0, len(failures))
		for _, f := range failures {
			ids = append(ids, string(f.ID))
		}
		findings = append(findings, finding{
			Severity: "WARN",
			Category: "failures",
			Summary:  fmt.Sprintf("%d failures in last 24h", len(failures)),
			RelatedIDs: ids,
		})
	} else if err == nil {
		findings = append(findings, finding{
			Severity: "PASS",
			Category: "failures",
			Summary:  "No failures in last 24h",
		})
	}

	return jsonResult(map[string]any{
		"audit_time": time.Now().UTC(),
		"findings":   findings,
		"total":      len(findings),
	}), nil
}

func (s *MCPServer) handleInspectMessage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	messageID := stringArg(args, "message_id")
	if messageID == "" {
		return errResult("message_id is required"), nil
	}

	msg, err := s.messages.GetMessageByID(ctx, domain.MessageID(messageID))
	if err != nil {
		return errResult(fmt.Sprintf("message not found: %v", err)), nil
	}

	return jsonResult(msg), nil
}

func (s *MCPServer) handleListParticipants(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	identities, err := s.identities.ListIdentities(ctx)
	if err != nil {
		return errResult(fmt.Sprintf("error: %v", err)), nil
	}

	participants := make([]map[string]any, 0, len(identities))
	for _, id := range identities {
		participants = append(participants, map[string]any{
			"id":           string(id.ID),
			"type":         "consumer",
			"phone":        string(id.Phone),
			"display_name": id.DisplayName,
		})
	}

	return jsonResult(map[string]any{
		"participants": participants,
		"total":        len(participants),
	}), nil
}
