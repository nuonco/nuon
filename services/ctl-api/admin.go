package main

import (
	_ "github.com/powertoolsdev/mono/pkg/metrics"
	_ "github.com/powertoolsdev/mono/pkg/types/state"
	_ "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

//	@title						Nuon Admin API
//	@version					1.0.0
//	@description				Admin API for managing nuon apps, components, installs, actions.
//	@contact.name				Nuon Support
//	@contact.email				support@nuon.co
//	@BasePath					/
//	@schemes					http
//
//	@securityDefinitions.apiKey	AdminEmail
//	@in							header
//	@name						X-Nuon-Admin-Email
//	@description				admin email
//
