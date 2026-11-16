package service

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/alignment"
)

type InstallAlignment = alignment.InstallAlignment
type InstallAlignmentResource = alignment.InstallAlignmentResource
type InstallAlignmentResourceKind = alignment.InstallAlignmentResourceKind
type InstallAlignmentCategory = alignment.InstallAlignmentCategory
type InstallAlignmentOpenDeployment = alignment.InstallAlignmentOpenDeployment
type InstallAlignmentPlan = alignment.InstallAlignmentPlan

const (
	InstallAlignmentResourceKindComponent = alignment.InstallAlignmentResourceKindComponent
	InstallAlignmentResourceKindImage     = alignment.InstallAlignmentResourceKindImage
	InstallAlignmentResourceKindSandbox   = alignment.InstallAlignmentResourceKindSandbox
	InstallAlignmentResourceKindStack     = alignment.InstallAlignmentResourceKindStack

	InstallAlignmentCategoryChangesNotApplied  = alignment.InstallAlignmentCategoryChangesNotApplied
	InstallAlignmentCategoryStateDriftDetected = alignment.InstallAlignmentCategoryStateDriftDetected
	InstallAlignmentCategoryInSync             = alignment.InstallAlignmentCategoryInSync
	InstallAlignmentCategoryUnknown            = alignment.InstallAlignmentCategoryUnknown
	InstallAlignmentCategoryDisabled           = alignment.InstallAlignmentCategoryDisabled
	InstallAlignmentCategoryTornDown           = alignment.InstallAlignmentCategoryTornDown
)
