package validate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
)

var stackDisallowedRefTypes = []refs.RefType{
	refs.RefTypeSandbox,
	refs.RefTypeComponents,
	refs.RefTypeActions,
	refs.RefTypeInstallStack,
}

var sandboxDisallowedRefTypes = []refs.RefType{
	refs.RefTypeComponents,
	refs.RefTypeActions,
}

type templateRefContext string

const (
	templateRefContextSandbox   templateRefContext = "sandbox"
	templateRefContextComponent templateRefContext = "component"
	templateRefContextAction    templateRefContext = "action"
	templateRefContextRunbook   templateRefContext = "runbook"
	templateRefContextOther     templateRefContext = "other"
)

type templateRefSite struct {
	entity  string
	context templateRefContext
	refs    []refs.Ref
}

type stackTemplateField struct {
	field string
	value string
}

// ValidateTemplateRefs checks that component and action template references
// resolve to declared names, and that stack/sandbox templates only use refs
// available in those render contexts. All findings are returned as one error.
func ValidateTemplateRefs(a *config.AppConfig) error {
	if a == nil {
		return nil
	}

	sites, stackFields := collectTemplateRefSites(a)
	var findings []string
	findings = append(findings, templateRefExistenceFindings(a, sites)...)
	findings = append(findings, templateScopeFindings(sites, stackFields)...)
	if len(findings) == 0 {
		return nil
	}

	sort.Strings(findings)
	return config.ErrConfig{
		Description: "invalid template references:\n" + strings.Join(findings, "\n"),
	}
}

func templateRefExistenceFindings(a *config.AppConfig, sites []templateRefSite) []string {
	componentNames, actionNames := declaredNames(a)
	var findings []string
	seen := make(map[string]bool)
	for _, site := range sites {
		for _, ref := range site.refs {
			key := string(ref.Type) + "|" + ref.Name + "|" + ref.Input
			if seen[key] {
				continue
			}
			switch ref.Type {
			case refs.RefTypeComponents:
				if !componentNames[ref.Name] {
					seen[key] = true
					findings = append(findings, fmt.Sprintf("%s references unknown component %q (%s)", site.entity, ref.Name, ref.Input))
				}
			case refs.RefTypeActions:
				if !actionNames[ref.Name] {
					seen[key] = true
					findings = append(findings, fmt.Sprintf("%s references unknown action %q (%s)", site.entity, ref.Name, ref.Input))
				}
			}
		}
	}
	return findings
}

func templateScopeFindings(sites []templateRefSite, stackFields []stackTemplateField) []string {
	var findings []string
	for _, site := range sites {
		if site.context != templateRefContextSandbox {
			continue
		}
		for _, ref := range site.refs {
			if generics.SliceContains(ref.Type, sandboxDisallowedRefTypes) {
				findings = append(findings, fmt.Sprintf("sandbox references %s, which is not available when rendering sandbox config (%s)", ref.Input, site.entity))
			}
		}
	}

	for _, f := range stackFields {
		if !strings.Contains(f.value, "{{") {
			continue
		}
		if strings.HasPrefix(f.field, "custom_nested_stacks") && strings.Contains(f.field, ".parameters.") {
			if err := config.ValidateStackParameterTemplate(f.value); err != nil {
				findings = append(findings, fmt.Sprintf("%s: %v", f.field, err))
			}
			continue
		}
		for _, ref := range refs.ParseFieldRefs(f.value) {
			if generics.SliceContains(ref.Type, stackDisallowedRefTypes) {
				findings = append(findings, fmt.Sprintf("%s references %s, which is not populated when the install stack is generated", f.field, ref.Input))
			}
		}
	}
	return findings
}

func declaredNames(a *config.AppConfig) (map[string]bool, map[string]bool) {
	components := make(map[string]bool)
	actions := make(map[string]bool)
	for _, c := range a.Components {
		if c != nil {
			components[c.Name] = true
		}
	}
	for _, act := range a.Actions {
		if act != nil {
			actions[act.Name] = true
		}
	}
	return components, actions
}

func collectTemplateRefSites(a *config.AppConfig) ([]templateRefSite, []stackTemplateField) {
	var sites []templateRefSite
	add := func(entity string, ctx templateRefContext, rs []refs.Ref) {
		if len(rs) == 0 {
			return
		}
		sites = append(sites, templateRefSite{entity: entity, context: ctx, refs: rs})
	}

	for _, c := range a.Components {
		if c == nil {
			continue
		}
		add("component:"+c.Name, templateRefContextComponent, c.References)
	}

	for _, act := range a.Actions {
		if act == nil {
			continue
		}
		add("action:"+act.Name, templateRefContextAction, act.References)
		for _, tr := range act.Triggers {
			if tr != nil && tr.ComponentName != "" {
				add("action:"+act.Name+":trigger:"+tr.Type, templateRefContextAction, []refs.Ref{{
					Type:  refs.RefTypeComponents,
					Name:  tr.ComponentName,
					Input: tr.ComponentName,
				}})
			}
		}
	}

	if a.Sandbox != nil {
		add("sandbox", templateRefContextSandbox, a.Sandbox.References)
	}

	for _, rb := range a.Runbooks {
		if rb == nil {
			continue
		}
		add("runbook:"+rb.Name, templateRefContextRunbook, rb.References)
	}

	if a.Permissions != nil {
		for _, role := range a.Permissions.Roles {
			if role == nil {
				continue
			}
			rs, err := refs.Parse(role)
			if err == nil {
				add("permission:"+role.Name, templateRefContextOther, rs)
			}
		}
	}

	if a.BreakGlass != nil {
		for _, role := range a.BreakGlass.Roles {
			if role == nil {
				continue
			}
			rs, err := refs.Parse(role)
			if err == nil {
				add("break_glass:"+role.Name, templateRefContextOther, rs)
			}
		}
	}

	if allRefs, err := refs.Parse(a); err == nil {
		known := make(map[string]bool)
		for _, s := range sites {
			for _, r := range s.refs {
				known[string(r.Type)+"|"+r.Name+"|"+r.Input] = true
			}
		}
		var extra []refs.Ref
		for _, r := range allRefs {
			if !known[string(r.Type)+"|"+r.Name+"|"+r.Input] {
				extra = append(extra, r)
			}
		}
		add("app", templateRefContextOther, extra)
	}

	var stackFields []stackTemplateField
	if a.Stack != nil {
		stackFields = collectStackTemplateFields(a.Stack)
	}

	return sites, stackFields
}

func collectStackTemplateFields(stack *config.StackConfig) []stackTemplateField {
	var fields []stackTemplateField
	add := func(field, value string) {
		if value == "" {
			return
		}
		fields = append(fields, stackTemplateField{field: field, value: value})
	}

	add("name", stack.Name)
	add("description", stack.Description)
	add("vpc_nested_template_url", stack.VPCNestedTemplateURL)
	add("runner_nested_template_url", stack.RunnerNestedTemplateURL)

	for i, ns := range stack.CustomNestedStacks {
		add(fmt.Sprintf("custom_nested_stacks[%d].template_url", i), ns.TemplateURL)
		for param, val := range ns.Parameters {
			add(fmt.Sprintf("custom_nested_stacks[%d].parameters.%s", i, param), val)
		}
	}

	return fields
}
