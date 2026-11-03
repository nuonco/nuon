package service

import (
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SetupResponse struct {
	IssuerURL         string                    `json:"issuer_url"`
	Subject           string                    `json:"subject"`
	Audience          string                    `json:"audience"`
	TrustPolicy       map[string]any            `json:"trust_policy"`
	PermissionsPolicy map[string]any            `json:"permissions_policy,omitempty"`
	PermissionsCLI    string                    `json:"permissions_cli,omitempty"`
	Terraform         string                    `json:"terraform"`
	CLI               string                    `json:"cli"`
	CloudFormation    string                    `json:"cloudformation"`
	Preset            app.CloudConnectionPreset `json:"preset"`
}

func (s *service) setup(connection *app.CloudConnection, repositories ...string) SetupResponse {
	return s.setupWithPrincipal(connection, false, repositories...)
}

func (s *service) setupWithPrincipal(connection *app.CloudConnection, existingPrincipal bool, repositories ...string) SetupResponse {
	if connection.Platform == app.CloudPlatformAzure {
		return s.acrSetup(connection, existingPrincipal, repositories)
	}
	if connection.Platform == app.CloudPlatformGCP {
		return s.garSetup(connection, existingPrincipal, repositories)
	}
	return s.awsSetup(connection, existingPrincipal, repositories)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
