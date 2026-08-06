package airgap

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

// renderStepPlan reproduces control-plane plan chaining and environment rebinding for offline execution.
func (c *Client) renderStepPlan(step *Step) (json.RawMessage, error) {
	return c.renderPlan(step.ID, step.CompositePlan, step.PlanFromStep)
}

func (c *Client) renderPlan(id string, compositePlan json.RawMessage, planFromStep string) (json.RawMessage, error) {
	var plan map[string]any
	if err := json.Unmarshal(compositePlan, &plan); err != nil {
		return nil, fmt.Errorf("decode composite plan for %s: %w", id, err)
	}

	if planFromStep != "" {
		contents, err := c.chainedPlanContents(planFromStep)
		if err != nil {
			return nil, fmt.Errorf("step %s: %w", id, err)
		}
		injected := false
		for _, key := range []string{"sandbox_run_plan", "deploy_plan"} {
			if inner, ok := plan[key].(map[string]any); ok {
				inner["apply_plan_contents"] = contents
				delete(inner, "apply_plan_display")
				injected = true
			}
		}
		if !injected {
			return nil, fmt.Errorf("step %s: composite plan has no sandbox_run_plan or deploy_plan to chain into", id)
		}
	}

	c.mu.Lock()
	stackOutputs := c.installStackOutputs
	installInputs := c.installInputs
	c.mu.Unlock()
	if len(stackOutputs) > 0 {
		rebindInstallStackOutputs(plan, stackOutputs)
	}

	sandboxOutputs := c.latestSandboxOutputs()
	if len(sandboxOutputs) > 0 {
		rebindSandboxOutputs(plan, sandboxOutputs)
	}

	// Envelope-wide snapshots provide substitutions for plans such as OCI sync that carry no snapshot themselves.
	globalSubs := map[string]string{}
	refInstallStack, refSandbox := c.referenceSnapshots()
	if len(stackOutputs) > 0 && len(refInstallStack) > 0 {
		collectOutputSubstitutions(globalSubs, refInstallStack, stackOutputs)
	}
	if len(sandboxOutputs) > 0 && len(refSandbox) > 0 {
		collectDeepSubstitutions(globalSubs, refSandbox, sandboxOutputs)
	}
	substituteStrings(plan, globalSubs)

	if cluster := c.latestClusterOutput(); cluster != nil {
		rebindClusterInfo(plan, cluster)
	}

	SubstituteInputValues(plan, ResolveInputValues(c.envelope.Inputs, installInputs))
	if missing := UnresolvedInputPlaceholders(plan, c.envelope.Inputs); len(missing) > 0 {
		return nil, fmt.Errorf("step %s: no value for install input(s) %s; supply them via --install-inputs", id, strings.Join(missing, ", "))
	}

	if err := c.bindComponentOutputs(id, plan); err != nil {
		return nil, err
	}

	if c.envelope.ForceDefaultCloudAuth {
		existingRoles := map[string]bool{}
		collectIAMRoleARNs(existingRoles, stackOutputs)
		collectIAMRoleARNs(existingRoles, sandboxOutputs)
		forceDefaultCloudAuth(plan, existingRoles)
	}

	patched, err := json.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("encode composite plan for %s: %w", id, err)
	}
	return patched, nil
}

// bindComponentOutputs substitutes cross-component output placeholders with
// values from the producing steps' recorded terraform outputs. Envelope
// dependency ordering guarantees producers apply before consumers render, so
// a token that cannot be resolved is a hard error.
func (c *Client) bindComponentOutputs(id string, plan map[string]any) error {
	if len(c.envelope.OutputBindings) == 0 {
		return nil
	}
	values := map[string]string{}
	for _, binding := range c.envelope.OutputBindings {
		if !containsString(plan, binding.Token) {
			continue
		}
		outputs, ok, err := c.stepOutputs(binding.StepID)
		if err != nil {
			return fmt.Errorf("step %s: read outputs of step %q for component %q: %w", id, binding.StepID, binding.ComponentName, err)
		}
		if !ok {
			continue
		}
		value, ok := ResolveOutputPath(outputs, binding.OutputPath)
		if !ok {
			return fmt.Errorf("step %s: component %q has no output %q (step %q)", id, binding.ComponentName, binding.OutputPath, binding.StepID)
		}
		rendered, err := OutputValueString(value)
		if err != nil {
			return fmt.Errorf("step %s: render component %q output %q: %w", id, binding.ComponentName, binding.OutputPath, err)
		}
		values[binding.Token] = rendered
	}
	SubstituteComponentOutputs(plan, values)
	if missing := UnresolvedComponentOutputs(plan, c.envelope.OutputBindings); len(missing) > 0 {
		refs := make([]string, 0, len(missing))
		for _, binding := range missing {
			refs = append(refs, binding.ComponentName+"."+binding.OutputPath)
		}
		return fmt.Errorf("step %s: component output(s) %s are not available yet; the producing component has not applied", id, strings.Join(refs, ", "))
	}
	return nil
}

// stepOutputs returns the recorded execution outputs for a step, reading
// through to the disk store for outputs recorded by previous invocations.
func (c *Client) stepOutputs(stepID string) (map[string]any, bool, error) {
	c.mu.Lock()
	raw, ok := c.status.Outputs[stepID]
	c.mu.Unlock()
	if !ok {
		var err error
		raw, ok, err = c.store.ReadOutputs(stepID)
		if err != nil || !ok {
			return nil, false, err
		}
		var persisted struct {
			Outputs map[string]any `json:"outputs"`
		}
		if err := json.Unmarshal(raw, &persisted); err != nil {
			return nil, false, fmt.Errorf("decode persisted outputs for step %q: %w", stepID, err)
		}
		return persisted.Outputs, persisted.Outputs != nil, nil
	}
	var outputs map[string]any
	if err := json.Unmarshal(raw, &outputs); err != nil {
		return nil, false, fmt.Errorf("decode outputs for step %q: %w", stepID, err)
	}
	return outputs, outputs != nil, nil
}

// chainedPlanContents converts a stored execution result into the
// std-base64 apply_plan_contents encoding the runner's apply handlers expect,
// mirroring ctl-api's RunnerJobExecutionResult.GetContentsB64String.
func (c *Client) chainedPlanContents(fromStepID string) (string, error) {
	c.mu.Lock()
	req := c.results[fromStepID]
	c.mu.Unlock()
	if req == nil {
		// Results from previous invocations exist only in the disk store.
		raw, ok, err := c.store.ReadResult(fromStepID)
		if err != nil {
			return "", fmt.Errorf("read persisted result for plan step %q: %w", fromStepID, err)
		}
		if ok {
			var persisted models.ServiceCreateRunnerJobExecutionResultRequest
			if err := json.Unmarshal(raw, &persisted); err != nil {
				return "", fmt.Errorf("decode persisted result for plan step %q: %w", fromStepID, err)
			}
			req = &persisted
			c.mu.Lock()
			c.results[fromStepID] = req
			c.mu.Unlock()
		}
	}
	if req == nil {
		return "", fmt.Errorf("no execution result recorded for plan step %q", fromStepID)
	}
	if req.ContentsCompressed != "" {
		raw, err := base64.URLEncoding.DecodeString(req.ContentsCompressed)
		if err != nil {
			return "", fmt.Errorf("decode plan contents from step %q: %w", fromStepID, err)
		}
		return base64.StdEncoding.EncodeToString(raw), nil
	}
	if req.Contents != "" {
		return req.Contents, nil
	}
	return "", fmt.Errorf("execution result for plan step %q has no plan contents", fromStepID)
}

// referenceSnapshots returns the vendor reference environment's install-stack
// and sandbox output snapshots, collected from every step plan in the
// envelope (first value wins per key). Plans that carry rendered reference
// values but no snapshot of their own (oci-sync's dst_registry, for example)
// are rebound by aligning these snapshots against the target environment's
// actual outputs.
func (c *Client) referenceSnapshots() (map[string]any, map[string]any) {
	c.refSnapshotsOnce.Do(func() {
		c.refInstallStack = map[string]any{}
		c.refSandbox = map[string]any{}
		for i := range c.envelope.Steps {
			var plan map[string]any
			if err := json.Unmarshal(c.envelope.Steps[i].CompositePlan, &plan); err != nil {
				continue
			}
			collectReferenceSnapshots(plan, c.refInstallStack, c.refSandbox)
		}
	})
	return c.refInstallStack, c.refSandbox
}

func collectReferenceSnapshots(node any, installStack, sandbox map[string]any) {
	merge := func(dst map[string]any, container any) {
		block, ok := container.(map[string]any)
		if !ok {
			return
		}
		outputs, ok := block["outputs"].(map[string]any)
		if !ok {
			return
		}
		for key, val := range outputs {
			if _, exists := dst[key]; !exists {
				dst[key] = val
			}
		}
	}
	switch v := node.(type) {
	case map[string]any:
		merge(installStack, v["install_stack"])
		merge(sandbox, v["sandbox"])
		for _, child := range v {
			collectReferenceSnapshots(child, installStack, sandbox)
		}
	case []any:
		for _, child := range v {
			collectReferenceSnapshots(child, installStack, sandbox)
		}
	}
}

func (c *Client) latestSandboxOutputs() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := len(c.status.Steps) - 1; i >= 0; i-- {
		st := c.status.Steps[i]
		if st.Status != string(models.AppRunnerJobStatusFinished) {
			continue
		}
		step, err := c.findStep(st.ID)
		if err != nil || step.JobGroup != "sandbox" {
			continue
		}
		raw, ok := c.status.Outputs[st.ID]
		if !ok {
			continue
		}
		var outputs map[string]any
		if err := json.Unmarshal(raw, &outputs); err != nil {
			continue
		}
		if len(outputs) > 0 {
			return outputs
		}
	}
	return nil
}

func (c *Client) latestClusterOutput() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := len(c.status.Steps) - 1; i >= 0; i-- {
		st := c.status.Steps[i]
		if st.Status != string(models.AppRunnerJobStatusFinished) {
			continue
		}
		raw, ok := c.status.Outputs[st.ID]
		if !ok {
			continue
		}
		var outputs map[string]any
		if err := json.Unmarshal(raw, &outputs); err != nil {
			continue
		}
		if cluster, ok := outputs["cluster"].(map[string]any); ok {
			return cluster
		}
	}
	return nil
}

// rebindClusterInfo rewrites every cluster_info block in the plan to point at
// the given sandbox cluster output. Handles both AWS/GCP-shaped
// (certificate_authority_data/endpoint) and Azure-shaped
// (cluster_ca_certificate/host) outputs.
func rebindClusterInfo(node any, cluster map[string]any) {
	str := func(k string) string { s, _ := cluster[k].(string); return s }
	name := str("name")
	endpoint := str("endpoint")
	if endpoint == "" {
		endpoint = str("host")
	}
	ca := str("certificate_authority_data")
	if ca == "" {
		ca = str("cluster_ca_certificate")
	}
	if name == "" || endpoint == "" || ca == "" {
		return
	}

	var walk func(any)
	walk = func(n any) {
		switch v := n.(type) {
		case map[string]any:
			if info, ok := v["cluster_info"].(map[string]any); ok && !clusterInfoBoundToComponent(info) {
				info["id"] = name
				info["endpoint"] = endpoint
				info["ca_data"] = ca
				delete(info, "kube_config")
			}
			for _, child := range v {
				walk(child)
			}
		case []any:
			for _, child := range v {
				walk(child)
			}
		}
	}
	walk(node)
}

// clusterInfoBoundToComponent reports whether a cluster_info block targets a
// peer component's cluster (a declared kubernetes_context) rather than the
// sandbox default. Those blocks carry component-output placeholders that are
// bound from the producing step's outputs, so the sandbox cluster rebind must
// leave them alone.
func clusterInfoBoundToComponent(info map[string]any) bool {
	for _, key := range []string{"id", "endpoint", "ca_data"} {
		if value, ok := info[key].(string); ok && strings.Contains(value, componentOutputPlaceholderPrefix) {
			return true
		}
	}
	return false
}

// rebindInstallStackOutputs rewrites every install_stack outputs block to the
// outputs of the stack that exists in the target account, then rewrites the
// rendered values elsewhere in the plan that the control plane interpolated
// from the stale outputs (terraform vars, state snapshots). Replacement only
// happens on exact value match — a whole string or a comma-separated element —
// so unrelated strings that merely contain an old value are never touched.
func rebindInstallStackOutputs(node any, outputs map[string]any) {
	subs := map[string]string{}
	var collect func(any)
	collect = func(n any) {
		switch v := n.(type) {
		case map[string]any:
			if stack, ok := v["install_stack"].(map[string]any); ok {
				if old, ok := stack["outputs"].(map[string]any); ok {
					collectOutputSubstitutions(subs, old, outputs)
					for key, fresh := range outputs {
						if _, wasMap := old[key].(map[string]any); wasMap {
							if m := decodeJSONObject(fresh); m != nil {
								old[key] = m
								continue
							}
						}
						old[key] = fresh
					}
				}
			}
			for _, child := range v {
				collect(child)
			}
		case []any:
			for _, child := range v {
				collect(child)
			}
		}
	}
	collect(node)
	substituteStrings(node, subs)
}

// collectOutputSubstitutions records old→new value pairs for every output key
// whose string value changed. Comma-joined list values (subnet lists) also
// contribute element-wise pairs so individually rendered elements rebind too.
// Map-valued outputs (break_glass_role_arns, custom_role_arns) contribute
// pairs by matching entry keys: the compile seeds name→token maps and the
// phone-home payload carries name→ARN, so each token rebinds to the ARN of
// the identically named role.
func collectOutputSubstitutions(subs map[string]string, old, fresh map[string]any) {
	for key, ov := range old {
		if om, ok := ov.(map[string]any); ok {
			fm := decodeJSONObject(fresh[key])
			for name, oldVal := range om {
				os, _ := oldVal.(string)
				ns, _ := fm[name].(string)
				if os != "" && ns != "" && os != ns {
					subs[os] = ns
				}
			}
			continue
		}
		os, ok := ov.(string)
		if !ok || os == "" {
			continue
		}
		ns, ok := fresh[key].(string)
		if !ok || ns == "" || ns == os {
			continue
		}
		subs[os] = ns
		oldParts := strings.Split(os, ",")
		newParts := strings.Split(ns, ",")
		if len(oldParts) > 1 && len(oldParts) == len(newParts) {
			for i := range oldParts {
				op := strings.TrimSpace(oldParts[i])
				np := strings.TrimSpace(newParts[i])
				if op != "" && np != "" && op != np {
					subs[op] = np
				}
			}
		}
	}
}

// decodeJSONObject returns the value as a map, decoding the JSON-object
// string form produced by SetInstallStackOutputs normalization.
func decodeJSONObject(v any) map[string]any {
	switch t := v.(type) {
	case map[string]any:
		return t
	case string:
		var m map[string]any
		if err := json.Unmarshal([]byte(t), &m); err == nil {
			return m
		}
	}
	return nil
}

// rebindSandboxOutputs rewrites values the control plane rendered from the
// vendor reference install's sandbox outputs (Route53 zone IDs, domains,
// ECR URLs, ...) to the outputs the sandbox apply actually produced in this
// environment. The reference snapshot lives in the composite plan under
// state's `sandbox.outputs`; substitutions are collected by structurally
// aligning that snapshot with the local outputs, then applied on exact value
// match across the whole plan, and the snapshot itself is replaced so
// anything reading it sees the local values.
func rebindSandboxOutputs(node any, fresh map[string]any) {
	subs := map[string]string{}
	var collect func(any)
	collect = func(n any) {
		switch v := n.(type) {
		case map[string]any:
			if sandbox, ok := v["sandbox"].(map[string]any); ok {
				if old, ok := sandbox["outputs"].(map[string]any); ok {
					collectDeepSubstitutions(subs, old, fresh)
					for key, val := range fresh {
						old[key] = val
					}
				}
			}
			for _, child := range v {
				collect(child)
			}
		case []any:
			for _, child := range v {
				collect(child)
			}
		}
	}
	collect(node)
	substituteStrings(node, subs)
}

// minRebindValueLength guards value-based substitution against generic short
// strings ("1", "true", availability-zone letters) that would rewrite
// unrelated plan fields on a coincidental match.
const minRebindValueLength = 6

func collectDeepSubstitutions(subs map[string]string, old, fresh any) {
	switch ov := old.(type) {
	case map[string]any:
		fm, ok := fresh.(map[string]any)
		if !ok {
			return
		}
		for key, child := range ov {
			collectDeepSubstitutions(subs, child, fm[key])
		}
	case []any:
		fl, ok := fresh.([]any)
		if !ok {
			return
		}
		for i := range ov {
			if i >= len(fl) {
				return
			}
			collectDeepSubstitutions(subs, ov[i], fl[i])
		}
	case string:
		fs, ok := fresh.(string)
		if !ok || fs == "" || ov == fs || len(ov) < minRebindValueLength {
			return
		}
		subs[ov] = fs
	}
}

func substituteStrings(node any, subs map[string]string) {
	if len(subs) == 0 {
		return
	}
	var rewrite func(any)
	rewrite = func(n any) {
		switch v := n.(type) {
		case map[string]any:
			for key, child := range v {
				if s, ok := child.(string); ok {
					v[key] = substituteExact(s, subs)
					continue
				}
				rewrite(child)
			}
		case []any:
			for i, child := range v {
				if s, ok := child.(string); ok {
					v[i] = substituteExact(s, subs)
					continue
				}
				rewrite(child)
			}
		}
	}
	rewrite(node)
}

func substituteExact(s string, subs map[string]string) string {
	if replacement, ok := subs[s]; ok {
		return replacement
	}
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		changed := false
		for i, part := range parts {
			if replacement, ok := subs[part]; ok {
				parts[i] = replacement
				changed = true
			}
		}
		if changed {
			s = strings.Join(parts, ",")
		}
	}
	// Compile-time placeholder tokens are globally unique markers, so unlike
	// value-based substitutions they are safe to rewrite when embedded inside
	// larger strings ("*.<token>", "https://<token>/path").
	if strings.Contains(s, airgapTokenPrefix) {
		for token, replacement := range subs {
			if isAirgapToken(token) && strings.Contains(s, token) {
				s = strings.ReplaceAll(s, token, replacement)
			}
		}
	}
	return s
}

const airgapTokenPrefix = "__NUON_AIRGAP_"

var airgapTokenPattern = regexp.MustCompile(`^__NUON_AIRGAP_[A-Za-z0-9_]+__$`)

func isAirgapToken(s string) bool {
	return airgapTokenPattern.MatchString(s)
}

// collectIAMRoleARNs records every IAM role ARN found in the given outputs'
// string leaves. These are roles known to exist in the target environment
// (created by the install stack or a finished sandbox apply), so cloud-auth
// blocks rebound to them can keep their assume-role configuration.
func collectIAMRoleARNs(arns map[string]bool, node any) {
	switch v := node.(type) {
	case map[string]any:
		for _, child := range v {
			collectIAMRoleARNs(arns, child)
		}
	case []any:
		for _, child := range v {
			collectIAMRoleARNs(arns, child)
		}
	case string:
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "arn:aws:iam::") && strings.Contains(part, ":role/") {
				arns[part] = true
			}
		}
	}
}

// forceDefaultCloudAuth rewrites every cloud-auth block (identified by the
// presence of both assume_role and use_default keys) to use the process's
// ambient credentials instead of roles provisioned by a sandbox that may not
// exist yet in an offline run. Blocks whose assume-role ARN is known to exist
// in this environment (rebound from install stack or sandbox outputs) are
// left intact so plans run with the role the vendor intended.
func forceDefaultCloudAuth(node any, existingRoleARNs map[string]bool) {
	var walk func(any)
	walk = func(n any) {
		switch v := n.(type) {
		case map[string]any:
			_, hasAssume := v["assume_role"]
			_, hasDefault := v["use_default"]
			if hasAssume && hasDefault && !hasExistingAssumeRole(v, existingRoleARNs) {
				v["use_default"] = true
				v["assume_role"] = nil
				v["static"] = nil
			}
			for _, child := range v {
				walk(child)
			}
		case []any:
			for _, child := range v {
				walk(child)
			}
		}
	}
	walk(node)
}

func hasExistingAssumeRole(authBlock map[string]any, existingRoleARNs map[string]bool) bool {
	assume, ok := authBlock["assume_role"].(map[string]any)
	if !ok {
		return false
	}
	arn, _ := assume["role_arn"].(string)
	return arn != "" && existingRoleARNs[arn]
}
