package all

import (
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/aws"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/generic"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/helm"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/terraform"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/vcs"
)
