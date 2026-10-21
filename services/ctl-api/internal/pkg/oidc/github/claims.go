package github

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc"

// Claims are the documented claims of a GitHub Actions OIDC token.
type Claims struct {
	oidc.Claims `json:",squash"`

	Actor                string `json:"actor"`
	ActorID              string `json:"actor_id"`
	BaseRef              string `json:"base_ref"`
	Environment          string `json:"environment"`
	EnvironmentNodeID    string `json:"environment_node_id"`
	EventName            string `json:"event_name"`
	HeadRef              string `json:"head_ref"`
	JobWorkflowRef       string `json:"job_workflow_ref"`
	JobWorkflowSHA       string `json:"job_workflow_sha"`
	Ref                  string `json:"ref"`
	RefProtected         string `json:"ref_protected"`
	RefType              string `json:"ref_type"`
	Repository           string `json:"repository"`
	RepositoryID         string `json:"repository_id"`
	RepositoryOwner      string `json:"repository_owner"`
	RepositoryOwnerID    string `json:"repository_owner_id"`
	RepositoryVisibility string `json:"repository_visibility"`
	RunAttempt           string `json:"run_attempt"`
	RunID                string `json:"run_id"`
	RunNumber            string `json:"run_number"`
	RunnerEnvironment    string `json:"runner_environment"`
	SHA                  string `json:"sha"`
	Workflow             string `json:"workflow"`
	WorkflowRef          string `json:"workflow_ref"`
	WorkflowSHA          string `json:"workflow_sha"`
	Enterprise           string `json:"enterprise"`
	EnterpriseID         string `json:"enterprise_id"`
}

// DecodeClaims decodes a raw GitHub Actions claim set.
func DecodeClaims(raw map[string]any) (Claims, error) {
	var claims Claims
	if err := oidc.Decode(raw, &claims); err != nil {
		return Claims{}, err
	}

	return claims, nil
}
