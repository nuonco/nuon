package permissions

const (
	RolesRead                PermissionID = "role.read"
	ServiceAccountsRead      PermissionID = "service_account.read"
	ServiceAccountCreate     PermissionID = "service_account.create"
	ServiceAccountUpdate     PermissionID = "service_account.update"
	ServiceAccountRoleUpdate PermissionID = "service_account.role.update"
	ServiceAccountDelete     PermissionID = "service_account.delete"
	TokensRead               PermissionID = "token.read"
	TokenCreate              PermissionID = "token.create"
	TokenDelete              PermissionID = "token.delete"
	TrustPoliciesRead        PermissionID = "trust_policy.read"
	TrustPolicyCreate        PermissionID = "trust_policy.create"
	TrustPolicyUpdate        PermissionID = "trust_policy.update"
	TrustPolicyDelete        PermissionID = "trust_policy.delete"
)

func AccessDefinitions() []PermissionDefinition {
	org := []ResourceKind{KindOrg}
	return []PermissionDefinition{
		{ID: RolesRead, Title: "Read roles", Description: "List the org's roles and assignment contexts.", Scopes: org, Targets: org},
		{ID: ServiceAccountsRead, Title: "Read service accounts", Description: "List org service accounts, their roles and managed ownership metadata; formerly requires org admin.", Scopes: org, Targets: org},
		{ID: ServiceAccountCreate, Title: "Create service accounts", Description: "Create service accounts with any existing assignable org role, including administrator roles.", Scopes: org, Targets: org},
		{ID: ServiceAccountUpdate, Title: "Update service accounts", Description: "Rename user-managed service accounts in the org.", Scopes: org, Targets: org},
		{ID: ServiceAccountRoleUpdate, Title: "Assign service account roles", Description: "Replace a user-managed service account's assignable org role, including administrator roles.", Scopes: org, Targets: org},
		{ID: ServiceAccountDelete, Title: "Remove service accounts", Description: "Remove user-managed service accounts from the org and invalidate their existing credentials.", Scopes: org, Targets: org},
		{ID: TokensRead, Title: "Read API token metadata", Description: "List org API token metadata without token secrets; formerly requires org admin.", Scopes: org, Targets: org},
		{ID: TokenCreate, Title: "Issue API tokens", Description: "Issue credentials for org service accounts, including administrator accounts and the deprecated personal-token flow.", Scopes: org, Targets: org},
		{ID: TokenDelete, Title: "Revoke API tokens", Description: "Revoke org API tokens, including dedicated-account cleanup on the deprecated endpoint.", Scopes: org, Targets: org},
		{ID: TrustPoliciesRead, Title: "Read federation trust policies", Description: "Read org OIDC issuer, subject and role configuration; formerly requires org admin.", Scopes: org, Targets: org},
		{ID: TrustPolicyCreate, Title: "Create federation trust policies", Description: "Authorize external subjects to obtain credentials with an assignable org role.", Scopes: org, Targets: org},
		{ID: TrustPolicyUpdate, Title: "Update federation trust policies", Description: "Change external subjects, issuers and assignable roles for org federation.", Scopes: org, Targets: org},
		{ID: TrustPolicyDelete, Title: "Delete federation trust policies", Description: "Remove org federation trust and revoke its credentials.", Scopes: org, Targets: org},
	}
}
