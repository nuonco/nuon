package airgap

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type Finding struct {
	Code    string `json:"code"`
	Member  string `json:"member"`
	Message string `json:"message"`
}

type QualificationReport struct {
	Qualified  bool      `json:"qualified"`
	Platform   string    `json:"platform"`
	Violations []Finding `json:"violations"`
	Warnings   []Finding `json:"warnings"`
}

func Qualify(cfg *app.AppConfig, platform string) QualificationReport {
	r := QualificationReport{Platform: platform, Violations: []Finding{}, Warnings: []Finding{}}
	add := func(code, member, message string) {
		r.Violations = append(r.Violations, Finding{Code: code, Member: member, Message: message})
	}
	warn := func(code, member, message string) {
		r.Warnings = append(r.Warnings, Finding{Code: code, Member: member, Message: message})
	}
	if platform != "linux/amd64" {
		add("platform.unsupported", platform, "immutable air-gap bundles support only linux/amd64")
	}
	if cfg == nil {
		add("app_config.missing", "app_config", "app config is required")
		finish(&r)
		return r
	}

	componentMembers := make(map[string]bool, len(cfg.ComponentIDs))
	for _, id := range cfg.ComponentIDs {
		if id == "" {
			add("component.member_invalid", "component", "AppConfig contains an empty component member ID")
			continue
		}
		if componentMembers[id] {
			add("component.member_duplicate", "component:"+id, "AppConfig contains a duplicate component member ID")
		}
		componentMembers[id] = true
	}
	connections := make(map[string]int, len(cfg.ComponentConfigConnections))
	for i := range cfg.ComponentConfigConnections {
		c := &cfg.ComponentConfigConnections[i]
		member := componentMember(c, i)
		connections[c.ComponentID]++
		if c.ID == "" {
			add("component.config_missing", member, "component config connection ID is missing")
		}
		if c.ComponentID == "" || !componentMembers[c.ComponentID] {
			add("component.membership_mismatch", member, "component connection is not an exact AppConfig component member")
		}
		qualifyComponent(c, member, add, warn)
	}
	for _, id := range cfg.ComponentIDs {
		if id == "" {
			continue
		}
		if connections[id] == 0 {
			add("component.config_missing", "component:"+id, "AppConfig component member has no config connection")
		} else if connections[id] > 1 {
			add("component.config_duplicate", "component:"+id, "AppConfig component member has multiple config connections")
		}
	}

	if cfg.SandboxConfig.ID == "" {
		add("sandbox.config_missing", "sandbox", "AppConfig sandbox config is missing")
	}
	if cfg.SandboxConfig.AppConfigID != cfg.ID {
		add("sandbox.app_config_mismatch", "sandbox", fmt.Sprintf("sandbox app config %q does not match %q", cfg.SandboxConfig.AppConfigID, cfg.ID))
	}
	if cfg.SandboxConfig.Type == "pulumi" {
		add("sandbox.pulumi_unsupported", "sandbox", "Pulumi sandboxes are not supported")
	} else if cfg.SandboxConfig.Type != "terraform" {
		add("sandbox.type_unsupported", "sandbox", fmt.Sprintf("sandbox type %q is unsupported", cfg.SandboxConfig.Type))
	}

	actionMembers := make(map[string]bool, len(cfg.ActionIDs))
	for _, id := range cfg.ActionIDs {
		if id == "" {
			add("action.member_invalid", "action", "AppConfig contains an empty action member ID")
			continue
		}
		if actionMembers[id] {
			add("action.member_duplicate", "action:"+id, "AppConfig contains a duplicate action member ID")
		}
		actionMembers[id] = true
	}
	actions := make(map[string]bool, len(cfg.ActionWorkflowConfigs))
	for i := range cfg.ActionWorkflowConfigs {
		a := &cfg.ActionWorkflowConfigs[i]
		member := actionMember(a, i)
		actions[a.ActionWorkflowID] = true
		if a.ID == "" {
			add("action.config_missing", member, "action workflow config ID is missing")
		}
		if a.AppConfigID != cfg.ID {
			add("action.app_config_mismatch", member, fmt.Sprintf("action config app config %q does not match %q", a.AppConfigID, cfg.ID))
		}
		if a.ActionWorkflowID == "" || !actionMembers[a.ActionWorkflowID] {
			add("action.membership_mismatch", member, "action config is not an exact AppConfig action member")
		}
		hasCron := a.HasTrigger(app.ActionWorkflowTriggerTypeCron)
		for j := range a.Steps {
			s := &a.Steps[j]
			step := fmt.Sprintf("%s/step:%s", member, logical(s.Name, s.ID, j))
			if s.AppConfigID != cfg.ID {
				add("action_step.app_config_mismatch", step, fmt.Sprintf("action step app config %q does not match %q", s.AppConfigID, cfg.ID))
			}
			if s.ActionWorkflowConfigID != a.ID {
				add("action_step.parent_mismatch", step, "action step does not belong to its loaded action workflow config")
			}
			if s.ConnectedGithubVCSConfig != nil || s.PublicGitVCSConfig != nil {
				if hasCron {
					add("action_step.git_unsupported", step, "cron action steps must be inline for air-gapped bundles; use inline_contents")
				} else {
					warn("action_step.git_excluded", step, "Git-backed non-cron action is excluded from the air-gapped bundle")
				}
			} else if s.InlineContents != "" || s.Command != "" {
				warn("action_step.inline_review", step, "inline action content requires review")
			}
		}
	}
	for _, id := range cfg.ActionIDs {
		if id == "" {
			continue
		}
		if !actions[id] {
			add("action.config_missing", "action:"+id, "AppConfig action member has no workflow config")
		}
	}

	urls := []struct{ member, value string }{
		{"stack:runner", cfg.StackConfig.RunnerNestedTemplateURL},
		{"stack:vpc", cfg.StackConfig.VPCNestedTemplateURL},
	}
	for i, stack := range cfg.StackConfig.CustomNestedStacks {
		member := "stack:custom:" + logical(stack.Name, "", i)
		urls = append(urls, struct{ member, value string }{member, stack.TemplateURL})
		if stack.ContentsHash == "" {
			add("stack_asset.unpinned_custom", member, "custom stack assets must be uploaded and content-addressed before bundling")
		}
	}
	for _, u := range urls {
		if strings.Contains(u.value, "{{") || strings.Contains(u.value, "}}") {
			add("stack_asset.templated_url", u.member, "templated stack asset URLs are not supported")
		}
	}

	finish(&r)
	return r
}

func qualifyComponent(c *app.ComponentConfigConnection, member string, add, warn func(string, string, string)) {
	surfaces := 0
	for _, present := range []bool{c.TerraformModuleComponentConfig != nil, c.HelmComponentConfig != nil, c.ExternalImageComponentConfig != nil, c.DockerBuildComponentConfig != nil, c.JobComponentConfig != nil, c.KubernetesManifestComponentConfig != nil, c.PulumiComponentConfig != nil} {
		if present {
			surfaces++
		}
	}
	if surfaces != 1 {
		add("component.source_config_invalid", member, "component must have exactly one source config")
	}
	switch c.Type {
	case app.ComponentTypeTerraformModule:
		if c.TerraformModuleComponentConfig == nil {
			add("component.source_config_mismatch", member, "terraform component source config is missing")
		}
	case app.ComponentTypeHelmChart:
		if c.HelmComponentConfig == nil {
			add("component.source_config_mismatch", member, "Helm component source config is missing")
		}
		warn("component.embedded_images_undetected", member, "images referenced by Helm chart values are not detected or bundled; declare them as external_image components")
	case app.ComponentTypeExternalImage:
		if c.ExternalImageComponentConfig == nil {
			add("component.source_config_mismatch", member, "external image component source config is missing")
		}
	case app.ComponentTypeKubernetesManifest:
		if c.KubernetesManifestComponentConfig == nil {
			add("component.source_config_mismatch", member, "Kubernetes manifest source config is missing")
		}
		warn("component.embedded_images_undetected", member, "images referenced by Kubernetes manifests are not detected or bundled; declare them as external_image components")
	case app.ComponentTypeDockerBuild:
		add("component.docker_build_unsupported", member, "docker_build components are not supported")
	case app.ComponentTypePulumi:
		add("component.pulumi_unsupported", member, "Pulumi components are not supported")
	case app.ComponentTypeJob:
		add("component.job_unsupported", member, "job components are not supported")
	default:
		add("component.type_unsupported", member, fmt.Sprintf("component type %q is unsupported", c.Type))
	}
}

func componentMember(c *app.ComponentConfigConnection, i int) string {
	return "component:" + logical(c.ComponentName, c.ComponentID, i)
}
func actionMember(a *app.ActionWorkflowConfig, i int) string {
	return "action:" + logical(a.ActionWorkflow.Name, a.ActionWorkflowID, i)
}
func logical(name, id string, i int) string {
	if name != "" {
		return name
	}
	if id != "" {
		return id
	}
	return fmt.Sprintf("#%d", i)
}

func finish(r *QualificationReport) {
	less := func(findings []Finding) {
		sort.SliceStable(findings, func(i, j int) bool {
			if findings[i].Code != findings[j].Code {
				return findings[i].Code < findings[j].Code
			}
			if findings[i].Member != findings[j].Member {
				return findings[i].Member < findings[j].Member
			}
			return findings[i].Message < findings[j].Message
		})
	}
	less(r.Violations)
	less(r.Warnings)
	r.Qualified = len(r.Violations) == 0
}
