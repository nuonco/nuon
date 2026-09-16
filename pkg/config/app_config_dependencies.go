package config

import (
	"strings"

	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/pkg/config/refs"
)

const (
	OperationRolesResourceID diff.NodeID = "operation_roles"
	RunnerResourceID         diff.NodeID = "runner"
	SandboxResourceID        diff.NodeID = "sandbox"
	StackResourceID          diff.NodeID = "stack"
)

func ComponentResourceID(name string) diff.NodeID {
	return diff.NodeID("component." + name)
}

func ActionResourceID(name string) diff.NodeID {
	return diff.NodeID("action." + name)
}

func InputResourceID(name string) diff.NodeID {
	return diff.NodeID("input." + name)
}

func SecretResourceID(name string) diff.NodeID {
	return diff.NodeID("secret." + name)
}

func RoleResourceID(name string) diff.NodeID {
	return diff.NodeID("role." + strings.ReplaceAll(strings.TrimSpace(name), " ", ""))
}

func roleNodeID(old, new *AppAWSIAMRole) diff.NodeID {
	if new != nil && new.Name != "" {
		return RoleResourceID(new.Name)
	}
	if old != nil {
		return RoleResourceID(old.Name)
	}
	return ""
}

func changedResourceIDs(root *diff.Diff) []diff.NodeID {
	var changed []diff.NodeID
	var walk func(*diff.Diff)
	walk = func(node *diff.Diff) {
		if node == nil {
			return
		}
		if node.ResourceID != "" && node.DirectSummary().HasChanged {
			changed = append(changed, node.ResourceID)
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	return changed
}

func appConfigDependencyGraph(configs ...*AppConfig) *diff.Graph {
	graph := diff.NewGraph()
	for _, cfg := range configs {
		addAppConfigDependencyEdges(graph, cfg)
	}
	return graph
}

func addAppConfigDependencyEdges(graph *diff.Graph, cfg *AppConfig) {
	if cfg == nil {
		return
	}

	graph.AddDependency(StackResourceID, RunnerResourceID, diff.EdgeReasonStackRender)
	for _, role := range appConfigRoles(cfg) {
		graph.AddDependency(StackResourceID, RoleResourceID(role.Name), diff.EdgeReasonStackRender)
	}
	if cfg.Secrets != nil {
		for _, secret := range cfg.Secrets.Secrets {
			if secret != nil {
				graph.AddDependency(StackResourceID, SecretResourceID(secret.Name), diff.EdgeReasonStackRender)
			}
		}
	}

	for _, component := range cfg.Components {
		if component == nil || component.Name == "" {
			continue
		}
		dependent := ComponentResourceID(component.Name)

		for _, dependency := range component.Dependencies {
			graph.AddDependency(dependent, ComponentResourceID(dependency), diff.EdgeReasonComponentDependency)
		}
		for _, ref := range componentRefs(component) {
			addReferenceDependency(graph, dependent, ref)
		}
		for _, role := range component.OperationRoles {
			graph.AddDependency(dependent, RoleResourceID(role.RoleName), diff.EdgeReasonOperationRole)
		}
		addDefaultComponentRoleEdges(graph, dependent, cfg.Permissions)
	}

	for _, action := range cfg.Actions {
		if action == nil || action.Name == "" {
			continue
		}
		dependent := ActionResourceID(action.Name)
		for _, dependency := range action.Dependencies {
			graph.AddDependency(dependent, ComponentResourceID(dependency), diff.EdgeReasonComponentDependency)
		}
		references, err := refs.Parse(action)
		if err == nil {
			allReferences := append([]refs.Ref(nil), action.References...)
			allReferences = append(allReferences, references...)
			for _, ref := range allReferences {
				addReferenceDependency(graph, dependent, ref)
			}
		}
		for _, roleName := range []string{action.Role, action.BreakGlassRole} {
			if roleName != "" {
				graph.AddDependency(dependent, RoleResourceID(roleName), diff.EdgeReasonOperationRole)
			}
		}
	}

	addOperationRoleEdges(graph, cfg)
}

func appConfigRoles(cfg *AppConfig) []*AppAWSIAMRole {
	var roles []*AppAWSIAMRole
	if cfg.Permissions != nil {
		roles = append(roles,
			cfg.Permissions.ProvisionRole,
			cfg.Permissions.DeprovisionRole,
			cfg.Permissions.MaintenanceRole,
		)
		roles = append(roles, cfg.Permissions.Roles...)
		roles = append(roles, cfg.Permissions.CustomRoles...)
	}
	if cfg.BreakGlass != nil {
		roles = append(roles, cfg.BreakGlass.Roles...)
	}

	filtered := roles[:0]
	for _, role := range roles {
		if role != nil && role.Name != "" {
			filtered = append(filtered, role)
		}
	}
	return filtered
}

func componentRefs(component *Component) []refs.Ref {
	references := append([]refs.Ref(nil), component.References...)
	parsed, err := refs.Parse(component)
	if err == nil {
		references = append(references, parsed...)
	}
	return references
}

func addReferenceDependency(graph *diff.Graph, dependent diff.NodeID, ref refs.Ref) {
	switch ref.Type {
	case refs.RefTypeComponents:
		graph.AddDependency(dependent, ComponentResourceID(ref.Name), diff.EdgeReasonComponentReference)
	case refs.RefTypeInputs, refs.RefTypeInstallInputs:
		graph.AddDependency(dependent, InputResourceID(ref.Name), diff.EdgeReasonInputReference)
	case refs.RefTypeInstallStack:
		graph.AddDependency(dependent, StackResourceID, diff.EdgeReasonInstallStackOutput)
	case refs.RefTypeSandbox:
		graph.AddDependency(dependent, SandboxResourceID, diff.EdgeReasonSandboxOutput)
	case refs.RefTypeSecrets:
		graph.AddDependency(dependent, SecretResourceID(ref.Name), diff.EdgeReasonSecretReference)
	}
}

func addDefaultComponentRoleEdges(graph *diff.Graph, dependent diff.NodeID, permissions *PermissionsConfig) {
	if permissions == nil {
		return
	}
	for _, role := range []*AppAWSIAMRole{
		permissions.ProvisionRole,
		permissions.DeprovisionRole,
		permissions.MaintenanceRole,
	} {
		if role != nil && role.Name != "" {
			graph.AddDependency(dependent, RoleResourceID(role.Name), diff.EdgeReasonOperationRole)
		}
	}
}

func addOperationRoleEdges(graph *diff.Graph, cfg *AppConfig) {
	if cfg.OperationRoles == nil {
		return
	}
	for _, rule := range cfg.OperationRoles.RuleMatrix {
		if rule == nil || rule.RoleName == "" {
			continue
		}
		principalType, principalName, err := rule.ParsePrincipal()
		if err != nil {
			continue
		}
		switch PrincipalType(principalType) {
		case PrincipalTypeComponent:
			if principalName == "*" {
				for _, component := range cfg.Components {
					if component != nil {
						graph.AddDependency(ComponentResourceID(component.Name), RoleResourceID(rule.RoleName), diff.EdgeReasonOperationRole)
					}
				}
			} else {
				graph.AddDependency(ComponentResourceID(principalName), RoleResourceID(rule.RoleName), diff.EdgeReasonOperationRole)
			}
		case PrincipalTypeSandbox:
			graph.AddDependency(SandboxResourceID, RoleResourceID(rule.RoleName), diff.EdgeReasonOperationRole)
		case PrincipalTypeAction:
			if principalName == "*" {
				for _, action := range cfg.Actions {
					if action != nil {
						graph.AddDependency(ActionResourceID(action.Name), RoleResourceID(rule.RoleName), diff.EdgeReasonOperationRole)
					}
				}
			} else {
				graph.AddDependency(ActionResourceID(principalName), RoleResourceID(rule.RoleName), diff.EdgeReasonOperationRole)
			}
		}
	}
}
