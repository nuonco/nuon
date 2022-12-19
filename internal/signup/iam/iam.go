package iam

func defaultTags(orgID string) [][2]string {
	return [][2]string{{"managed-by", "workers-orgs"}, {"org-id", orgID}}
}

type iamRolePolicy struct {
	Version   string             `json:"Version"`
	Statement []iamRoleStatement `json:"Statement"`
}

type iamRoleStatement struct {
	Action   []string `json:"Action,omitempty"`
	Effect   string   `json:"Effect,omitempty"`
	Resource string   `json:"Resource,omitempty"`
	Sid      string   `json:"Sid"`
}

type iamRoleTrustPolicy struct {
	Version   string                  `json:"Version"`
	Statement []iamRoleTrustStatement `json:"Statement"`
}

type iamRoleTrustStatement struct {
	Action    string `json:"Action,omitempty"`
	Effect    string `json:"Effect,omitempty"`
	Resource  string `json:"Resource,omitempty"`
	Sid       string `json:"Sid"`
	Principal struct {
		Federated string `json:"Federated,omitempty"`
	} `json:"Principal,omitempty"`
	Condition struct {
		StringEquals map[string]string `json:"StringEquals"`
	} `json:"Condition,omitempty"`
}
