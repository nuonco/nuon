// Package teardown decides which terraform resources a component teardown
// destroys and which it removes from state, leaving them in place.
package teardown

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
)

// Partition splits managed resource addresses into the ones selected by a
// retain entry and the ones to destroy. Entries that select nothing are
// returned as unmatched so callers can surface them.
func Partition(stateAddrs, retain []string) (retained, destroy, unmatched []string, err error) {
	entries := make([]address, 0, len(retain))
	for _, r := range retain {
		parsed, err := parseAddress(r)
		if err != nil {
			return nil, nil, nil, err
		}
		entries = append(entries, parsed)
	}

	used := make([]bool, len(entries))
	for _, addr := range stateAddrs {
		parsed, err := parseAddress(addr)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("unable to parse state address: %w", err)
		}

		isRetained := false
		for idx, entry := range entries {
			if entry.matches(parsed) {
				used[idx] = true
				isRetained = true
			}
		}

		if isRetained {
			retained = append(retained, addr)
		} else {
			destroy = append(destroy, addr)
		}
	}

	for idx, ok := range used {
		if !ok {
			unmatched = append(unmatched, retain[idx])
		}
	}

	return retained, destroy, unmatched, nil
}

// ManagedAddresses returns the address of every managed resource instance in
// the state, including those in child modules.
func ManagedAddresses(state *tfjson.State) []string {
	resources := managedResources(state)

	addrs := make([]string, 0, len(resources))
	for addr := range resources {
		addrs = append(addrs, addr)
	}
	sort.Strings(addrs)
	return addrs
}

type stateResource struct {
	resource      *tfjson.StateResource
	moduleAddress string
}

func managedResources(state *tfjson.State) map[string]stateResource {
	resources := make(map[string]stateResource)
	if state == nil || state.Values == nil {
		return resources
	}

	var walk func(mod *tfjson.StateModule)
	walk = func(mod *tfjson.StateModule) {
		if mod == nil {
			return
		}
		for _, res := range mod.Resources {
			if res != nil && res.Mode == tfjson.ManagedResourceMode {
				resources[res.Address] = stateResource{resource: res, moduleAddress: mod.Address}
			}
		}
		for _, child := range mod.ChildModules {
			walk(child)
		}
	}
	walk(state.Values.RootModule)

	return resources
}

// CheckPlan returns an error when the plan deletes a resource selected by a
// retain entry. Terraform destroys everything that depends on a targeted
// resource, so a retained resource with a destroyed dependency ends up in the
// plan.
func CheckPlan(plan *tfjson.Plan, retain []string) error {
	if plan == nil || len(retain) == 0 {
		return nil
	}

	var deleting []string
	for _, rc := range plan.ResourceChanges {
		if rc == nil || rc.Change == nil || rc.Mode != tfjson.ManagedResourceMode {
			continue
		}
		if !rc.Change.Actions.Delete() && !rc.Change.Actions.DestroyBeforeCreate() && !rc.Change.Actions.CreateBeforeDestroy() {
			continue
		}
		deleting = append(deleting, rc.Address)
	}

	retained, _, _, err := Partition(deleting, retain)
	if err != nil {
		return err
	}
	if len(retained) == 0 {
		return nil
	}

	return fmt.Errorf("the teardown plan would destroy retained resources: %s. "+
		"A retained resource is destroyed when it depends on a resource that is not retained; "+
		"retain its dependencies too, or move it into its own component",
		strings.Join(retained, ", "))
}

// AddForgetChanges records each retained address in the plan as a "forget"
// change, the action terraform uses for resources removed from state without
// being destroyed. The resource's current attributes are carried as the
// change's before value so reviewers can see what is being left in place.
func AddForgetChanges(plan *tfjson.Plan, state *tfjson.State, retained []string) {
	if plan == nil {
		return
	}

	resources := managedResources(state)
	for _, addr := range retained {
		rc := &tfjson.ResourceChange{
			Address: addr,
			Mode:    tfjson.ManagedResourceMode,
			Change: &tfjson.Change{
				Actions: tfjson.Actions{tfjson.ActionForget},
			},
		}
		if sr, ok := resources[addr]; ok {
			rc.ModuleAddress = sr.moduleAddress
			rc.Type = sr.resource.Type
			rc.Name = sr.resource.Name
			rc.Index = sr.resource.Index
			rc.ProviderName = sr.resource.ProviderName
			rc.Change.Before = sr.resource.AttributeValues
			rc.Change.BeforeSensitive = sensitiveValues(sr.resource.SensitiveValues)
		}
		plan.ResourceChanges = append(plan.ResourceChanges, rc)
	}
}

func sensitiveValues(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}
