package roles

func DefaultTags(orgID string) [][2]string {
	return [][2]string{
		{"org-id", orgID},
		{"managed-by", "workers-orgs"},
	}
}

type iamRolePolicy struct {
	Version   string             `json:"Version"`
	Statement []iamRoleStatement `json:"Statement"`
}

type iamRoleStatement struct {
	Action    []string     `json:"Action,omitempty"`
	Effect    string       `json:"Effect,omitempty"`
	Resource  string       `json:"Resource,omitempty"`
	Sid       string       `json:"Sid"`
	Condition iamCondition `json:"Condition,omitempty"`
}

type iamRoleTrustPolicy struct {
	Version   string                  `json:"Version"`
	Statement []iamRoleTrustStatement `json:"Statement"`
}

type iamPrincipal struct {
	Federated string   `json:"Federated,omitempty"`
	AWS       []string `json:"AWS,omitempty"`
}

type iamCondition struct {
	StringEquals map[string]string `json:"StringEquals,omitempty"`
	StringLike   map[string]string `json:"StringLike,omitempty"`
}

type iamRoleTrustStatement struct {
	Action    []string     `json:"Action,omitempty"`
	Effect    string       `json:"Effect,omitempty"`
	Resource  string       `json:"Resource,omitempty"`
	Sid       string       `json:"Sid"`
	Principal iamPrincipal `json:"Principal,omitempty"`
	Condition iamCondition `json:"Condition,omitempty"`
}
