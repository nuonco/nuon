package installcreate

import (
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func GroupLabels(group *models.AppAppBranchInstallGroup) (map[string]string, error) {
	if group == nil || group.ID == "" || group.LabelSelector == nil || len(group.LabelSelector.MatchLabels) == 0 {
		return nil, fmt.Errorf("install group must use a label selector with match labels")
	}

	result := make(map[string]string, len(group.LabelSelector.MatchLabels))
	for key, value := range group.LabelSelector.MatchLabels {
		if value == "*" {
			return nil, fmt.Errorf("install group label %q uses a wildcard and cannot be applied during creation", key)
		}
		result[key] = value
	}
	return result, nil
}

func MergeGroupLabels(into, groupLabels map[string]string) error {
	for key, value := range groupLabels {
		if existing, ok := into[key]; ok && existing != value {
			return fmt.Errorf("label %q is %q, but the selected install group requires %q", key, existing, value)
		}
		into[key] = value
	}
	return nil
}

func EligibleGroups(groups []*models.AppAppBranchInstallGroup) []*models.AppAppBranchInstallGroup {
	eligible := make([]*models.AppAppBranchInstallGroup, 0, len(groups))
	for _, group := range groups {
		if _, err := GroupLabels(group); err == nil {
			eligible = append(eligible, group)
		}
	}
	return eligible
}
