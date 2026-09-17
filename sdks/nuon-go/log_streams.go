package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// LogStreamLogFilters mirrors the filter query parameters shared by the log
// read and tail endpoints. Zero-value fields are ignored.
type LogStreamLogFilters struct {
	StartTime              string
	EndTime                string
	ServiceNames           []string
	ScopeNames             []string
	ScopeVersions          []string
	SeverityTexts          []string
	SeverityNumberMin      int64
	SeverityNumberMax      int64
	ResourceSchemaURLs     []string
	ScopeSchemaURLs        []string
	TraceID                string
	SpanID                 string
	TraceFlags             int64
	RunnerID               string
	RunnerJobID            string
	RunnerGroupID          string
	RunnerJobExecutionID   string
	RunnerJobExecutionStep string
	Tools                  []string
	HelmReleaseName        string
	HelmChartName          string
	HelmChartID            string
	HelmNamespace          string
	HelmOperation          string
	TfWorkspaceID          string
	TfOperation            string
	K8sKind                string
	K8sNamespace           string
	K8sName                string
	K8sOperation           string
	Attrs                  []string
	ResourceAttrs          []string
	ScopeAttrs             []string
	BodyContains           string
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func int64Ptr(n int64) *int64 {
	if n == 0 {
		return nil
	}
	return &n
}

func (f *LogStreamLogFilters) applyToReadLogs(p *operations.LogStreamReadLogsParams) {
	if f == nil {
		return
	}
	p.StartTime = strPtr(f.StartTime)
	p.EndTime = strPtr(f.EndTime)
	p.ServiceName = f.ServiceNames
	p.ScopeName = f.ScopeNames
	p.ScopeVersion = f.ScopeVersions
	p.SeverityText = f.SeverityTexts
	p.SeverityNumberMin = int64Ptr(f.SeverityNumberMin)
	p.SeverityNumberMax = int64Ptr(f.SeverityNumberMax)
	p.ResourceSchemaURL = f.ResourceSchemaURLs
	p.ScopeSchemaURL = f.ScopeSchemaURLs
	p.TraceID = strPtr(f.TraceID)
	p.SpanID = strPtr(f.SpanID)
	p.TraceFlags = int64Ptr(f.TraceFlags)
	p.RunnerID = strPtr(f.RunnerID)
	p.RunnerJobID = strPtr(f.RunnerJobID)
	p.RunnerGroupID = strPtr(f.RunnerGroupID)
	p.RunnerJobExecutionID = strPtr(f.RunnerJobExecutionID)
	p.RunnerJobExecutionStep = strPtr(f.RunnerJobExecutionStep)
	p.Tool = f.Tools
	p.HelmReleaseName = strPtr(f.HelmReleaseName)
	p.HelmChartName = strPtr(f.HelmChartName)
	p.HelmChartID = strPtr(f.HelmChartID)
	p.HelmNamespace = strPtr(f.HelmNamespace)
	p.HelmOperation = strPtr(f.HelmOperation)
	p.TfWorkspaceID = strPtr(f.TfWorkspaceID)
	p.TfOperation = strPtr(f.TfOperation)
	p.K8sKind = strPtr(f.K8sKind)
	p.K8sNamespace = strPtr(f.K8sNamespace)
	p.K8sName = strPtr(f.K8sName)
	p.K8sOperation = strPtr(f.K8sOperation)
	p.Attr = f.Attrs
	p.ResourceAttr = f.ResourceAttrs
	p.ScopeAttr = f.ScopeAttrs
	p.Q = strPtr(f.BodyContains)
}

func (f *LogStreamLogFilters) applyToTailLogs(p *operations.LogStreamTailLogsParams) {
	if f == nil {
		return
	}
	p.StartTime = strPtr(f.StartTime)
	p.EndTime = strPtr(f.EndTime)
	p.ServiceName = f.ServiceNames
	p.ScopeName = f.ScopeNames
	p.ScopeVersion = f.ScopeVersions
	p.SeverityText = f.SeverityTexts
	p.SeverityNumberMin = int64Ptr(f.SeverityNumberMin)
	p.SeverityNumberMax = int64Ptr(f.SeverityNumberMax)
	p.ResourceSchemaURL = f.ResourceSchemaURLs
	p.ScopeSchemaURL = f.ScopeSchemaURLs
	p.TraceID = strPtr(f.TraceID)
	p.SpanID = strPtr(f.SpanID)
	p.TraceFlags = int64Ptr(f.TraceFlags)
	p.RunnerID = strPtr(f.RunnerID)
	p.RunnerJobID = strPtr(f.RunnerJobID)
	p.RunnerGroupID = strPtr(f.RunnerGroupID)
	p.RunnerJobExecutionID = strPtr(f.RunnerJobExecutionID)
	p.RunnerJobExecutionStep = strPtr(f.RunnerJobExecutionStep)
	p.Tool = f.Tools
	p.HelmReleaseName = strPtr(f.HelmReleaseName)
	p.HelmChartName = strPtr(f.HelmChartName)
	p.HelmChartID = strPtr(f.HelmChartID)
	p.HelmNamespace = strPtr(f.HelmNamespace)
	p.HelmOperation = strPtr(f.HelmOperation)
	p.TfWorkspaceID = strPtr(f.TfWorkspaceID)
	p.TfOperation = strPtr(f.TfOperation)
	p.K8sKind = strPtr(f.K8sKind)
	p.K8sNamespace = strPtr(f.K8sNamespace)
	p.K8sName = strPtr(f.K8sName)
	p.K8sOperation = strPtr(f.K8sOperation)
	p.Attr = f.Attrs
	p.ResourceAttr = f.ResourceAttrs
	p.ScopeAttr = f.ScopeAttrs
	p.Q = strPtr(f.BodyContains)
}

func (c *client) GetLogStream(ctx context.Context, logStreamID string) (*models.AppLogStream, error) {
	resp, err := c.genClient.Operations.GetLogStream(&operations.GetLogStreamParams{
		LogStreamID: logStreamID,
		Context:     ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) LogStreamReadLogs(ctx context.Context, logStreamId string, offset string, order string, filters *LogStreamLogFilters) ([]*models.AppOtelLogRecord, error) {
	params := &operations.LogStreamReadLogsParams{
		LogStreamID:    logStreamId,
		XNuonAPIOffset: &offset,
		Context:        ctx,
	}
	if order != "" {
		params.Order = &order
	}
	filters.applyToReadLogs(params)
	resp, err := c.genClient.Operations.LogStreamReadLogs(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

// LogStreamTailLogs hits the long-poll tail endpoint. `since` is a composite
// cursor (`<unix_nano>:<id>`, empty for "from oldest"); `wait` is an optional
// Go duration string, capped server-side at 30s when omitted.
//
// The endpoint only supports ASC ordering — callers paginate history
// through the read endpoint instead.
func (c *client) LogStreamTailLogs(ctx context.Context, logStreamID string, since string, wait string, filters *LogStreamLogFilters) (*models.ServiceLogStreamTailLogsResponse, error) {
	params := &operations.LogStreamTailLogsParams{
		LogStreamID: logStreamID,
		Context:     ctx,
	}
	if since != "" {
		params.Since = &since
	}
	if wait != "" {
		params.Wait = &wait
	}
	filters.applyToTailLogs(params)
	resp, err := c.genClient.Operations.LogStreamTailLogs(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *client) LogStreamReadLogsWithNextOffset(ctx context.Context, logStreamId string, offset string, order string, filters *LogStreamLogFilters) ([]*models.AppOtelLogRecord, string, error) {
	hr := newResponseHeaderReader(&operations.LogStreamReadLogsReader{})

	params := &operations.LogStreamReadLogsParams{
		LogStreamID:    logStreamId,
		XNuonAPIOffset: &offset,
		Context:        ctx,
	}
	if order != "" {
		params.Order = &order
	}
	filters.applyToReadLogs(params)
	resp, err := c.genClient.Operations.LogStreamReadLogs(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, "", err
	}

	nextOffset := hr.GetHeader("X-Nuon-API-Next")
	return resp.Payload, nextOffset, nil
}
