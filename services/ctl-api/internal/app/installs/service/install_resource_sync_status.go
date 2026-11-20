package service

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/syncstatus"
)

type InstallResourceSyncStatus = syncstatus.InstallResourceSyncStatus
type InstallResourceSyncStatusResource = syncstatus.InstallResourceSyncStatusResource
type InstallResourceSyncStatusResourceKind = syncstatus.InstallResourceSyncStatusResourceKind
type InstallResourceSyncStatusCategory = syncstatus.InstallResourceSyncStatusCategory
type InstallResourceSyncStatusAttempt = syncstatus.InstallResourceSyncStatusAttempt
type InstallResourceSyncStatusOpenDeployment = syncstatus.InstallResourceSyncStatusOpenDeployment
type InstallResourceSyncStatusExplanation = syncstatus.InstallResourceSyncStatusExplanation
type InstallResourceSyncStatusPlan = syncstatus.InstallResourceSyncStatusPlan

const (
	InstallResourceSyncStatusResourceKindComponent = syncstatus.InstallResourceSyncStatusResourceKindComponent
	InstallResourceSyncStatusResourceKindImage     = syncstatus.InstallResourceSyncStatusResourceKindImage
	InstallResourceSyncStatusResourceKindSandbox   = syncstatus.InstallResourceSyncStatusResourceKindSandbox
	InstallResourceSyncStatusResourceKindStack     = syncstatus.InstallResourceSyncStatusResourceKindStack

	InstallResourceSyncStatusCategoryChangesNotApplied  = syncstatus.InstallResourceSyncStatusCategoryChangesNotApplied
	InstallResourceSyncStatusCategoryStateDriftDetected = syncstatus.InstallResourceSyncStatusCategoryStateDriftDetected
	InstallResourceSyncStatusCategoryInSync             = syncstatus.InstallResourceSyncStatusCategoryInSync
	InstallResourceSyncStatusCategoryUnknown            = syncstatus.InstallResourceSyncStatusCategoryUnknown
	InstallResourceSyncStatusCategoryDisabled           = syncstatus.InstallResourceSyncStatusCategoryDisabled
	InstallResourceSyncStatusCategoryTornDown           = syncstatus.InstallResourceSyncStatusCategoryTornDown
)
