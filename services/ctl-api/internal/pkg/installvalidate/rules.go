package installvalidate

import (
	"fmt"
	"sort"
)

const CodeDisabledDependency = "disabled_dependency"

func disabledDependencyDiagnostic(c *Context, dependentID, dependencyID string) Diagnostic {
	dependent := c.name(dependentID)
	dependency := c.name(dependencyID)
	return Diagnostic{
		Severity:   SeverityError,
		Code:       CodeDisabledDependency,
		Summary:    fmt.Sprintf("component %q is enabled but its dependency %q is disabled", dependent, dependency),
		Detail:     fmt.Sprintf("Enable %q, or disable %q, so the install reaches a consistent state.", dependency, dependent),
		Components: []string{dependent, dependency},
	}
}

type EnableRequiresDependenciesEnabledRule struct{}

func (EnableRequiresDependenciesEnabledRule) Name() string {
	return "enable_requires_dependencies_enabled"
}

func (EnableRequiresDependenciesEnabledRule) Check(c *Context) Diagnostics {
	if c.Op.Kind == OperationDisable {
		return nil
	}

	var ds Diagnostics
	edges := c.Resolver.DepEdges()
	for _, compID := range sortedKeys(edges) {
		if c.Op.Kind == OperationEnable && compID != c.Op.ComponentID {
			continue
		}
		if !c.Resolver.OwnEnabled(compID) {
			continue
		}
		deps := make([]string, 0, len(edges[compID]))
		for dep := range edges[compID] {
			deps = append(deps, dep)
		}
		sort.Strings(deps)
		for _, dep := range deps {
			if !c.Resolver.OwnEnabled(dep) {
				ds = append(ds, disabledDependencyDiagnostic(c, compID, dep))
			}
		}
	}
	return ds
}

type DisableRequiresDependentsDisabledRule struct{}

func (DisableRequiresDependentsDisabledRule) Name() string {
	return "disable_requires_dependents_disabled"
}

func (DisableRequiresDependentsDisabledRule) Check(c *Context) Diagnostics {
	if c.Op.Kind == OperationEnable {
		return nil
	}

	var ds Diagnostics
	for _, compID := range sortedComponentIDs(c) {
		if c.Op.Kind == OperationDisable && compID != c.Op.ComponentID {
			continue
		}
		if c.Resolver.OwnEnabled(compID) {
			continue
		}
		for _, dependent := range c.Resolver.DirectDependents(compID) {
			if c.Resolver.OwnEnabled(dependent) {
				ds = append(ds, disabledDependencyDiagnostic(c, dependent, compID))
			}
		}
	}
	return ds
}

func sortedKeys(m map[string]map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedComponentIDs(c *Context) []string {
	ids := make([]string, 0, len(c.CCCByID))
	for id := range c.CCCByID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
