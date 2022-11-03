package waypoint

import "fmt"

func DefaultOrgServerAddress(rootDomain, orgID string) string {
	return fmt.Sprintf("%s.%s:9701", orgID, rootDomain)
}
