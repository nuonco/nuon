package state

import (
	"time"

	"github.com/nuonco/nuon/pkg/metrics"
	pkgstate "github.com/nuonco/nuon/pkg/types/state"
)

type ExecuteRegenerationRequest struct {
	InstallID       string
	TriggeredByID   string
	TriggeredByType string

	Targets        []PartialTarget
	ForceAll       bool
	CachedState    *pkgstate.State
	LastModifiedAt map[PartialName]time.Time

	// MetricsWriter is optional — when set, Regenerate emits timing and count metrics.
	MetricsWriter metrics.Writer
}

type ExecuteRegenerationResponse struct {
	State           *pkgstate.State
	UpdatedPartials []PartialName
	LastModifiedAt  map[PartialName]time.Time
	GeneratedAt     time.Time
	AppID           string
	AppName         string
	AppConfigID     string
}
