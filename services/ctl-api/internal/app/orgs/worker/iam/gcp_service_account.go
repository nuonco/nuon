package orgiam

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	artifactregistry "google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

type CreateGCPServiceAccountRequest struct {
	ProjectID             string `validate:"required"`
	OrgID                 string `validate:"required"`
	GARRepositoryURL      string `validate:"required"`
	K8sNamespace          string // Runner ID used as namespace; empty skips WI binding
	K8sServiceAccountName string // K8s SA name; empty skips WI binding
	// SharedServiceAccountEmail switches to the shared org-runner SA: no SA is
	// created and no GAR grant is made (both happen at install-stack /
	// component-deploy time); only the per-org WI binding is appended.
	SharedServiceAccountEmail string
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

	saEmail := req.SharedServiceAccountEmail
	if saEmail == "" {
		saName := truncateGCPServiceAccountID(req.OrgID)
		saEmail = fmt.Sprintf("%s@%s.iam.gserviceaccount.com", saName, req.ProjectID)
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

	// The shared SA carries its GAR access from deploy-time bindings.
	if req.SharedServiceAccountEmail != "" {
		return &CreateGCPServiceAccountResponse{
			ServiceAccountEmail: saEmail,
		}, nil
	}

	// Grant the service account Artifact Registry writer on the management
	// repository only. A project-level grant would require ctl-api to hold
	// projectIamAdmin, a one-hop escalation to owner.
	if err := addGARRepositoryIAMBinding(ctx, req.GARRepositoryURL, "roles/artifactregistry.writer", fmt.Sprintf("serviceAccount:%s", saEmail)); err != nil {
		return nil, fmt.Errorf("unable to grant GAR writer: %w", err)
	}

	return &CreateGCPServiceAccountResponse{
		ServiceAccountEmail: saEmail,
	}, nil
}

// garRepositoryResource converts a repository URL like
// "us-west1-docker.pkg.dev/my-project/my-repo" into the API resource name
// "projects/my-project/locations/us-west1/repositories/my-repo".
func garRepositoryResource(repoURL string) (string, error) {
	parts := strings.Split(strings.TrimSuffix(repoURL, "/"), "/")
	if len(parts) < 3 || !strings.HasSuffix(parts[0], "-docker.pkg.dev") {
		return "", fmt.Errorf("unexpected GAR repository URL %q", repoURL)
	}
	location := strings.TrimSuffix(parts[0], "-docker.pkg.dev")
	return fmt.Sprintf("projects/%s/locations/%s/repositories/%s", parts[1], location, parts[2]), nil
}

func addGARRepositoryIAMBinding(ctx context.Context, repoURL, role, member string) error {
	resource, err := garRepositoryResource(repoURL)
	if err != nil {
		return err
	}

	arService, err := artifactregistry.NewService(ctx, option.WithScopes(artifactregistry.CloudPlatformScope))
	if err != nil {
		return fmt.Errorf("unable to create Artifact Registry service: %w", err)
	}

	repos := arService.Projects.Locations.Repositories
	policy, err := repos.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("unable to get repository IAM policy: %w", err)
	}

	for _, binding := range policy.Bindings {
		if binding.Role == role {
			for _, m := range binding.Members {
				if m == member {
					return nil
				}
			}
			binding.Members = append(binding.Members, member)
			_, err := repos.SetIamPolicy(resource, &artifactregistry.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
			return err
		}
	}

	policy.Bindings = append(policy.Bindings, &artifactregistry.Binding{
		Role:    role,
		Members: []string{member},
	})
	_, err = repos.SetIamPolicy(resource, &artifactregistry.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
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
