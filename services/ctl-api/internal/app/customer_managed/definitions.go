package customermanaged

import (
	"encoding/json"
	"sort"

	"github.com/opencontainers/go-digest"

	bundle "github.com/nuonco/nuon/pkg/customer_managed/bundle"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func ObjectDigest(value any) string {
	raw, _ := json.Marshal(value)
	return digest.FromBytes(raw).String()
}

func CanonicalObject(value any) (map[string]any, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	stripPersistenceFields(result)
	return result, nil
}

func CanonicalComponentDefinition(connection app.ComponentConfigConnection, connections []app.ComponentConfigConnection) (bundle.ComponentDefinition, error) {
	raw, err := json.Marshal(connection)
	if err != nil {
		return nil, err
	}
	var definition bundle.ComponentDefinition
	if err := json.Unmarshal(raw, &definition); err != nil {
		return nil, err
	}
	stripPersistenceFields(map[string]any(definition))
	for _, key := range []string{"Refs", "app_config_version", "checksum", "component_dependency_ids", "component_id", "component_name", "latest_build_id", "version"} {
		delete(definition, key)
	}
	componentNames := make(map[string]string, len(connections))
	for _, candidate := range connections {
		name := candidate.ComponentName
		if name == "" {
			name = candidate.ComponentID
		}
		componentNames[candidate.ComponentID] = name
	}
	dependencies := make([]string, 0, len(connection.ComponentDependencyIDs))
	for _, componentID := range connection.ComponentDependencyIDs {
		name := componentNames[componentID]
		if name == "" {
			name = componentID
		}
		dependencies = append(dependencies, name)
	}
	if len(dependencies) > 0 {
		sort.Strings(dependencies)
		definition["dependencies"] = dependencies
	}
	return definition, nil
}

func CanonicalActionDefinition(actionConfig app.ActionWorkflowConfig, connections []app.ComponentConfigConnection) bundle.ActionDefinition {
	componentNames := make(map[string]string, len(connections))
	for _, connection := range connections {
		componentNames[connection.ComponentID] = connection.ComponentName
	}
	definition := bundle.ActionDefinition{
		TimeoutNanos:          actionConfig.Timeout.Nanoseconds(),
		Role:                  actionConfig.Role,
		BreakGlassRoleARN:     actionConfig.BreakGlassRoleARN.ValueString(),
		EnableKubeConfig:      actionConfig.EnableKubeConfig.Valid && actionConfig.EnableKubeConfig.Bool,
		KubernetesContextName: actionConfig.KubernetesContextName,
		References:            append([]string(nil), actionConfig.References...),
	}
	for _, componentID := range actionConfig.ComponentDependencyIDs {
		name := componentNames[componentID]
		if name == "" {
			name = componentID
		}
		definition.ComponentDependencies = append(definition.ComponentDependencies, name)
	}
	for _, trigger := range actionConfig.Triggers {
		componentID := trigger.ComponentID.ValueString()
		componentName := componentNames[componentID]
		if componentName == "" {
			componentName = componentID
		}
		definition.Triggers = append(definition.Triggers, bundle.ActionTriggerDefinition{
			Type: string(trigger.Type), Index: trigger.Index, CronSchedule: trigger.CronSchedule, ComponentName: componentName,
		})
	}
	for _, step := range actionConfig.Steps {
		environment := make(map[string]string, len(step.EnvVars))
		for name, value := range step.EnvVars {
			if value == nil {
				environment[name] = ""
				continue
			}
			environment[name] = digest.FromString(*value).String()
		}
		inlineDigest := ""
		if step.InlineContents != "" {
			inlineDigest = digest.FromString(step.InlineContents).String()
		}
		definition.Steps = append(definition.Steps, bundle.ActionStepDefinition{
			Name: step.Name, Index: step.Idx, Command: step.Command, InlineContentsDigest: inlineDigest, Environment: environment,
		})
	}
	sort.Strings(definition.ComponentDependencies)
	sort.Strings(definition.References)
	sort.Slice(definition.Triggers, func(i, j int) bool {
		if definition.Triggers[i].Index == definition.Triggers[j].Index {
			return definition.Triggers[i].Type < definition.Triggers[j].Type
		}
		return definition.Triggers[i].Index < definition.Triggers[j].Index
	})
	sort.Slice(definition.Steps, func(i, j int) bool {
		if definition.Steps[i].Index == definition.Steps[j].Index {
			return definition.Steps[i].Name < definition.Steps[j].Name
		}
		return definition.Steps[i].Index < definition.Steps[j].Index
	})
	return definition
}

func CanonicalRunbookDefinition(runbookConfig app.RunbookConfig, actionConfigs []app.ActionWorkflowConfig) bundle.RunbookDefinition {
	actionNames := make(map[string]string, len(actionConfigs))
	for _, action := range actionConfigs {
		actionNames[action.ActionWorkflowID] = action.ActionWorkflow.Name
	}
	definition := bundle.RunbookDefinition{}
	if runbookConfig.Readme != "" {
		definition.ReadmeDigest = digest.FromString(runbookConfig.Readme).String()
	}
	for _, input := range runbookConfig.Inputs {
		defaultValue := input.Default
		if input.Sensitive && defaultValue != "" {
			defaultValue = digest.FromString(defaultValue).String()
		}
		definition.Inputs = append(definition.Inputs, bundle.RunbookInputDefinition{
			Name: input.Name, DisplayName: input.DisplayName, Description: input.Description,
			Default: defaultValue, Type: string(input.Type), Index: input.Idx, Required: input.Required, Sensitive: input.Sensitive,
		})
	}
	for _, step := range runbookConfig.Steps {
		reference := ""
		if actionID := step.ActionWorkflowID.ValueString(); actionID != "" {
			reference = actionNames[actionID]
			if reference == "" {
				reference = actionID
			}
		}
		environment := make(map[string]string, len(step.EnvVars))
		for key, value := range step.EnvVars {
			if value == nil {
				environment[key] = ""
			} else {
				environment[key] = digest.FromString(*value).String()
			}
		}
		inlineDigest := ""
		if step.InlineContents != "" {
			inlineDigest = digest.FromString(step.InlineContents).String()
		}
		filtersDigest := ""
		if len(step.Filters) > 0 {
			filtersDigest = ObjectDigest(step.Filters)
		}
		eventTypes := append([]string(nil), step.EventTypes...)
		sort.Strings(eventTypes)
		definition.Steps = append(definition.Steps, bundle.RunbookStepDefinition{
			Kind: string(step.Type), Name: step.Name, Index: step.Idx, Reference: reference, Component: step.ComponentName,
			Role: step.Role, PlanOnly: step.PlanOnly, DeployDependents: step.DeployDependents,
			TearDownDependents: step.TearDownDependents, SkipComponentDeploys: step.SkipComponentDeploys,
			Command: step.Command, InlineContentsDigest: inlineDigest, Environment: environment,
			TimeoutNanos: step.Timeout.Nanoseconds(), TriggerName: step.TriggerName, EventTypes: eventTypes, FiltersDigest: filtersDigest,
		})
	}
	sort.Slice(definition.Inputs, func(i, j int) bool {
		if definition.Inputs[i].Index == definition.Inputs[j].Index {
			return definition.Inputs[i].Name < definition.Inputs[j].Name
		}
		return definition.Inputs[i].Index < definition.Inputs[j].Index
	})
	sort.Slice(definition.Steps, func(i, j int) bool {
		if definition.Steps[i].Index == definition.Steps[j].Index {
			return definition.Steps[i].Name < definition.Steps[j].Name
		}
		return definition.Steps[i].Index < definition.Steps[j].Index
	})
	return definition
}

func stripPersistenceFields(value any) {
	persistenceFields := map[string]bool{
		"app_config_id": true, "component_config_connection_id": true, "component_config_id": true,
		"component_config_type": true, "created_at": true, "created_by_id": true, "deleted_at": true,
		"id": true, "org_id": true, "updated_at": true, "vcs_connection_id": true,
	}
	switch value := value.(type) {
	case map[string]any:
		for key, child := range value {
			if persistenceFields[key] {
				delete(value, key)
				continue
			}
			stripPersistenceFields(child)
		}
	case []any:
		for _, child := range value {
			stripPersistenceFields(child)
		}
	}
}
