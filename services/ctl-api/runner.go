package main

import (
	_ "github.com/nuonco/nuon/pkg/metrics"
	_ "github.com/nuonco/nuon/pkg/types/state"
	_ "github.com/nuonco/nuon/pkg/types/workflows/executors/v1/plan/v1"
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
