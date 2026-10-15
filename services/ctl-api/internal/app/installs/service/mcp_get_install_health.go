package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

const (
	mcpInstallResourceCardLimit = 200
	mcpRunnerHeartbeatRecent    = 60 * time.Second
)

type mcpInstallComponentHealth struct {
	Name                    string `json:"name"`
	DeployStatus            string `json:"deploy_status,omitempty"`
	DeployStatusDescription string `json:"deploy_status_description,omitempty"`
	HealthStatus            string `json:"health_status,omitempty"`
	HealthStatusDescription string `json:"health_status_description,omitempty"`
}

type mcpInstallStatusAxis struct {
	Status      string `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
}

type mcpInstallHealthStatuses struct {
	Runner     mcpInstallStatusAxis  `json:"runner"`
	Sandbox    mcpInstallStatusAxis  `json:"sandbox"`
	Components mcpInstallStatusAxis  `json:"components"`
	Health     *mcpInstallStatusAxis `json:"health,omitempty"`
}

type mcpInstallResourceCard struct {
	ComponentName     string           `json:"component_name,omitempty"`
	Kind              string           `json:"kind,omitempty"`
	Name              string           `json:"name"`
	Namespace         string           `json:"namespace,omitempty"`
	Health            string           `json:"health,omitempty"`
	Message           string           `json:"message,omitempty"`
	RemovedFromConfig bool             `json:"removed_from_config,omitempty"`
	Fields            []mcpStatusField `json:"fields,omitempty"`
}

type mcpStatusField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type mcpRunnerProcessStatus struct {
	ID                string `json:"id"`
	Type              string `json:"type"`
	Label             string `json:"label"`
	Status            string `json:"status,omitempty"`
	Connectivity      string `json:"connectivity"`
	StartedAt         string `json:"started_at,omitempty"`
	LastHeartbeat     string `json:"last_heartbeat,omitempty"`
	ConfiguredVersion string `json:"configured_version,omitempty"`
	ReportedVersion   string `json:"reported_version,omitempty"`
	StatusDescription string `json:"status_description,omitempty"`
}

type mcpGetInstallHealthResult struct {
	InstallID          string                      `json:"install_id"`
	InstallName        string                      `json:"install_name"`
	Statuses           mcpInstallHealthStatuses    `json:"statuses"`
	ClusterAccessError string                      `json:"cluster_access_error,omitempty"`
	LastReportedAt     string                      `json:"last_reported_at,omitempty"`
	Components         []mcpInstallComponentHealth `json:"components"`
	Resources          []mcpInstallResourceCard    `json:"resources"`
	ResourcesOmitted   int                         `json:"resources_omitted,omitempty"`
	RunnerProcesses    []mcpRunnerProcessStatus    `json:"runner_processes"`
}

func (s *service) mcpGetInstallHealth(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetInstallInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}
	install, err := s.findInstall(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, err
	}

	statuses := mcpInstallHealthStatuses{
		Runner: mcpInstallStatusAxis{
			Status:      string(install.RunnerStatus),
			Description: install.RunnerStatusDescription,
		},
		Sandbox: mcpSandboxStatus(install),
		Components: mcpInstallStatusAxis{
			Status:      string(install.CompositeComponentStatus),
			Description: install.CompositeComponentStatusDescription,
		},
	}
	if install.CompositeHealthStatus != "" {
		statuses.Health = &mcpInstallStatusAxis{
			Status:      string(install.CompositeHealthStatus),
			Description: install.CompositeHealthStatusDescription,
		}
	}

	components := make([]mcpInstallComponentHealth, 0, len(install.InstallComponents))
	componentNames := map[string]string{}
	for _, component := range install.InstallComponents {
		componentNames[component.ID] = component.Component.Name
		components = append(components, mcpInstallComponentHealth{
			Name:                    component.Component.Name,
			DeployStatus:            string(component.Status),
			DeployStatusDescription: component.StatusDescription,
			HealthStatus:            string(component.HealthStatus),
			HealthStatusDescription: component.HealthStatusDescription,
		})
	}

	resources, omitted, err := s.mcpInstallResourceCards(ctx, install, componentNames)
	if err != nil {
		return nil, nil, err
	}
	processes, err := s.mcpInstallRunnerProcesses(ctx, orgID, install)
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(mcpGetInstallHealthResult{
		InstallID:          install.ID,
		InstallName:        install.Name,
		Statuses:           statuses,
		ClusterAccessError: install.HealthClusterError,
		LastReportedAt:     apiPkg.MCPTimePtr(install.LastHealthReportAt),
		Components:         components,
		Resources:          resources,
		ResourcesOmitted:   omitted,
		RunnerProcesses:    processes,
	})
}

func mcpSandboxStatus(install *app.Install) mcpInstallStatusAxis {
	axis := mcpInstallStatusAxis{Status: string(install.SandboxStatus)}
	if install.SandboxStatus == app.InstallSandboxStatusActive && install.SandboxHealthStatus != "" {
		axis.Status = install.SandboxHealthStatus
		axis.Description = install.SandboxHealthMessage
	}
	return axis
}

func (s *service) mcpInstallResourceCards(ctx context.Context, install *app.Install, componentNames map[string]string) ([]mcpInstallResourceCard, int, error) {
	rows, err := s.getInstallResources(ctx, install.OrgID, install.ID, installResourceFilters{})
	if err != nil {
		return nil, 0, err
	}
	omitted := 0
	if len(rows) > mcpInstallResourceCardLimit {
		omitted = len(rows) - mcpInstallResourceCardLimit
		rows = rows[:mcpInstallResourceCardLimit]
	}
	cards := make([]mcpInstallResourceCard, 0, len(rows))
	for _, row := range rows {
		cards = append(cards, mcpInstallResourceCard{
			ComponentName:     componentNames[row.InstallComponentID],
			Kind:              row.Kind,
			Name:              row.Name,
			Namespace:         row.Namespace,
			Health:            row.Health,
			Message:           row.Message,
			RemovedFromConfig: row.RemovedFromConfig,
			Fields:            mcpResourceStatusFields(row),
		})
	}
	return cards, omitted, nil
}

func (s *service) mcpInstallRunnerProcesses(ctx context.Context, orgID string, install *app.Install) ([]mcpRunnerProcessStatus, error) {
	if install.RunnerID == "" {
		return []mcpRunnerProcessStatus{}, nil
	}
	var processes []app.RunnerProcess
	resp := s.db.WithContext(ctx).
		Where(
			"runner_id = ? AND org_id = ? AND composite_status->>'status' IN ?",
			install.RunnerID,
			orgID,
			[]string{
				string(app.StatusPending),
				string(app.RunnerProcessStatusActive),
				string(app.RunnerProcessStatusOffline),
				string(app.RunnerProcessStatusPendingShutdown),
			},
		).
		Order("created_at DESC").
		Find(&processes)
	if resp.Error != nil {
		return nil, resp.Error
	}

	latest := map[app.RunnerProcessType]app.RunnerProcess{}
	for _, process := range processes {
		if _, ok := latest[process.Type]; ok {
			continue
		}
		latest[process.Type] = process
	}

	var settings app.RunnerGroupSettings
	if install.RunnerGroup.ID != "" {
		if err := s.db.WithContext(ctx).
			Where("runner_group_id = ?", install.RunnerGroup.ID).
			First(&settings).Error; err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	out := make([]mcpRunnerProcessStatus, 0, 2)
	for _, processType := range []app.RunnerProcessType{app.RunnerProcessTypeInstall, app.RunnerProcessTypeMng} {
		process, ok := latest[processType]
		if !ok {
			continue
		}
		row, err := s.mcpRunnerProcessStatus(ctx, process, settings)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, nil
}

func (s *service) mcpRunnerProcessStatus(ctx context.Context, process app.RunnerProcess, settings app.RunnerGroupSettings) (mcpRunnerProcessStatus, error) {
	row := mcpRunnerProcessStatus{
		ID:                process.ID,
		Type:              string(process.Type),
		Label:             mcpRunnerProcessLabel(process.Type),
		Status:            string(process.CompositeStatus.Status),
		StatusDescription: process.CompositeStatus.StatusHumanDescription,
		Connectivity:      "not-connected",
		StartedAt:         apiPkg.MCPTimePtr(process.StartedAt),
		ReportedVersion:   process.Version,
	}
	if process.Type == app.RunnerProcessTypeMng {
		row.ConfiguredVersion = settings.BinaryVersion
	} else {
		row.ConfiguredVersion = settings.ContainerImageTag
	}

	var beat app.LatestRunnerHeartBeat
	err := s.chDB.WithContext(ctx).
		Where("runner_id = ? AND process_id = ?", process.RunnerID, process.ID).
		First(&beat).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return row, nil
		}
		return mcpRunnerProcessStatus{}, err
	}
	row.LastHeartbeat = apiPkg.MCPTime(beat.CreatedAt)
	if beat.Version != "" {
		row.ReportedVersion = beat.Version
	}
	if !beat.CreatedAt.IsZero() && time.Since(beat.CreatedAt) >= 0 && time.Since(beat.CreatedAt) < mcpRunnerHeartbeatRecent {
		row.Connectivity = "connected"
	}
	return row, nil
}

func mcpRunnerProcessLabel(processType app.RunnerProcessType) string {
	switch processType {
	case app.RunnerProcessTypeInstall:
		return "Install process"
	case app.RunnerProcessTypeMng:
		return "Mng process"
	default:
		return string(processType) + " process"
	}
}

func mcpResourceStatusFields(row app.InstallComponentResourceState) []mcpStatusField {
	var details map[string]any
	if err := json.Unmarshal([]byte(row.Details), &details); err != nil {
		return nil
	}
	spec := mcpObject(details["spec"])
	statusObj := mcpObject(details["status"])
	fields := []mcpStatusField{}
	switch row.Kind {
	case "Deployment", "StatefulSet", "ReplicaSet":
		if ready, ok := mcpJoinCount(statusObj["readyReplicas"], spec["replicas"]); ok {
			fields = append(fields, mcpStatusField{Name: "Ready", Value: ready})
		}
	case "DaemonSet":
		if ready, ok := mcpJoinCount(statusObj["numberReady"], statusObj["desiredNumberScheduled"]); ok {
			fields = append(fields, mcpStatusField{Name: "Ready", Value: ready})
		}
	case "Service":
		if value := mcpScalar(spec["type"]); value != "" {
			fields = append(fields, mcpStatusField{Name: "Type", Value: value})
		}
		if ports := mcpServicePorts(spec["ports"]); ports != "" {
			fields = append(fields, mcpStatusField{Name: "Ports", Value: ports})
		}
	case "Ingress":
		if value := mcpScalar(spec["ingressClassName"]); value != "" {
			fields = append(fields, mcpStatusField{Name: "Class", Value: value})
		}
	}
	for _, condition := range mcpObjects(statusObj["conditions"]) {
		name := mcpScalar(condition["type"])
		value := mcpScalar(condition["status"])
		if name != "" && value != "" {
			fields = append(fields, mcpStatusField{Name: name, Value: value})
		}
	}
	if address := mcpLoadBalancerAddress(statusObj); address != "" {
		fields = append(fields, mcpStatusField{Name: "Address", Value: address})
	}
	return fields
}

func mcpServicePorts(value any) string {
	ports := []string{}
	for _, port := range mcpObjects(value) {
		ports = append(ports, fmt.Sprintf("%s → %s/%s", mcpScalar(port["port"]), mcpScalar(port["targetPort"]), mcpScalar(port["protocol"])))
	}
	return strings.Join(ports, ", ")
}

func mcpLoadBalancerAddress(statusObj map[string]any) string {
	addresses := []string{}
	for _, ingress := range mcpObjects(mcpObject(statusObj["loadBalancer"])["ingress"]) {
		if host := mcpScalar(ingress["hostname"]); host != "" {
			addresses = append(addresses, host)
			continue
		}
		if ip := mcpScalar(ingress["ip"]); ip != "" {
			addresses = append(addresses, ip)
		}
	}
	return strings.Join(addresses, ", ")
}

func mcpJoinCount(ready, total any) (string, bool) {
	readyText := mcpScalar(ready)
	if readyText == "" {
		return "", false
	}
	if totalText := mcpScalar(total); totalText != "" {
		return readyText + " / " + totalText, true
	}
	return readyText, true
}

func mcpObject(value any) map[string]any {
	object, _ := value.(map[string]any)
	if object == nil {
		return map[string]any{}
	}
	return object
}

func mcpObjects(value any) []map[string]any {
	items, _ := value.([]any)
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if object, ok := item.(map[string]any); ok {
			out = append(out, object)
		}
	}
	return out
}

func mcpScalar(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		if typed == float64(int64(typed)) {
			return fmt.Sprintf("%d", int64(typed))
		}
		return fmt.Sprintf("%g", typed)
	case bool:
		if typed {
			return "True"
		}
		return "False"
	default:
		return ""
	}
}
