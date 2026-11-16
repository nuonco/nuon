package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	azureSubscription = "11111111-1111-1111-1111-111111111111"
	azureClient       = "22222222-2222-2222-2222-222222222222"
	azureTenant       = "33333333-3333-3333-3333-333333333333"
)

func awsConnection() app.CloudConnection {
	return app.CloudConnection{Name: "acme", Platform: app.CloudPlatformAWS, Preset: app.CloudConnectionPresetStacks, TargetID: "123456789012", Principal: "arn:aws:iam::123456789012:role/acme"}
}

func gcpConnection() app.CloudConnection {
	return app.CloudConnection{Name: "acme", Platform: app.CloudPlatformGCP, Preset: app.CloudConnectionPresetGAR, TargetID: "acme-project", GCPProjectNumber: "123456789012", Principal: "acme-pull@acme-project.iam.gserviceaccount.com", DefaultRegion: "us-central1"}
}

func azureConnection() app.CloudConnection {
	return app.CloudConnection{
		Name: "acme", Platform: app.CloudPlatformAzure, Preset: app.CloudConnectionPresetACR, TargetID: azureSubscription, Principal: azureClient, AzureTenantID: azureTenant,
		AzureRegistryID:          "/subscriptions/" + azureSubscription + "/resourceGroups/acme/providers/Microsoft.ContainerRegistry/registries/acmeregistry",
		AzureRegistryLoginServer: "acmeregistry.azurecr.io",
	}
}

func TestValidateConnection(t *testing.T) {
	for name, tc := range map[string]struct {
		connection func() app.CloudConnection
		edit       func(*app.CloudConnection)
		wantErr    string
	}{
		"aws":                            {connection: awsConnection},
		"gcp":                            {connection: gcpConnection},
		"azure":                          {connection: azureConnection},
		"blank name":                     {connection: awsConnection, edit: func(c *app.CloudConnection) { c.Name = " " }, wantErr: "name is required"},
		"unknown platform":               {connection: awsConnection, edit: func(c *app.CloudConnection) { c.Platform = "oracle" }, wantErr: "only aws, gcp, and azure"},
		"aws with gcp project number":    {connection: awsConnection, edit: func(c *app.CloudConnection) { c.GCPProjectNumber = "123456789012" }, wantErr: "gcp_project_number requires the gcp platform"},
		"aws with azure tenant":          {connection: awsConnection, edit: func(c *app.CloudConnection) { c.AzureTenantID = azureTenant }, wantErr: "require the azure platform"},
		"gcp with azure registry":        {connection: gcpConnection, edit: func(c *app.CloudConnection) { c.AzureRegistryID = "/subscriptions/x" }, wantErr: "require the azure platform"},
		"gcp with azure login server":    {connection: gcpConnection, edit: func(c *app.CloudConnection) { c.AzureRegistryLoginServer = "acmeregistry.azurecr.io" }, wantErr: "require the azure platform"},
		"azure with gcp project number":  {connection: azureConnection, edit: func(c *app.CloudConnection) { c.GCPProjectNumber = "123456789012" }, wantErr: "gcp_project_number requires the gcp platform"},
		"cross-field check before rules": {connection: awsConnection, edit: func(c *app.CloudConnection) { c.TargetID, c.AzureTenantID = "bad", azureTenant }, wantErr: "require the azure platform"},
	} {
		t.Run(name, func(t *testing.T) {
			connection := tc.connection()
			if tc.edit != nil {
				tc.edit(&connection)
			}
			err := validateConnection(&connection)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateAWS(t *testing.T) {
	for name, tc := range map[string]struct {
		edit       func(*app.CloudConnection)
		wantErr    string
		wantRegion string
	}{
		"stacks defaults region": {wantRegion: "us-east-1"},
		"custom keeps region":    {edit: func(c *app.CloudConnection) { c.Preset, c.DefaultRegion = app.CloudConnectionPresetCustom, "eu-west-1" }, wantRegion: "eu-west-1"},
		"role path":              {edit: func(c *app.CloudConnection) { c.Principal = "arn:aws:iam::123456789012:role/path/acme" }, wantRegion: "us-east-1"},
		"ecr with region":        {edit: func(c *app.CloudConnection) { c.Preset, c.DefaultRegion = app.CloudConnectionPresetECR, "us-west-2" }, wantRegion: "us-west-2"},
		"ecr requires region":    {edit: func(c *app.CloudConnection) { c.Preset = app.CloudConnectionPresetECR }, wantErr: "default_region is required"},
		"ecr rejects govcloud": {edit: func(c *app.CloudConnection) {
			c.Preset, c.DefaultRegion = app.CloudConnectionPresetECR, "us-gov-west-1"
		}, wantErr: "commercial AWS region"},
		"ecr rejects china":       {edit: func(c *app.CloudConnection) { c.Preset, c.DefaultRegion = app.CloudConnectionPresetECR, "cn-north-1" }, wantErr: "commercial AWS region"},
		"missing preset":          {edit: func(c *app.CloudConnection) { c.Preset = "" }, wantErr: "preset must be"},
		"gar preset":              {edit: func(c *app.CloudConnection) { c.Preset = app.CloudConnectionPresetGAR }, wantErr: "preset must be"},
		"short account":           {edit: func(c *app.CloudConnection) { c.TargetID = "12345678901" }, wantErr: "exactly 12 digits"},
		"role in another account": {edit: func(c *app.CloudConnection) { c.Principal = "arn:aws:iam::210987654321:role/acme" }, wantErr: "IAM role ARN in account"},
		"user instead of role":    {edit: func(c *app.CloudConnection) { c.Principal = "arn:aws:iam::123456789012:user/acme" }, wantErr: "IAM role ARN in account"},
		"wildcard role":           {edit: func(c *app.CloudConnection) { c.Principal = "arn:aws:iam::123456789012:role/acme*" }, wantErr: "IAM role ARN in account"},
		"empty role name":         {edit: func(c *app.CloudConnection) { c.Principal = "arn:aws:iam::123456789012:role/" }, wantErr: "IAM role ARN in account"},
		"regional principal":      {edit: func(c *app.CloudConnection) { c.Principal = "arn:aws:iam:us-east-1:123456789012:role/acme" }, wantErr: "IAM role ARN in account"},
		"non aws partition":       {edit: func(c *app.CloudConnection) { c.Principal = "arn:aws-cn:iam::123456789012:role/acme" }, wantErr: "IAM role ARN in account"},
	} {
		t.Run(name, func(t *testing.T) {
			connection := awsConnection()
			if tc.edit != nil {
				tc.edit(&connection)
			}
			err := validateAWS(&connection)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantRegion, connection.DefaultRegion)
		})
	}
}

func TestValidateGCP(t *testing.T) {
	for name, tc := range map[string]struct {
		edit    func(*app.CloudConnection)
		wantErr string
	}{
		"valid":                          {},
		"multi-region location":          {edit: func(c *app.CloudConnection) { c.DefaultRegion = "us" }},
		"ecr preset":                     {edit: func(c *app.CloudConnection) { c.Preset = app.CloudConnectionPresetECR }, wantErr: "GCP requires the gar preset"},
		"uppercase project":              {edit: func(c *app.CloudConnection) { c.TargetID = "Acme-Project" }, wantErr: "GCP project ID"},
		"short project":                  {edit: func(c *app.CloudConnection) { c.TargetID = "acme" }, wantErr: "GCP project ID"},
		"project ends with hyphen":       {edit: func(c *app.CloudConnection) { c.TargetID = "acme-project-" }, wantErr: "GCP project ID"},
		"missing project number":         {edit: func(c *app.CloudConnection) { c.GCPProjectNumber = "" }, wantErr: "numeric GCP project number"},
		"project number leading zero":    {edit: func(c *app.CloudConnection) { c.GCPProjectNumber = "012345678901" }, wantErr: "numeric GCP project number"},
		"service account other project":  {edit: func(c *app.CloudConnection) { c.Principal = "acme-pull@other-project.iam.gserviceaccount.com" }, wantErr: "service account email in project acme-project"},
		"service account name too short": {edit: func(c *app.CloudConnection) { c.Principal = "acme@acme-project.iam.gserviceaccount.com" }, wantErr: "service account email in project"},
		"user email":                     {edit: func(c *app.CloudConnection) { c.Principal = "acme@example.com" }, wantErr: "service account email in project"},
		"missing region":                 {edit: func(c *app.CloudConnection) { c.DefaultRegion = "" }, wantErr: "GAR region"},
		"uppercase region":               {edit: func(c *app.CloudConnection) { c.DefaultRegion = "US-CENTRAL1" }, wantErr: "GAR region"},
	} {
		t.Run(name, func(t *testing.T) {
			connection := gcpConnection()
			if tc.edit != nil {
				tc.edit(&connection)
			}
			err := validateGCP(&connection)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateAzure(t *testing.T) {
	for name, tc := range map[string]struct {
		edit    func(*app.CloudConnection)
		wantErr string
	}{
		"valid": {},
		"uppercase subscription in arm id": {edit: func(c *app.CloudConnection) {
			c.TargetID = "abcdef12-1111-1111-1111-111111111111"
			c.AzureRegistryID = "/subscriptions/ABCDEF12-1111-1111-1111-111111111111/resourceGroups/acme/providers/Microsoft.ContainerRegistry/registries/acmeregistry"
		}},
		"stacks preset":         {edit: func(c *app.CloudConnection) { c.Preset = app.CloudConnectionPresetStacks }, wantErr: "Azure requires the acr preset"},
		"subscription not uuid": {edit: func(c *app.CloudConnection) { c.TargetID = "acme" }, wantErr: "subscription UUID"},
		"principal not uuid":    {edit: func(c *app.CloudConnection) { c.Principal = "acme-app" }, wantErr: "client ID UUID"},
		"tenant not uuid":       {edit: func(c *app.CloudConnection) { c.AzureTenantID = "acme.onmicrosoft.com" }, wantErr: "tenant UUID"},
		"registry in other subscription": {edit: func(c *app.CloudConnection) {
			c.AzureRegistryID = "/subscriptions/" + azureTenant + "/resourceGroups/acme/providers/Microsoft.ContainerRegistry/registries/acmeregistry"
		}, wantErr: "ARM ID in subscription " + azureSubscription},
		"resource group ends with dot": {edit: func(c *app.CloudConnection) {
			c.AzureRegistryID = "/subscriptions/" + azureSubscription + "/resourceGroups/acme./providers/Microsoft.ContainerRegistry/registries/acmeregistry"
		}, wantErr: "ARM ID"},
		"not a registry": {edit: func(c *app.CloudConnection) {
			c.AzureRegistryID = "/subscriptions/" + azureSubscription + "/resourceGroups/acme/providers/Microsoft.Storage/storageAccounts/acme"
		}, wantErr: "ARM ID"},
		"uppercase login server": {edit: func(c *app.CloudConnection) { c.AzureRegistryLoginServer = "AcmeRegistry.azurecr.io" }, wantErr: "lowercase"},
		"foreign login server":   {edit: func(c *app.CloudConnection) { c.AzureRegistryLoginServer = "registry.example.com" }, wantErr: "azurecr.io"},
	} {
		t.Run(name, func(t *testing.T) {
			connection := azureConnection()
			if tc.edit != nil {
				tc.edit(&connection)
			}
			err := validateAzure(&connection)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
