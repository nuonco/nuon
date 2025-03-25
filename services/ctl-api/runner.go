package main

import (
	_ "github.com/powertoolsdev/mono/pkg/metrics"
	_ "github.com/powertoolsdev/mono/pkg/types/state"
	_ "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

//	@title						Nuon Runner API
//	@version					1.0.0
//	@description				API for runners.
//	@contact.name				Nuon Support
//	@contact.email				support@nuon.co
//	@BasePath					/
//	@schemes					https
//
//	@securityDefinitions.apiKey	APIKey
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and token.
