package models

// AppPolicyReport represents a policy evaluation report.
// swagger:model app.PolicyReport
type AppPolicyReport struct {
	ID          string `json:"id,omitempty"`
	CreatedByID string `json:"created_by_id,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`

	OrgID       string `json:"org_id,omitempty"`
	AppID       string `json:"app_id,omitempty"`
	InstallID   string `json:"install_id,omitempty"`
	ComponentID string `json:"component_id,omitempty"`

	OwnerID   string `json:"owner_id,omitempty"`
	OwnerType string `json:"owner_type,omitempty"`

	EvaluatedAt string                `json:"evaluated_at,omitempty"`
	Violations  []*AppPolicyViolation `json:"violations,omitempty"`
	PolicyIDs   []string              `json:"policy_ids,omitempty"`
	Policies    []*AppPolicyResult    `json:"policies,omitempty"`
	Inputs      []*AppPolicyInputRef  `json:"inputs,omitempty"`
	Status      *AppCompositeStatus   `json:"status,omitempty"`

	DenyCount int `json:"deny_count,omitempty"`
	WarnCount int `json:"warn_count,omitempty"`
	PassCount int `json:"pass_count,omitempty"`
}

// AppPolicyViolation represents a single policy violation.
type AppPolicyViolation struct {
	PolicyID      string `json:"policy_id,omitempty"`
	InputIndex    int    `json:"input_index,omitempty"`
	InputIdentity string `json:"input_identity,omitempty"`
	Message       string `json:"message,omitempty"`
	Severity      string `json:"severity,omitempty"`
}

// AppPolicyResult represents the evaluation result for a single policy.
type AppPolicyResult struct {
	PolicyID   string `json:"policy_id,omitempty"`
	Status     string `json:"status,omitempty"`
	DenyCount  int    `json:"deny_count,omitempty"`
	WarnCount  int    `json:"warn_count,omitempty"`
	PassCount  int    `json:"pass_count,omitempty"`
	InputCount int    `json:"input_count,omitempty"`
}

// AppPolicyInputRef represents a reference to an input that was evaluated.
type AppPolicyInputRef struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}
