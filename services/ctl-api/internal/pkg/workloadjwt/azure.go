package workloadjwt

import (
	"fmt"
	"regexp"
	"strings"
)

// why: Graph looks like the safer audience -- a token for a role-less identity can do nothing
// -- but Microsoft signs Graph access tokens with a key that is not in the tenant JWKS, so
// a third party cannot verify them. ARM tokens are ordinary v1 tokens and verify against
// the tenant keys, which is why runner auth already uses this audience.
//
// The identity is created with no role assignments, so an ARM token minted for it is
// authorized for nothing either.
const AzureManagementAudience = "https://management.azure.com/"

// why: Applied before a tenant or subscription is interpolated into a URL or compared, so a
// claim can never reshape the URL it lands in.
var azureGUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func AzureIssuer(tenantID string) (string, error) {
	if !azureGUID.MatchString(tenantID) {
		return "", fmt.Errorf("tenant id %q is not a guid", tenantID)
	}

	return fmt.Sprintf("https://sts.windows.net/%s/", tenantID), nil
}

type AzureManagedIdentity struct {
	SubscriptionID string
	ResourceGroup  string
	Name           string
	PrincipalID    string
	TenantID       string
}

const azureMIRIDSegments = 8

func ParseAzureManagedIdentity(claims map[string]any) (*AzureManagedIdentity, error) {
	tenantID, ok := StringClaim(claims, "tid")
	if !ok {
		return nil, fmt.Errorf("token is missing the tid claim")
	}
	if !azureGUID.MatchString(tenantID) {
		return nil, fmt.Errorf("tid claim %q is not a guid", tenantID)
	}

	principalID, ok := StringClaim(claims, "oid")
	if !ok {
		return nil, fmt.Errorf("token is missing the oid claim")
	}

	mirID, ok := StringClaim(claims, "xms_mirid")
	if !ok {
		return nil, fmt.Errorf("token is missing the xms_mirid claim")
	}

	parts := strings.Split(strings.TrimPrefix(mirID, "/"), "/")
	if len(parts) != azureMIRIDSegments {
		return nil, fmt.Errorf("xms_mirid is not a managed identity resource id")
	}

	for i, want := range map[int]string{
		0: "subscriptions",
		2: "resourcegroups",
		4: "providers",
		5: "Microsoft.ManagedIdentity",
		6: "userAssignedIdentities",
	} {
		if !strings.EqualFold(parts[i], want) {
			return nil, fmt.Errorf("xms_mirid is not a user-assigned managed identity resource id")
		}
	}

	if !azureGUID.MatchString(parts[1]) {
		return nil, fmt.Errorf("xms_mirid subscription %q is not a guid", parts[1])
	}
	if parts[3] == "" || parts[7] == "" {
		return nil, fmt.Errorf("xms_mirid is missing a resource group or identity name")
	}

	return &AzureManagedIdentity{
		SubscriptionID: parts[1],
		ResourceGroup:  parts[3],
		Name:           parts[7],
		PrincipalID:    principalID,
		TenantID:       tenantID,
	}, nil
}

// why: Rendered into the ARM template and compared against the verified xms_mirid, so both
// sides must stay in step.
func AzurePhoneHomeIdentityName(installID string) string {
	return installID + "-phone-home"
}
