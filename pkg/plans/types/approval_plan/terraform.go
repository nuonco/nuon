package approvalplan

import (
	"encoding/json"
	"reflect"

	tfjson "github.com/hashicorp/terraform-json"
)

type TerraformApprovalPlan struct {
	planJSON []byte
}

func NewTerraformApprovalPlan(planJSON []byte) *TerraformApprovalPlan {
	return &TerraformApprovalPlan{
		planJSON: planJSON,
	}
}

// NoopPlan imeplementation for terraform is based on https://github.com/hashicorp/terraform/blob/eee744c8874f15c131651d8b34bd4860fdebcaed/internal/command/jsonformat/plan.go#L58
func terraformPlanNoop(planJSON []byte) (bool, error) {
	if len(planJSON) == 0 {
		return false, nil
	}

	var plan tfjson.Plan
	if err := json.Unmarshal(planJSON, &plan); err != nil {
		return false, err
	}

	// Check for resource changes, output changes, and action invocations.
	noResourceChanges := len(plan.ResourceChanges) == 0
	noOutputChanges := len(plan.OutputChanges) == 0

	// Deferred changes are not considered current changes. If there are deferred changes,
	// infrastructure is not up to date.
	noDeferredChanges := len(plan.DeferredChanges) == 0

	// If all are empty, infra is up to date.
	if noResourceChanges && noOutputChanges && noDeferredChanges {
		return true, nil
	}

	// Check each resource change for no-op actions
	for _, rc := range plan.ResourceChanges {
		if rc.Change != nil && len(rc.Change.Actions) == 1 && rc.Change.Actions[0] != tfjson.ActionNoop {
			return false, nil
		}
	}

	// Check each output change for no-op actions
	for _, oc := range plan.OutputChanges {
		if oc != nil && len(oc.Actions) == 1 && oc.Actions[0] != tfjson.ActionNoop {
			return false, nil
		}
	}

	relevant := relevantAttributesByResource(plan.RelevantAttributes)
	for _, rd := range plan.ResourceDrift {
		if rd != nil && rd.Change != nil && isRelevantDrift(rd, relevant[rd.Address]) {
			return false, nil
		}
	}

	return true, nil
}

type relevantAttributes struct {
	wholeResource bool
	paths         [][]interface{}
}

func relevantAttributesByResource(attrs []tfjson.ResourceAttribute) map[string]*relevantAttributes {
	byResource := make(map[string]*relevantAttributes)
	for _, attr := range attrs {
		ra, ok := byResource[attr.Resource]
		if !ok {
			ra = &relevantAttributes{}
			byResource[attr.Resource] = ra
		}

		path, err := decodeAttributePath(attr.Attribute)
		if err != nil || len(path) == 0 {
			ra.wholeResource = true
			continue
		}
		ra.paths = append(ra.paths, path)
	}
	return byResource
}

func decodeAttributePath(raw []json.RawMessage) ([]interface{}, error) {
	path := make([]interface{}, 0, len(raw))
	for _, step := range raw {
		var v interface{}
		if err := json.Unmarshal(step, &v); err != nil {
			return nil, err
		}
		path = append(path, v)
	}
	return path, nil
}

func valueAtPath(v interface{}, path []interface{}) (interface{}, bool) {
	for _, step := range path {
		switch key := step.(type) {
		case string:
			m, ok := v.(map[string]interface{})
			if !ok {
				return nil, false
			}
			if v, ok = m[key]; !ok {
				return nil, false
			}
		case float64:
			l, ok := v.([]interface{})
			idx := int(key)
			if !ok || idx < 0 || idx >= len(l) {
				return nil, false
			}
			v = l[idx]
		default:
			return nil, false
		}
	}
	return v, true
}

func isRelevantDrift(rd *tfjson.ResourceChange, relevant *relevantAttributes) bool {
	actions := rd.Change.Actions
	if len(actions) == 0 || (len(actions) == 1 && actions[0] == tfjson.ActionNoop) {
		return false
	}
	if len(actions) > 1 || actions[0] != tfjson.ActionUpdate {
		return true
	}
	if relevant == nil {
		return false
	}
	if relevant.wholeResource {
		return true
	}

	for _, path := range relevant.paths {
		before, beforeOK := valueAtPath(rd.Change.Before, path)
		after, afterOK := valueAtPath(rd.Change.After, path)
		if beforeOK != afterOK || !reflect.DeepEqual(before, after) {
			return true
		}
	}
	return false
}

func (t *TerraformApprovalPlan) IsNoop() (bool, error) {
	return terraformPlanNoop(t.planJSON)
}
