package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"

	"github.com/nuonco/nuon/pkg/azure/acr"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnectionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

const (
	azureOwnerRoleID        = "8e3af657-a8ff-443c-a75c-2fe8c4bcb635"
	azureContributorRoleID  = "b24988ac-6180-42a0-ab88-20f7382dd24c"
	azureUserAccessRoleID   = "18d7d88d-d35e-4c88-9ee4-2d2a4a5008b3"
	azureManagementScope    = "https://management.azure.com/.default"
	azureManagementEndpoint = "https://management.azure.com"
)

type VerifyOptions struct {
	Repositories []string
	Registry     string
}

type cloudVerifier struct {
	aws   Verifier
	azure Verifier
	gcp   Verifier
}

func NewCloudVerifier(issuer *oidcissuer.Issuer) Verifier {
	return &cloudVerifier{aws: NewAWSVerifier(issuer), azure: &azureVerifier{issuer: issuer, client: http.DefaultClient}, gcp: &gcpVerifier{issuer: issuer, client: http.DefaultClient}}
}

func (v *cloudVerifier) Verify(ctx context.Context, connection *app.CloudConnection, options VerifyOptions) (VerificationResult, error) {
	switch connection.Platform {
	case app.CloudPlatformAWS:
		return v.aws.Verify(ctx, connection, options)
	case app.CloudPlatformAzure:
		return v.azure.Verify(ctx, connection, options)
	case app.CloudPlatformGCP:
		return v.gcp.Verify(ctx, connection, options)
	default:
		return VerificationResult{}, fmt.Errorf("unsupported cloud platform %q", connection.Platform)
	}
}

type azureVerifier struct {
	issuer *oidcissuer.Issuer
	client *http.Client
}

type azureTokenClaims struct {
	AppID string `json:"appid"`
	OID   string `json:"oid"`
}

type azureRoleAssignment struct {
	PrincipalID      string
	RoleDefinitionID string
	Scope            string
}

func (v *azureVerifier) Verify(ctx context.Context, connection *app.CloudConnection, options VerifyOptions) (VerificationResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	credential, err := cloudconnectionshelpers.AzureCredential(v.issuer, connection)
	if err != nil {
		return VerificationResult{}, err
	}
	token, err := credential.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{azureManagementScope}})
	if err != nil {
		return verificationFailure("Nuon OIDC identity is not trusted by this Entra application."), nil
	}
	claims, err := decodeAzureTokenClaims(token.Token)
	if err != nil {
		return VerificationResult{}, err
	}
	if !strings.EqualFold(claims.AppID, connection.Principal) || claims.OID == "" {
		return verificationFailure("The exchanged identity does not match the configured application."), nil
	}
	if err := v.probeSubscription(ctx, token.Token, connection.TargetID); err != nil {
		return verificationFailure("The application cannot access the configured Azure subscription."), nil
	}
	if err := v.negativeProbe(ctx, connection); err != nil {
		return VerificationResult{}, err
	}

	capabilities := make([]app.CloudConnectionCapability, 0, len(connection.Capabilities))
	for _, capability := range connection.Capabilities {
		switch capability {
		case app.CloudConnectionCapabilityStacks:
			ok, err := v.probeStackRoles(ctx, credential, connection.TargetID, claims.OID)
			if err != nil {
				return VerificationResult{}, fmt.Errorf("probe stacks capability: %w", err)
			}
			if ok {
				capabilities = append(capabilities, capability)
			}
		case app.CloudConnectionCapabilityImages:
			if options.Registry == "" {
				continue
			}
			if err := v.probeRegistry(ctx, credential, connection.TenantID, options.Registry, options.Repositories); err == nil {
				capabilities = append(capabilities, capability)
			}
		}
	}
	if len(capabilities) == 0 {
		return VerificationResult{Status: app.CloudConnectionStatusError, Message: "The application lacks all requested capabilities.", Capabilities: capabilities}, nil
	}
	return VerificationResult{Status: app.CloudConnectionStatusVerified, Message: "Cloud connection verified.", Capabilities: capabilities}, nil
}

func decodeAzureTokenClaims(token string) (azureTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return azureTokenClaims{}, fmt.Errorf("Azure access token is not a JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return azureTokenClaims{}, fmt.Errorf("decode Azure access token: %w", err)
	}
	var claims azureTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return azureTokenClaims{}, fmt.Errorf("parse Azure access token: %w", err)
	}
	return claims, nil
}

func (v *azureVerifier) probeSubscription(ctx context.Context, token, subscriptionID string) error {
	url := fmt.Sprintf("%s/subscriptions/%s?api-version=2022-12-01", azureManagementEndpoint, subscriptionID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	response, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("subscription probe returned %s", response.Status)
	}
	return nil
}

func (v *azureVerifier) negativeProbe(ctx context.Context, connection *app.CloudConnection) error {
	credential, err := azidentity.NewClientAssertionCredential(connection.TenantID, connection.Principal, func(ctx context.Context) (string, error) {
		return v.issuer.Mint(ctx, "org:foreign:connection:foreign", azureTokenExchangeAudience, 10*time.Minute)
	}, nil)
	if err != nil {
		return err
	}
	_, err = credential.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{azureManagementScope}})
	if err == nil {
		return fmt.Errorf("the Entra application accepts a foreign Nuon connection subject")
	}
	message := err.Error()
	if !strings.Contains(message, "AADSTS70021") && !strings.Contains(message, "AADSTS700213") {
		return fmt.Errorf("probe foreign subject: %w", err)
	}
	return nil
}

func (v *azureVerifier) probeStackRoles(ctx context.Context, credential azcore.TokenCredential, subscriptionID, principalID string) (bool, error) {
	client, err := armauthorization.NewRoleAssignmentsClient(subscriptionID, credential, nil)
	if err != nil {
		return false, err
	}
	scope := "/subscriptions/" + subscriptionID
	filter := fmt.Sprintf("principalId eq '%s'", principalID)
	pager := client.NewListForScopePager(scope, &armauthorization.RoleAssignmentsClientListForScopeOptions{Filter: &filter})
	assignments := make([]azureRoleAssignment, 0)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return false, err
		}
		for _, assignment := range page.Value {
			if assignment.Properties == nil {
				continue
			}
			assignments = append(assignments, azureRoleAssignment{PrincipalID: value(assignment.Properties.PrincipalID), RoleDefinitionID: value(assignment.Properties.RoleDefinitionID), Scope: value(assignment.Properties.Scope)})
		}
	}
	return azureStackRolesSufficient(assignments, principalID, scope), nil
}

func azureStackRolesSufficient(assignments []azureRoleAssignment, principalID, scope string) bool {
	owner, contributor, userAccess := false, false, false
	for _, assignment := range assignments {
		if !strings.EqualFold(assignment.PrincipalID, principalID) || !strings.EqualFold(strings.TrimSuffix(assignment.Scope, "/"), strings.TrimSuffix(scope, "/")) {
			continue
		}
		roleID := assignment.RoleDefinitionID[strings.LastIndex(assignment.RoleDefinitionID, "/")+1:]
		switch {
		case strings.EqualFold(roleID, azureOwnerRoleID):
			owner = true
		case strings.EqualFold(roleID, azureContributorRoleID):
			contributor = true
		case strings.EqualFold(roleID, azureUserAccessRoleID):
			userAccess = true
		}
	}
	return owner || contributor && userAccess
}

func (v *azureVerifier) probeRegistry(ctx context.Context, credential azcore.TokenCredential, tenantID, registry string, repositories []string) error {
	host := registry
	if !strings.Contains(host, ".") {
		host += ".azurecr.io"
	}
	refreshToken, err := acr.GetRepositoryTokenWithCredential(ctx, credential, tenantID, host)
	if err != nil {
		return err
	}
	paths := []string{"/v2/_catalog"}
	if len(repositories) > 0 {
		paths = paths[:0]
		for _, repository := range repositories {
			paths = append(paths, "/v2/"+repository+"/tags/list")
		}
	}
	for _, path := range paths {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+host+path, nil)
		if err != nil {
			return err
		}
		req.SetBasicAuth(acr.DefaultACRUsername, refreshToken)
		response, err := v.client.Do(req)
		if err != nil {
			return err
		}
		_, _ = io.Copy(io.Discard, response.Body)
		response.Body.Close()
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return fmt.Errorf("ACR probe returned %s", response.Status)
		}
	}
	return nil
}

func value(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
