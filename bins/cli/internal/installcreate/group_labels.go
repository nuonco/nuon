package installcreate

import (
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// GroupLabels returns the labels an install must carry to join a group. Only
// label-based groups with fully concrete match labels can be selected at
// creation time, since a wildcard has no single value to apply.
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

// MergeGroupLabels applies a group's labels onto an install's labels, refusing to
// overwrite a value the user set explicitly.
func MergeGroupLabels(into, groupLabels map[string]string) error {
	for key, value := range groupLabels {
		if existing, ok := into[key]; ok && existing != value {
			return fmt.Errorf("label %q is %q, but the selected install group requires %q", key, existing, value)
		}
		into[key] = value
	}
	return nil
}

// EligibleGroups filters a branch config's install groups down to those that can
// be applied during install creation.
func EligibleGroups(groups []*models.AppAppBranchInstallGroup) []*models.AppAppBranchInstallGroup {
	eligible := make([]*models.AppAppBranchInstallGroup, 0, len(groups))
	for _, group := range groups {
		if _, err := GroupLabels(group); err == nil {
			eligible = append(eligible, group)
		}
	}
	return eligible
}
