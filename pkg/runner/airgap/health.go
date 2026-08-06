package airgap

import (
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

// HealthSnapshot is the persisted form of one component-health report. It is
// written to the state store (and mirrored to customer S3) instead of being
// POSTed to ctl-api, and is the document `nuon-bundle health` renders.
type HealthSnapshot struct {
	ObservedAt         time.Time              `json:"observed_at"`
	Kind               string                 `json:"kind,omitempty"`
	ClusterAccessError string                 `json:"cluster_access_error,omitempty"`
	Components         []ComponentHealth      `json:"components,omitempty"`
	SandboxReleases    []SandboxReleaseHealth `json:"sandbox_releases,omitempty"`
}

type ComponentHealth struct {
	InstallComponentID string           `json:"install_component_id,omitempty"`
	ComponentID        string           `json:"component_id,omitempty"`
	ComponentName      string           `json:"component_name,omitempty"`
	ComponentType      string           `json:"component_type,omitempty"`
	Health             string           `json:"health"`
	Truncated          bool             `json:"truncated,omitempty"`
	Resources          []ResourceHealth `json:"resources,omitempty"`
}

type SandboxReleaseHealth struct {
	ReleaseName string           `json:"release_name"`
	Namespace   string           `json:"namespace,omitempty"`
	Health      string           `json:"health"`
	Resources   []ResourceHealth `json:"resources,omitempty"`
}

type ResourceHealth struct {
	Kind      string `json:"kind,omitempty"`
	APIGroup  string `json:"api_group,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Provider  string `json:"provider,omitempty"`
	Health    string `json:"health,omitempty"`
	Message   string `json:"message,omitempty"`
}

// HealthTransition records one component (or sandbox release) changing its
// aggregate health. Transitions are appended immutably so a degradation that
// recovers before anyone looks is still visible afterwards.
type HealthTransition struct {
	ComponentID   string    `json:"component_id,omitempty"`
	ComponentName string    `json:"component_name,omitempty"`
	From          string    `json:"from"`
	To            string    `json:"to"`
	ObservedAt    time.Time `json:"observed_at"`
}

// healthSeverity orders verdicts for aggregation; the component's health is
// its worst resource. not-applicable and empty verdicts carry no signal and
// are excluded (severity -1).
var healthSeverity = map[string]int{
	"healthy":     0,
	"progressing": 1,
	"unknown":     2,
	"degraded":    3,
	"unhealthy":   4,
}

func aggregateHealth(resources []ResourceHealth) string {
	worst, found := "", false
	for _, r := range resources {
		severity, ok := healthSeverity[r.Health]
		if !ok {
			continue
		}
		if !found || severity > healthSeverity[worst] {
			worst, found = r.Health, true
		}
	}
	if !found {
		return "unknown"
	}
	return worst
}

// NewHealthSnapshot converts a runner health report into the persisted form,
// resolving component names from the envelope's component specs.
func NewHealthSnapshot(req *models.ServiceCreateComponentHealthRequest, specs []ComponentSpec) *HealthSnapshot {
	byInstallComponent := make(map[string]ComponentSpec, len(specs))
	byComponent := make(map[string]ComponentSpec, len(specs))
	for _, spec := range specs {
		byInstallComponent[spec.InstallComponentID] = spec
		byComponent[spec.ComponentID] = spec
	}

	snapshot := &HealthSnapshot{
		ObservedAt:         time.Now().UTC(),
		Kind:               req.Kind,
		ClusterAccessError: req.ClusterAccessError,
	}
	if req.ObservedAt != "" {
		if t, err := time.Parse(time.RFC3339, req.ObservedAt); err == nil {
			snapshot.ObservedAt = t.UTC()
		}
	}

	for _, c := range req.Components {
		if c == nil {
			continue
		}
		component := ComponentHealth{
			ComponentID:   c.ComponentID,
			ComponentType: c.ComponentType,
			Truncated:     c.Truncated,
			Resources:     convertResources(c.Resources),
		}
		if c.InstallComponentID != nil {
			component.InstallComponentID = *c.InstallComponentID
		}
		spec, ok := byComponent[c.ComponentID]
		if !ok && component.InstallComponentID != "" {
			spec, ok = byInstallComponent[component.InstallComponentID]
		}
		if ok {
			component.ComponentName = spec.ComponentName
			if component.ComponentType == "" {
				component.ComponentType = spec.ComponentType
			}
		}
		component.Health = aggregateHealth(component.Resources)
		snapshot.Components = append(snapshot.Components, component)
	}

	for _, r := range req.SandboxReleases {
		if r == nil {
			continue
		}
		release := SandboxReleaseHealth{
			Namespace: r.Namespace,
			Resources: convertResources(r.Resources),
		}
		if r.ReleaseName != nil {
			release.ReleaseName = *r.ReleaseName
		}
		release.Health = aggregateHealth(release.Resources)
		snapshot.SandboxReleases = append(snapshot.SandboxReleases, release)
	}

	return snapshot
}

func convertResources(resources []*models.ServiceComponentHealthResource) []ResourceHealth {
	out := make([]ResourceHealth, 0, len(resources))
	for _, r := range resources {
		if r == nil {
			continue
		}
		out = append(out, ResourceHealth{
			Kind:      r.Kind,
			APIGroup:  r.APIGroup,
			Name:      r.Name,
			Namespace: r.Namespace,
			Provider:  r.Provider,
			Health:    r.Health,
			Message:   r.Message,
		})
	}
	return out
}

// HealthTransitions diffs two snapshots and returns one transition per
// component or sandbox release whose aggregate health changed. A previous nil
// snapshot yields transitions from "" so first observations are recorded too.
func HealthTransitions(previous, current *HealthSnapshot) []HealthTransition {
	before := map[string]string{}
	if previous != nil {
		for _, c := range previous.Components {
			before["component:"+c.ComponentID] = c.Health
		}
		for _, r := range previous.SandboxReleases {
			before["sandbox:"+r.ReleaseName] = r.Health
		}
	}

	var out []HealthTransition
	for _, c := range current.Components {
		if from := before["component:"+c.ComponentID]; from != c.Health {
			out = append(out, HealthTransition{
				ComponentID:   c.ComponentID,
				ComponentName: c.ComponentName,
				From:          from,
				To:            c.Health,
				ObservedAt:    current.ObservedAt,
			})
		}
	}
	for _, r := range current.SandboxReleases {
		if from := before["sandbox:"+r.ReleaseName]; from != r.Health {
			out = append(out, HealthTransition{
				ComponentName: "sandbox/" + r.ReleaseName,
				From:          from,
				To:            r.Health,
				ObservedAt:    current.ObservedAt,
			})
		}
	}
	return out
}
