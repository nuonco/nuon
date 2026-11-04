package garimpersonation

import (
	"fmt"
	"strings"
)

type Allowlist map[string]map[string]struct{}

func Parse(entries []string) (Allowlist, error) {
	list := Allowlist{}
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		orgID, principal, ok := strings.Cut(entry, ":")
		orgID, principal = strings.TrimSpace(orgID), strings.TrimSpace(principal)
		if !ok || orgID == "" || principal == "" {
			return nil, fmt.Errorf("invalid gcp_gar_impersonation_allowlist entry %q, expected <org_id>:<principal>", entry)
		}
		if list[orgID] == nil {
			list[orgID] = map[string]struct{}{}
		}
		list[orgID][principal] = struct{}{}
	}
	return list, nil
}

func (a Allowlist) Validate(orgID, serviceAccountEmail, workloadIdentityProvider string) error {
	if serviceAccountEmail == "" && workloadIdentityProvider == "" {
		return nil
	}
	if orgID == "" {
		return fmt.Errorf("gcp_gar impersonation requires an org")
	}

	allowed := a[orgID]
	for _, p := range []struct{ field, value string }{
		{"service_account_email", serviceAccountEmail},
		{"workload_identity_provider", workloadIdentityProvider},
	} {
		if p.value == "" {
			continue
		}
		if _, ok := allowed[p.value]; !ok {
			return fmt.Errorf(
				"gcp_gar %s %q is not allowlisted for org %s on this control plane; "+
					"the control plane operator must add it to gcp_gar_impersonation_allowlist",
				p.field, p.value, orgID,
			)
		}
	}

	return nil
}
