package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
)

const maxStackPropertyChanges = 200

type stackTemplate struct {
	Resources map[string]struct {
		Type       string         `json:"Type"`
		Properties map[string]any `json:"Properties"`
	} `json:"Resources"`
}

func addStackPropertyChanges(candidate *day2.StackCandidate, beforeRaw, afterRaw []byte) error {
	var before, after stackTemplate
	if err := json.Unmarshal(beforeRaw, &before); err != nil {
		return fmt.Errorf("decode deployed stack template: %w", err)
	}
	if err := json.Unmarshal(afterRaw, &after); err != nil {
		return fmt.Errorf("decode candidate stack template: %w", err)
	}
	for i := range candidate.Changes {
		candidate.Changes[i].PropertyChangesCaptured = true
		oldResource, hadOld := before.Resources[candidate.Changes[i].LogicalResourceID]
		newResource, hasNew := after.Resources[candidate.Changes[i].LogicalResourceID]
		var values []day2.StackPropertyChange
		diffStackValue("Properties", oldResource.Properties, hadOld, newResource.Properties, hasNew, &values)
		if len(values) > maxStackPropertyChanges {
			candidate.Changes[i].PropertyChanges = values[:maxStackPropertyChanges]
			candidate.Changes[i].PropertyChangesTruncated = true
			continue
		}
		candidate.Changes[i].PropertyChanges = values
	}
	return nil
}

func stackCandidateTemplateKey(bundleDigest string) string {
	return "stack/candidates/" + strings.ReplaceAll(bundleDigest, ":", "-") + "/stack/root-template.json"
}

func stackPropertyChangesCaptured(candidate day2.StackCandidate) bool {
	for _, change := range candidate.Changes {
		if change.PropertyChangesCaptured || len(change.PropertyChanges) > 0 || change.PropertyChangesTruncated {
			return true
		}
	}
	return false
}

func diffStackValue(path string, before any, beforeSet bool, after any, afterSet bool, changes *[]day2.StackPropertyChange) {
	if beforeSet && afterSet && reflect.DeepEqual(before, after) {
		return
	}
	beforeKeyed, beforeIsKeyed := keyedStackList(before)
	afterKeyed, afterIsKeyed := keyedStackList(after)
	if beforeIsKeyed && afterIsKeyed {
		diffStackMap(path, beforeKeyed, afterKeyed, changes, func(parent, key string) string {
			return parent + "[" + key + "]"
		})
		return
	}
	beforeMap, beforeIsMap := before.(map[string]any)
	afterMap, afterIsMap := after.(map[string]any)
	if beforeIsMap && afterIsMap {
		diffStackMap(path, beforeMap, afterMap, changes, func(parent, key string) string { return parent + "." + key })
		return
	}
	change := day2.StackPropertyChange{Path: path}
	if beforeSet {
		change.Before = before
	}
	if afterSet {
		change.After = after
	}
	*changes = append(*changes, change)
}

func diffStackMap(path string, before, after map[string]any, changes *[]day2.StackPropertyChange, childPath func(string, string) string) {
	keys := make([]string, 0, len(before)+len(after))
	seen := make(map[string]bool, len(before)+len(after))
	for key := range before {
		keys = append(keys, key)
		seen[key] = true
	}
	for key := range after {
		if !seen[key] {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		oldValue, hadOld := before[key]
		newValue, hasNew := after[key]
		diffStackValue(childPath(path, key), oldValue, hadOld, newValue, hasNew, changes)
	}
}

func keyedStackList(value any) (map[string]any, bool) {
	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		return nil, false
	}
	result := make(map[string]any, len(items))
	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			return nil, false
		}
		key, ok := object["Key"].(string)
		if !ok || key == "" || result[key] != nil {
			return nil, false
		}
		result[key] = object
	}
	return result, true
}
