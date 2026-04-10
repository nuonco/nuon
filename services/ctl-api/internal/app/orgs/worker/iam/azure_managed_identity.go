package orgiam

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"
	"github.com/google/uuid"
)

type CreateAzureManagedIdentityRequest struct {
	SubscriptionID        string `validate:"required"`
	ResourceGroup         string `validate:"required"`
	TenantID              string `validate:"required"`
	OrgID                 string `validate:"required"`
	Location              string `validate:"required"`
	AKSOIDCIssuerURL      string `validate:"required"`
	K8sNamespace          string // Runner ID used as namespace; empty skips federated credential
	K8sServiceAccountName string // K8s SA name; empty skips federated credential
	ACRResourceID         string `validate:"required"`
}

type CreateAzureManagedIdentityResponse struct {
	ClientID string
}

// @temporal-gen-v2 activity
// @schedule-to-close-timeout 5m
// @start-to-close-timeout 5m
// @max-retries 3
func (a *Activities) CreateAzureManagedIdentity(ctx context.Context, req *CreateAzureManagedIdentityRequest) (*CreateAzureManagedIdentityResponse, error) {
	if err := a.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	cred, err := azureCredential()
	if err != nil {
		return nil, fmt.Errorf("unable to create Azure credential: %w", err)
	}

	// Step 1: Create User-Assigned Managed Identity.
	miClient, err := armmsi.NewUserAssignedIdentitiesClient(req.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create managed identity client: %w", err)
	}

	miName := fmt.Sprintf("runner-%s", req.OrgID)
	mi, err := miClient.CreateOrUpdate(ctx, req.ResourceGroup, miName, armmsi.Identity{
		Location: &req.Location,
		Tags: map[string]*string{
			"nuon-org-id": &req.OrgID,
			"managed-by":  to.Ptr("nuon"),
		},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create managed identity: %w", err)
	}

	// Step 2: Create Federated Identity Credential (only if K8s info provided).
	if req.K8sNamespace != "" && req.K8sServiceAccountName != "" {
		ficClient, err := armmsi.NewFederatedIdentityCredentialsClient(req.SubscriptionID, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to create federated identity credentials client: %w", err)
		}

		ficName := fmt.Sprintf("runner-%s", req.OrgID)
		subject := fmt.Sprintf("system:serviceaccount:%s:%s", req.K8sNamespace, req.K8sServiceAccountName)
		audiences := []*string{to.Ptr("api://AzureADTokenExchange")}
		_, err = ficClient.CreateOrUpdate(ctx, req.ResourceGroup, miName, ficName, armmsi.FederatedIdentityCredential{
			Properties: &armmsi.FederatedIdentityCredentialProperties{
				Issuer:    &req.AKSOIDCIssuerURL,
				Subject:   &subject,
				Audiences: audiences,
			},
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to create federated identity credential: %w", err)
		}
	}

	// Step 3: Assign AcrPush role on the ACR.
	roleAssignmentClient, err := armauthorization.NewRoleAssignmentsClient(req.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create role assignments client: %w", err)
	}

	// AcrPush built-in role definition ID.
	acrPushRoleID := fmt.Sprintf(
		"/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/8311e382-0749-4cb8-b61a-304f252e45ec",
		req.SubscriptionID,
	)
	principalID := *mi.Properties.PrincipalID
	assignmentName := uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("nuon-acr-push-%s", req.OrgID))).String()

	_, err = roleAssignmentClient.Create(ctx, req.ACRResourceID, assignmentName, armauthorization.RoleAssignmentCreateParameters{
		Properties: &armauthorization.RoleAssignmentProperties{
			RoleDefinitionID: &acrPushRoleID,
			PrincipalID:      &principalID,
			PrincipalType:    to.Ptr(armauthorization.PrincipalTypeServicePrincipal),
		},
	}, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == http.StatusConflict {
			// Role assignment already exists — idempotent success.
		} else {
			return nil, fmt.Errorf("unable to assign AcrPush role: %w", err)
		}
	}

	return &CreateAzureManagedIdentityResponse{
		ClientID: *mi.Properties.ClientID,
	}, nil
}

type DeleteAzureManagedIdentityRequest struct {
	SubscriptionID string `validate:"required"`
	ResourceGroup  string `validate:"required"`
	OrgID          string `validate:"required"`
}

type DeleteAzureManagedIdentityResponse struct{}

// @temporal-gen-v2 activity
// @schedule-to-close-timeout 5m
// @start-to-close-timeout 5m
// @max-retries 3
func (a *Activities) DeleteAzureManagedIdentity(ctx context.Context, req *DeleteAzureManagedIdentityRequest) (*DeleteAzureManagedIdentityResponse, error) {
	if err := a.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	cred, err := azureCredential()
	if err != nil {
		return nil, fmt.Errorf("unable to create Azure credential: %w", err)
	}

	miClient, err := armmsi.NewUserAssignedIdentitiesClient(req.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create managed identity client: %w", err)
	}

	miName := fmt.Sprintf("runner-%s", req.OrgID)
	_, err = miClient.Delete(ctx, req.ResourceGroup, miName, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == http.StatusNotFound {
			// Already deleted — idempotent success.
		} else {
			return nil, fmt.Errorf("unable to delete managed identity: %w", err)
		}
	}

	return &DeleteAzureManagedIdentityResponse{}, nil
}

// azureCredential returns an Azure token credential. In development it uses
// the CLI credential so that local runs work without managed identity.
func azureCredential() (azcore.TokenCredential, error) {
	if os.Getenv("ENV") == "development" {
		return azidentity.NewAzureCLICredential(nil)
	}
	return azidentity.NewDefaultAzureCredential(nil)
}
