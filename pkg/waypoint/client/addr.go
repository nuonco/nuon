package client

import "fmt"

func DefaultOrgServerDomain(rootDomain, orgID string) string {
	return fmt.Sprintf("%s.%s", orgID, rootDomain)
}

func DefaultOrgServerAddress(rootDomain, orgID string) string {
	return fmt.Sprintf("%s.%s:9701", orgID, rootDomain)
}
