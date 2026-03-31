package orgiam

import (
	"context"
	"errors"
	"fmt"
	"time"

	crm "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

type CreateGCPServiceAccountRequest struct {
	ProjectID             string `validate:"required"`
	OrgID                 string `validate:"required"`
	K8sNamespace          string // Runner ID used as namespace; empty skips WI binding
	K8sServiceAccountName string // K8s SA name; empty skips WI binding
}

type CreateGCPServiceAccountResponse struct {
	ServiceAccountEmail string
}

// @temporal-gen-v2 activity
// @schedule-to-close-timeout 5m
// @start-to-close-timeout 5m
// @max-retries 3
func (a *Activities) CreateGCPServiceAccount(ctx context.Context, req *CreateGCPServiceAccountRequest) (*CreateGCPServiceAccountResponse, error) {
	if err := a.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	iamService, err := iam.NewService(ctx, option.WithScopes(iam.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("unable to create IAM service: %w", err)
	}

	saName := truncateGCPServiceAccountID(req.OrgID)
	saEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", saName, req.ProjectID)
	projectResource := fmt.Sprintf("projects/%s", req.ProjectID)

	// Create the service account. 409 means it already exists — that's fine.
	_, createErr := iamService.Projects.ServiceAccounts.Create(projectResource, &iam.CreateServiceAccountRequest{
		AccountId: saName,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: fmt.Sprintf("Nuon org runner %s", req.OrgID),
		},
	}).Context(ctx).Do()
	if createErr != nil && !isGoogleAPIError(createErr, 409) {
		return nil, fmt.Errorf("unable to create service account: %w", createErr)
	}
	// GCP IAM is eventually consistent — brief pause before setting bindings.
	if createErr == nil {
		time.Sleep(5 * time.Second)
	}

	// Add Workload Identity binding if K8s namespace and SA name are provided.
	// On reprovision the runner ID may not be available; the binding was already created during initial provision.
	if req.K8sNamespace != "" && req.K8sServiceAccountName != "" {
		wiMember := fmt.Sprintf(
			"serviceAccount:%s.svc.id.goog[%s/%s]",
			req.ProjectID, req.K8sNamespace, req.K8sServiceAccountName,
		)

		saResource := fmt.Sprintf("projects/%s/serviceAccounts/%s", req.ProjectID, saEmail)
		policy, err := iamService.Projects.ServiceAccounts.GetIamPolicy(saResource).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("unable to get IAM policy for service account: %w", err)
		}

		wiRole := "roles/iam.workloadIdentityUser"
		bindingExists := false
		for _, binding := range policy.Bindings {
			if binding.Role == wiRole {
				for _, member := range binding.Members {
					if member == wiMember {
						bindingExists = true
						break
					}
				}
				if !bindingExists {
					binding.Members = append(binding.Members, wiMember)
					bindingExists = true
				}
				break
			}
		}
		if !bindingExists {
			policy.Bindings = append(policy.Bindings, &iam.Binding{
				Role:    wiRole,
				Members: []string{wiMember},
			})
		}

		_, err = iamService.Projects.ServiceAccounts.SetIamPolicy(saResource, &iam.SetIamPolicyRequest{
			Policy: policy,
		}).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("unable to set Workload Identity binding: %w", err)
		}
	}

	// Grant the service account Artifact Registry writer at the project level
	// so the org runner can push and pull build artifacts from GAR.
	if err := addProjectIAMBinding(ctx, req.ProjectID, "roles/artifactregistry.writer", fmt.Sprintf("serviceAccount:%s", saEmail)); err != nil {
		return nil, fmt.Errorf("unable to grant GAR writer: %w", err)
	}

	return &CreateGCPServiceAccountResponse{
		ServiceAccountEmail: saEmail,
	}, nil
}

func addProjectIAMBinding(ctx context.Context, projectID, role, member string) error {
	crmService, err := crm.NewService(ctx, option.WithScopes(crm.CloudPlatformScope))
	if err != nil {
		return fmt.Errorf("unable to create CRM service: %w", err)
	}

	policy, err := crmService.Projects.GetIamPolicy(projectID, &crm.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("unable to get project IAM policy: %w", err)
	}

	// Check if binding already exists
	for _, binding := range policy.Bindings {
		if binding.Role == role {
			for _, m := range binding.Members {
				if m == member {
					return nil // already bound
				}
			}
			binding.Members = append(binding.Members, member)
			_, err := crmService.Projects.SetIamPolicy(projectID, &crm.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
			return err
		}
	}

	policy.Bindings = append(policy.Bindings, &crm.Binding{
		Role:    role,
		Members: []string{member},
	})
	_, err = crmService.Projects.SetIamPolicy(projectID, &crm.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	return err
}

// truncateGCPServiceAccountID truncates an org ID to fit within GCP's 30-char
// service account ID limit (must be 6-30 chars, match [a-z][a-z0-9-]{4,28}[a-z0-9]).
func truncateGCPServiceAccountID(orgID string) string {
	if len(orgID) <= 30 {
		return orgID
	}
	return orgID[:30]
}

func isGoogleAPIError(err error, code int) bool {
	if err == nil {
		return false
	}
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.Code == code
	}
	return false
}
