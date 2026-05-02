package views

import (
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// RunnerDetailView is the view data for the runner detail page.
type RunnerDetailView struct {
	Runner        app.Runner                           `json:"runner"`
	InstallID     string                               `json:"install_id"`
	InstallName   string                               `json:"install_name"`
	Process       *app.RunnerProcess                   `json:"process"`
	ProcessOnline bool                                 `json:"process_online"`
	Configs       map[string]*app.SandboxModeJobConfig `json:"configs"`
}

// SandboxRunnerView is the view data for the sandbox runners list.
type SandboxRunnerView struct {
	Runner        app.Runner                 `json:"runner"`
	ProcessOnline bool                       `json:"process_online"`
	Version       string                     `json:"version"`
	Configs       []app.SandboxModeJobConfig `json:"configs"`
	InstallID     string                     `json:"install_id"`
	InstallName   string                     `json:"install_name"`
}

// LabelSearchResult represents a labeled entity for the browse page.
type LabelSearchResult struct {
	EntityType string        `json:"entity_type"`
	EntityID   string        `json:"entity_id"`
	EntityName string        `json:"entity_name"`
	Labels     labels.Labels `json:"labels"`
	DetailURL  string        `json:"detail_url"`
}

// AllRunnerView is the view data for the all-runners list page.
type AllRunnerView struct {
	Runner        app.Runner `json:"runner"`
	OrgName       string     `json:"org_name"`
	GroupType     string     `json:"group_type"`
	ProcessOnline bool       `json:"process_online"`
	Version       string     `json:"version"`
	ProcessType   string     `json:"process_type"`
	InstallID     string     `json:"install_id"`
	InstallName   string     `json:"install_name"`
}

// OrgOption is used to populate the org selector dropdown.
type OrgOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
