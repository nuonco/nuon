package main

import (
	_ "github.com/powertoolsdev/mono/pkg/metrics"
	_ "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/cmd"
)

// @title			Nuon
// @version		1.0.0
// @description	API for managing nuon apps, components, and installs.
// @contact.name	Nuon Support
// @contact.email	support@nuon.co
// @BasePath		/
// @schemes		https
//
// @securityDefinitions.apiKey APIKey
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and token.
//
// @securityDefinitions.apiKey OrgID
// @in header
// @name X-Nuon-Org-ID
// @description Nuon org ID
//
//go:generate ./generate.sh
func main() {
	cmd.Execute()
}
