package runner

import (
	"fmt"
)

func getOrgPrefix(orgShortID string, appShortID string) string {
	return fmt.Sprintf("installations/org=%s/app=%s", orgShortID, appShortID)
}
