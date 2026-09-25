package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestResolveAWSConnection(t *testing.T) {
	images := func(id, name, target, principal string) app.CloudConnection {
		return app.CloudConnection{ID: id, Name: name, Platform: app.CloudPlatformAWS, TargetID: target, Principal: principal, Capabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityImages}}
	}
	const (
		account = "123456789012"
		image   = "123456789012.dkr.ecr.us-west-2.amazonaws.com/acme/api"
		role    = "arn:aws:iam::123456789012:role/ecr-pull"
	)
	tests := map[string]struct {
		name         string
		role         string
		connections  []app.CloudConnection
		wantID       string
		wantImplicit bool
		wantErr      string
	}{
		"explicit connection wins":  {name: "chosen", role: role, connections: []app.CloudConnection{images("chosen-id", "chosen", account, "arn:aws:iam::123456789012:role/chosen"), images("role-id", "role", account, role)}, wantID: "chosen-id"},
		"bare ARN matches existing": {role: role, connections: []app.CloudConnection{images("role-id", "role", account, role)}, wantID: "role-id"},
		"bare ARN creates implicit": {role: role, wantImplicit: true},
		"unique target match":       {connections: []app.CloudConnection{images("unique-id", "unique", account, role)}, wantID: "unique-id"},
		"ambiguous target":          {connections: []app.CloudConnection{images("one", "one", account, role), images("two", "two", account, "arn:aws:iam::123456789012:role/other")}, wantErr: "multiple images-capable"},
		"no target match":           {wantErr: "no images-capable"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := ResolveAWSConnection(test.name, test.role, image, "us-west-2", "org-test", test.connections)
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.wantImplicit, result.Implicit)
			if test.wantImplicit {
				assert.Equal(t, role, result.Connection.Principal)
				assert.Equal(t, app.CloudConnectionAuthModeLegacy, result.Connection.AuthMode)
			} else {
				assert.Equal(t, test.wantID, result.Connection.ID)
			}
		})
	}
}

func TestResolveCloudConnectionByImageURL(t *testing.T) {
	images := []app.CloudConnection{
		{ID: "aws-one", Name: "aws-one", Platform: app.CloudPlatformAWS, TargetID: "123456789012", Capabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityImages}},
		{ID: "gcp-one", Name: "gcp-one", Platform: app.CloudPlatformGCP, TargetID: "acme-project", Capabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityImages}},
		{ID: "azure-one", Name: "azure-one", Platform: app.CloudPlatformAzure, Registries: []string{"acme.azurecr.io"}, Capabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityImages}},
	}
	tests := map[string]struct {
		resolve func([]app.CloudConnection) (*app.CloudConnection, error)
		input   []app.CloudConnection
		wantID  string
		wantErr string
	}{
		"ECR match": {resolve: func(connections []app.CloudConnection) (*app.CloudConnection, error) {
			result, err := ResolveAWSConnection("", "", "123456789012.dkr.ecr.us-west-2.amazonaws.com/acme/api", "us-west-2", "org-test", connections)
			return result.Connection, err
		}, input: images, wantID: "aws-one"},
		"GAR match": {resolve: func(connections []app.CloudConnection) (*app.CloudConnection, error) {
			result, err := ResolveGCPConnection("", "", "", "acme-project", "us-docker.pkg.dev/acme-project/repo/api", "org-test", connections)
			return result.Connection, err
		}, input: images, wantID: "gcp-one"},
		"ACR match": {resolve: func(connections []app.CloudConnection) (*app.CloudConnection, error) {
			result, err := ResolveAzureConnection("", "", "", "acme.azurecr.io/api", "org-test", connections)
			return result.Connection, err
		}, input: images, wantID: "azure-one"},
		"GAR ambiguous lists candidates": {resolve: func(connections []app.CloudConnection) (*app.CloudConnection, error) {
			result, err := ResolveGCPConnection("", "", "", "acme-project", "us-docker.pkg.dev/acme-project/repo/api", "org-test", connections)
			return result.Connection, err
		}, input: append(images, app.CloudConnection{ID: "gcp-two", Name: "gcp-two", Platform: app.CloudPlatformGCP, TargetID: "acme-project", Capabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityImages}}), wantErr: "gcp-one, gcp-two"},
		"ACR no candidates": {resolve: func(connections []app.CloudConnection) (*app.CloudConnection, error) {
			result, err := ResolveAzureConnection("", "", "", "other.azurecr.io/api", "org-test", connections)
			return result.Connection, err
		}, input: images, wantErr: "no images-capable Azure"},
		"explicit wins over URL candidates": {resolve: func(connections []app.CloudConnection) (*app.CloudConnection, error) {
			result, err := ResolveAzureConnection("chosen", "", "", "acme.azurecr.io/api", "org-test", connections)
			return result.Connection, err
		}, input: append(images, app.CloudConnection{ID: "chosen", Name: "chosen", Platform: app.CloudPlatformAzure, Capabilities: []app.CloudConnectionCapability{app.CloudConnectionCapabilityImages}}), wantID: "chosen"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			connection, err := test.resolve(test.input)
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.wantID, connection.ID)
		})
	}
}
