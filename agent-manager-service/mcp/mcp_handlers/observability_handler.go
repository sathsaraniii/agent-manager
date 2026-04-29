package mcp_handlers

import (
	"context"
	"fmt"
	"time"

	traceobserversvc "github.com/wso2/agent-manager/agent-manager-service/clients/traceobserversvc"
	"github.com/wso2/agent-manager/agent-manager-service/models"
	"github.com/wso2/agent-manager/agent-manager-service/services"
	"github.com/wso2/agent-manager/agent-manager-service/spec"
)

// ObserverHandler bridges MCP observer tools (logs/metrics) to the agent manager service layer.
type ObserverHandler struct {
	agentSvc services.AgentManagerService
}

func NewObserverHandler(agentSvc services.AgentManagerService) *ObserverHandler {
	return &ObserverHandler{agentSvc: agentSvc}
}

func (h *ObserverHandler) GetRuntimeLogs(ctx context.Context, orgName string, projectName string, agentName string, payload spec.LogFilterRequest) (*models.LogsResponse, error) {
	return h.agentSvc.GetAgentRuntimeLogs(ctx, orgName, projectName, agentName, payload)
}

// RuntimeLogHandler bridges MCP runtime log tools to the agent manager service layer.
type RuntimeLogHandler struct {
	agentSvc services.AgentManagerService
}

func NewRuntimeLogHandler(agentSvc services.AgentManagerService) *RuntimeLogHandler {
	return &RuntimeLogHandler{agentSvc: agentSvc}
}

func (h *RuntimeLogHandler) GetRuntimeLogs(ctx context.Context, orgName string, projectName string, agentName string, payload spec.LogFilterRequest) (*models.LogsResponse, error) {
	return h.agentSvc.GetAgentRuntimeLogs(ctx, orgName, projectName, agentName, payload)
}

// TraceHandler bridges MCP trace tools to the trace-observer service.
type TraceHandler struct {
	traceClient traceobserversvc.TraceObserverClient
}

func NewTraceHandler(traceClient traceobserversvc.TraceObserverClient) *TraceHandler {
	return &TraceHandler{traceClient: traceClient}
}

func (h *TraceHandler) ListTraces(ctx context.Context, orgName string, projectName string, agentName string, environment string, startTime string, endTime string, sortOrder string, limit int, offset int) (map[string]any, error) {
	if h.traceClient == nil {
		return nil, fmt.Errorf("trace observer client is not configured")
	}

	params := traceobserversvc.TraceListParams{
		Organization: orgName,
		Project:      projectName,
		Component:    agentName,
		Environment:  environment,
		StartTime:    startTime,
		EndTime:      endTime,
		Limit:        limit,
		Offset:       offset,
		SortOrder:    sortOrder,
	}

	return h.traceClient.ListTraces(ctx, params)
}

func (h *TraceHandler) ExportTraces(ctx context.Context, orgName string, projectName string, agentName string, environment string, startTime string, endTime string, sortOrder string, limit int, offset int) (map[string]any, error) {
	if h.traceClient == nil {
		return nil, fmt.Errorf("trace observer client is not configured")
	}

	params := traceobserversvc.TraceListParams{
		Organization: orgName,
		Project:      projectName,
		Component:    agentName,
		Environment:  environment,
		StartTime:    startTime,
		EndTime:      endTime,
		Limit:        limit,
		Offset:       offset,
		SortOrder:    sortOrder,
	}

	return h.traceClient.ExportTraces(ctx, params)
}

func (h *TraceHandler) GetTraceDetails(ctx context.Context, orgName string, projectName string, agentName string, traceID string, environment string) (map[string]any, error) {
	if h.traceClient == nil {
		return nil, fmt.Errorf("trace observer client is not configured")
	}

	now := time.Now().UTC()
	params := traceobserversvc.TraceDetailsParams{
		TraceID:      traceID,
		Organization: orgName,
		Project:      projectName,
		Component:    agentName,
		Environment:  environment,
		StartTime:    now.AddDate(0, 0, -7).Format(time.RFC3339),
		EndTime:      now.Format(time.RFC3339),
		Limit:        1000, // fetch all spans for the trace
	}

	return h.traceClient.GetTrace(ctx, params)
}
