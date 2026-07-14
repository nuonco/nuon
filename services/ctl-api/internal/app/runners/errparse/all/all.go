// Package all blank-imports every errparse parser package so their init()
// registrations land in the default registry.
//
// It is the single authoritative wiring point. Import it for its side effects
// wherever composite-error parsing must be enabled, and add any new parser
// package here so every consumer picks it up.
package all

import (
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/aws"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/generic"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/helm"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/terraform"
)
