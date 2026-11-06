package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *MCPServer) registerPrompts() {
	s.mcpServer.AddPrompt(mcp.Prompt{
		Name:        "investigate_message_failure",
		Description: "Guide investigation of why a message failed or was not delivered",
		Arguments: []mcp.PromptArgument{
			{Name: "message_id", Description: "The message ID to investigate", Required: true},
		},
	}, s.handleInvestigateMessageFailure)

	s.mcpServer.AddPrompt(mcp.Prompt{
		Name:        "investigate_conversation",
		Description: "Guide analysis of a conversation's full lifecycle",
		Arguments: []mcp.PromptArgument{
			{Name: "conversation_id", Description: "The conversation ID to investigate", Required: true},
		},
	}, s.handleInvestigateConversation)

	s.mcpServer.AddPrompt(mcp.Prompt{
		Name:        "compare_applications",
		Description: "Guide comparison of two or more business Applications",
		Arguments: []mcp.PromptArgument{
			{Name: "application_ids", Description: "Comma-separated application IDs to compare", Required: true},
		},
	}, s.handleCompareApplicationsPrompt)

	s.mcpServer.AddPrompt(mcp.Prompt{
		Name:        "audit_world",
		Description: "Guide a comprehensive audit of the Wabasi world",
		Arguments:  []mcp.PromptArgument{},
	}, s.handleAuditWorldPrompt)

	s.mcpServer.AddPrompt(mcp.Prompt{
		Name:        "explain_incident",
		Description: "Guide reconstruction of what happened during a time period",
		Arguments: []mcp.PromptArgument{
			{Name: "time_range", Description: "Time range to investigate (e.g. '1h', '24h')", Required: true},
		},
	}, s.handleExplainIncident)
}

func textMessage(text string) mcp.PromptMessage {
	return mcp.PromptMessage{
		Role:    "user",
		Content: mcp.TextContent{Type: "text", Text: text},
	}
}

func (s *MCPServer) handleInvestigateMessageFailure(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	messageID := request.Params.Arguments["message_id"]

	guide := `## Investigate Message Failure: ` + messageID + `

Follow this structured investigation:

1. **Inspect the message**
   - Use inspect_message with message_id="` + messageID + `"
   - Note: sender, recipient, type, status, timestamps, media references

2. **Reconstruct the trace**
   - Use trace_message with message_id="` + messageID + `"
   - Note each step: creation, delivery attempts, status transitions

3. **Check for related failures**
   - Use search_failures to find failures around the same time
   - Filter by conversation_id if available

4. **Analyze delivery path**
   - If business message: check webhook delivery health
   - If consumer message: check delivery state transitions

5. **Separate evidence from hypothesis**
   - List only confirmed facts from the trace
   - Mark assumptions explicitly

6. **Identify root cause**
   - Was the failure in transport, application, or state transition?
   - What exact component/step failed?

Report your findings with evidence references (message_id, trace_id, event_id).`

	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{textMessage(guide)},
	}, nil
}

func (s *MCPServer) handleInvestigateConversation(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	convID := request.Params.Arguments["conversation_id"]

	guide := `## Investigate Conversation: ` + convID + `

Follow this structured analysis:

1. **List participants**
   - Use list_participants to identify all participants

2. **Trace conversation timeline**
   - Use trace_conversation with conversation_id="` + convID + `"
   - Note message flow, delivery states, time gaps

3. **Identify patterns**
   - Normal conversation flow vs anomalies
   - Delivery state progression (pending → sent → delivered → read)
   - Any stuck or failed messages

4. **Check for business integration**
   - Are there webhook events related to this conversation?
   - Any API requests from external applications?

5. **Measure latency**
   - Time between messages
   - Delivery/read latency where timestamps exist

6. **Report findings**
   - Conversation health: normal, degraded, failing
   - Key events with timestamps
   - Any anomalies detected`

	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{textMessage(guide)},
	}, nil
}

func (s *MCPServer) handleCompareApplicationsPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	ids := request.Params.Arguments["application_ids"]

	guide := `## Compare Applications: ` + ids + `

Follow this comparison methodology:

1. **Gather metrics for each Application**
   - Use analyze_application for each application_id
   - Note: total events, failures, failure rate, event type distribution

2. **Compare key dimensions**
   - Event volume (are they handling similar traffic?)
   - Failure rates (which is more reliable?)
   - Event type distribution (what operations are they performing?)

3. **Check webhook health**
   - Use analyze_webhook_health for each application
   - Compare delivery success rates

4. **Assess sample sizes**
   - WARN if comparing applications with very different sample sizes
   - Small samples may not be statistically meaningful

5. **Identify anomalies**
   - Which application has unusual patterns?
   - Any application showing degradation?

6. **Report with evidence**
   - Present metrics side by side
   - Include sample counts (n=X) for all comparisons
   - Do not declare a "winner" from incomparable data`

	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{textMessage(guide)},
	}, nil
}

func (s *MCPServer) handleAuditWorldPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	guide := `## Audit Wabasi World

Perform a comprehensive integrity audit:

1. **World summary**
   - Read wabasi://world/summary resource
   - Note: consumers, events, failures

2. **Run integrity checks**
   - Use audit_world_integrity
   - Review all findings by severity

3. **Check recent failures**
   - Use search_failures with since="1h"
   - Identify any patterns

4. **Review event distribution**
   - Read wabasi://metrics/current resource
   - Check for unusual event type ratios

5. **Assess overall health**
   - PASS: no findings or only PASS items
   - WARNING: WARN findings exist
   - CRITICAL: ERROR findings exist

6. **Report findings**
   - List each finding with severity, category, summary
   - Include related IDs for drill-down
   - Separate facts from assessments`

	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{textMessage(guide)},
	}, nil
}

func (s *MCPServer) handleExplainIncident(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	timeRange := request.Params.Arguments["time_range"]

	guide := `## Explain Incident: last ` + timeRange + `

Reconstruct what happened:

1. **Gather events in time range**
   - Use search_events with since="` + timeRange + `"
   - Note total volume and distribution

2. **Identify failures**
   - Use search_failures with since="` + timeRange + `"
   - Group by type, source, application

3. **Cluster related events**
   - Look for events sharing trace_id, conversation_id, or message_id
   - Identify causal chains

4. **Reconstruct timelines**
   - For each cluster, use trace_message or trace_conversation
   - Build causal timelines

5. **Separate evidence layers**
   - CONFIRMED: directly observed in events/traces
   - INFERRED: logically derived from confirmed facts
   - HYPOTHESIZED: possible explanations not yet confirmed

6. **Report**
   - What happened (confirmed facts)
   - How it happened (causal reconstruction)
   - Why it happened (inferred root cause, if supported by evidence)
   - What to investigate next (open questions)`

	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{textMessage(guide)},
	}, nil
}
