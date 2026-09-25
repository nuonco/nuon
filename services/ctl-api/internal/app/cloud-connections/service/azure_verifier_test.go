package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAzureStackRolesSufficient(t *testing.T) {
	const (
		principal = "00000000-0000-0000-0000-000000000001"
		scope     = "/subscriptions/00000000-0000-0000-0000-000000000002"
	)
	role := func(roleID, assignmentScope string) azureRoleAssignment {
		return azureRoleAssignment{PrincipalID: principal, RoleDefinitionID: scope + "/providers/Microsoft.Authorization/roleDefinitions/" + roleID, Scope: assignmentScope}
	}
	tests := map[string]struct {
		assignments []azureRoleAssignment
		want        bool
	}{
		"Owner alone passes":              {assignments: []azureRoleAssignment{role(azureOwnerRoleID, scope)}, want: true},
		"Contributor alone fails":         {assignments: []azureRoleAssignment{role(azureContributorRoleID, scope)}},
		"Contributor plus UAA passes":     {assignments: []azureRoleAssignment{role(azureContributorRoleID, scope), role(azureUserAccessRoleID, scope)}, want: true},
		"role on a different scope fails": {assignments: []azureRoleAssignment{role(azureOwnerRoleID, scope+"/resourceGroups/acme")}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, azureStackRolesSufficient(test.assignments, principal, scope))
		})
	}
}
