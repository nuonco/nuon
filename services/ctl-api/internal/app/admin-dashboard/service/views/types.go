package views

import (
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type RunnerDetailView struct {
	Runner        app.Runner                           `json:"runner"`
	InstallID     string                               `json:"install_id"`
	InstallName   string                               `json:"install_name"`
	Process       *app.RunnerProcess                   `json:"process"`
	ProcessOnline bool                                 `json:"process_online"`
	Configs       map[string]*app.SandboxModeJobConfig `json:"configs"`
}

type LabelSearchResult struct {
	EntityType string        `json:"entity_type"`
	EntityID   string        `json:"entity_id"`
	EntityName string        `json:"entity_name"`
	Labels     labels.Labels `json:"labels"`
	DetailURL  string        `json:"detail_url"`
}

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

type OrgOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
