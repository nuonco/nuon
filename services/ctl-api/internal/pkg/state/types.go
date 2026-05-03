package state

import (
	"time"

	pkgstate "github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ForceRegenerateRequest struct {
	TriggeredByID   string
	TriggeredByType app.InstallStateGenerateSource
}

type ForceRegenerateResponse struct{}

type HintRequest struct {
	Targets         []PartialTarget
	TriggeredByID   string
	TriggeredByType app.InstallStateGenerateSource
}

type HintResponse struct{}

type ExecuteRegenerationRequest struct {
	InstallID       string
	TriggeredByID   string
	TriggeredByType app.InstallStateGenerateSource

	Targets        []PartialTarget
	ForceAll       bool
	CachedState    *pkgstate.State
	LastModifiedAt map[PartialName]time.Time
}

type ExecuteRegenerationResponse struct {
	State           *pkgstate.State
	UpdatedPartials []PartialName
	LastModifiedAt  map[PartialName]time.Time
	GeneratedAt     time.Time
}
