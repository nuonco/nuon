package service

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/diff"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	mcpPlanMaxBytes     = 4 << 20
	mcpPlanDiffMaxRunes = 4000
)

type mcpPlanDiffChange struct {
	Action  string `json:"action"`
	Kind    string `json:"kind,omitempty"`
	Address string `json:"address"`
	Diff    string `json:"diff,omitempty"`
	Clipped bool   `json:"clipped,omitempty"`
}

type mcpPlanDiffBuild struct {
	Changes      []mcpPlanDiffChange
	TooLarge     bool
	ChangesState string
}

func buildApprovalPlanDiff(approvalType app.WorkflowStepApprovalType, contents, action, search string) (mcpPlanDiffBuild, error) {
	if !planDiffSupported(approvalType) {
		return mcpPlanDiffBuild{ChangesState: string(app.StepChangeStateUnsupported)}, nil
	}
	if len(contents) > mcpPlanMaxBytes {
		return mcpPlanDiffBuild{TooLarge: true}, nil
	}

	var changes []mcpPlanDiffChange
	var err error
	switch approvalType {
	case app.TerraformPlanApprovalType:
		changes, err = terraformPlanDiffs(contents)
	case app.PulumiApprovalType:
		changes, err = pulumiPlanDiffs(contents)
	case app.HelmApprovalApprovalType:
		changes, err = resourcePlanDiffs(contents, "helm_content_diff")
	case app.KubernetesManifestApprovalType:
		changes, err = resourcePlanDiffs(contents, "k8s_content_diff")
	case app.AppBranchPlanApprovalType:
		changes, err = appBranchPlanDiffs(contents)
	default:
		return mcpPlanDiffBuild{ChangesState: string(app.StepChangeStateUnsupported)}, nil
	}
	if err != nil {
		return mcpPlanDiffBuild{}, err
	}

	return mcpPlanDiffBuild{Changes: filterPlanDiffs(changes, action, search)}, nil
}

func pagePlanDiffs(changes []mcpPlanDiffChange, limit, offset int) ([]mcpPlanDiffChange, bool) {
	if offset < 0 {
		offset = 0
	}
	if offset > len(changes) {
		offset = len(changes)
	}
	end := offset + limit
	hasMore := end < len(changes)
	if end > len(changes) {
		end = len(changes)
	}
	page := changes[offset:end]
	if page == nil {
		page = []mcpPlanDiffChange{}
	}
	return page, hasMore
}

func planDiffSupported(approvalType app.WorkflowStepApprovalType) bool {
	switch approvalType {
	case app.TerraformPlanApprovalType,
		app.PulumiApprovalType,
		app.HelmApprovalApprovalType,
		app.KubernetesManifestApprovalType,
		app.AppBranchPlanApprovalType:
		return true
	default:
		return false
	}
}

func filterPlanDiffs(changes []mcpPlanDiffChange, action, search string) []mcpPlanDiffChange {
	want := normalizePlanAction(action)
	search = strings.ToLower(strings.TrimSpace(search))
	out := make([]mcpPlanDiffChange, 0, len(changes))
	for _, change := range changes {
		if want == "" && (change.Action == "no-op" || change.Action == "read") {
			continue
		}
		if want != "" && change.Action != want {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(change.Address+" "+change.Kind), search) {
			continue
		}
		out = append(out, change)
	}
	return out
}

func normalizePlanAction(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "", "all":
		return ""
	case "add", "added", "create", "created":
		return "create"
	case "change", "changed", "update", "updated", "modify", "modified":
		return "update"
	case "destroy", "destroyed", "delete", "deleted", "remove", "removed":
		return "delete"
	case "replace":
		return "replace"
	case "noop", "no-op", "same", "unchanged":
		return "no-op"
	case "read":
		return "read"
	default:
		return strings.ToLower(strings.TrimSpace(action))
	}
}

func finishPlanDiff(action, kind, address, diffText string) mcpPlanDiffChange {
	clippedText, clipped := clipPlanDiff(diffText)
	return mcpPlanDiffChange{
		Action:  action,
		Kind:    kind,
		Address: address,
		Diff:    clippedText,
		Clipped: clipped,
	}
}

func clipPlanDiff(diffText string) (string, bool) {
	runes := []rune(diffText)
	if len(runes) <= mcpPlanDiffMaxRunes {
		return diffText, false
	}
	return string(runes[:mcpPlanDiffMaxRunes]) + "\n... (clipped)", true
}

type tfPlanDocument struct {
	ResourceDrift   []tfPlanResource        `json:"resource_drift"`
	ResourceChanges []tfPlanResource        `json:"resource_changes"`
	OutputChanges   map[string]tfPlanChange `json:"output_changes"`
}

type tfPlanResource struct {
	Address string       `json:"address"`
	Type    string       `json:"type"`
	Change  tfPlanChange `json:"change"`
}

type tfPlanChange struct {
	Actions         []string `json:"actions"`
	Before          any      `json:"before"`
	After           any      `json:"after"`
	AfterUnknown    any      `json:"after_unknown"`
	BeforeSensitive any      `json:"before_sensitive"`
	AfterSensitive  any      `json:"after_sensitive"`
}

func terraformPlanDiffs(contents string) ([]mcpPlanDiffChange, error) {
	var plan tfPlanDocument
	if err := json.Unmarshal([]byte(contents), &plan); err != nil {
		return nil, errors.Wrap(err, "unable to parse terraform plan")
	}

	changes := make([]mcpPlanDiffChange, 0, len(plan.ResourceDrift)+len(plan.ResourceChanges)+len(plan.OutputChanges))
	for _, resource := range plan.ResourceDrift {
		if change, ok := terraformChange("drift", resource.Type, resource.Address, resource.Change); ok {
			changes = append(changes, change)
		}
	}
	for _, resource := range plan.ResourceChanges {
		if change, ok := terraformChange(resource.Type, resource.Type, resource.Address, resource.Change); ok {
			changes = append(changes, change)
		}
	}

	names := make([]string, 0, len(plan.OutputChanges))
	for name := range plan.OutputChanges {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if change, ok := terraformChange("output", "output", name, plan.OutputChanges[name]); ok {
			changes = append(changes, change)
		}
	}
	return changes, nil
}

func terraformChange(kind, resourceType, address string, change tfPlanChange) (mcpPlanDiffChange, bool) {
	action := terraformAction(change.Actions)
	before := applySensitive(change.Before, change.BeforeSensitive)
	after := applyUnknown(applySensitive(change.After, change.AfterSensitive), change.AfterUnknown)
	if action == "no-op" && reflect.DeepEqual(before, after) {
		return mcpPlanDiffChange{}, false
	}
	if kind == "drift" {
		kind = "drift"
		if resourceType != "" {
			kind = "drift:" + resourceType
		}
	}
	lines := diffValues(before, after, "", nil, 0)
	lines = appendSensitiveChanges(change.Before, change.After, before, after, "", lines)
	return finishPlanDiff(action, kind, address, strings.Join(lines, "\n")), true
}

func appendSensitiveChanges(rawBefore, rawAfter, before, after any, path string, lines []string) []string {
	if rawMap, ok := rawBefore.(map[string]any); ok {
		if rawAfterMap, ok := rawAfter.(map[string]any); ok {
			beforeMap, _ := before.(map[string]any)
			afterMap, _ := after.(map[string]any)
			for _, key := range unionKeys(rawMap, rawAfterMap) {
				beforeValue, _ := lookup(beforeMap, key)
				afterValue, _ := lookup(afterMap, key)
				rawBeforeValue, _ := lookup(rawMap, key)
				rawAfterValue, _ := lookup(rawAfterMap, key)
				lines = appendSensitiveChanges(rawBeforeValue, rawAfterValue, beforeValue, afterValue, joinPath(path, key), lines)
			}
			return lines
		}
	}
	if reflect.DeepEqual(rawBefore, rawAfter) || !reflect.DeepEqual(before, after) {
		return lines
	}
	return append(lines, "~ "+displayPath(path)+": (sensitive value)")
}

func terraformAction(actions []string) string {
	hasCreate := false
	hasDelete := false
	for _, action := range actions {
		switch action {
		case "create":
			hasCreate = true
		case "delete":
			hasDelete = true
		}
	}
	if hasCreate && hasDelete {
		return "replace"
	}
	order := []string{"delete", "create", "update", "read", "no-op"}
	for _, want := range order {
		for _, action := range actions {
			if action == want {
				return action
			}
		}
	}
	if len(actions) == 0 {
		return "no-op"
	}
	return actions[0]
}

type pulumiPlanDocument struct {
	ResourceChanges []pulumiPlanResource `json:"resource_changes"`
}

type pulumiPlanResource struct {
	URN       string         `json:"urn"`
	Type      string         `json:"type"`
	Name      string         `json:"name"`
	Action    string         `json:"action"`
	Diffs     []string       `json:"diffs"`
	OldInputs map[string]any `json:"old_inputs"`
	NewInputs map[string]any `json:"new_inputs"`
}

func pulumiPlanDiffs(contents string) ([]mcpPlanDiffChange, error) {
	var plan pulumiPlanDocument
	if err := json.Unmarshal([]byte(contents), &plan); err != nil {
		return nil, errors.Wrap(err, "unable to parse pulumi plan")
	}
	changes := make([]mcpPlanDiffChange, 0, len(plan.ResourceChanges))
	for _, resource := range plan.ResourceChanges {
		action := normalizePlanAction(resource.Action)
		if action == "" {
			action = resource.Action
		}
		address := resource.Name
		if address == "" {
			address = resource.URN
		}
		oldInputs := any(resource.OldInputs)
		newInputs := any(resource.NewInputs)
		if len(resource.Diffs) > 0 {
			oldInputs = pickKeys(resource.OldInputs, resource.Diffs)
			newInputs = pickKeys(resource.NewInputs, resource.Diffs)
		}
		lines := diffValues(oldInputs, newInputs, "", nil, 0)
		changes = append(changes, finishPlanDiff(action, resource.Type, address, strings.Join(lines, "\n")))
	}
	return changes, nil
}

func pickKeys(inputs map[string]any, keys []string) map[string]any {
	if inputs == nil {
		return nil
	}
	out := make(map[string]any, len(keys))
	for _, key := range keys {
		if value, ok := inputs[key]; ok {
			out[key] = value
		}
	}
	return out
}

type resourcePlanItem struct {
	Name      string             `json:"name"`
	Namespace string             `json:"namespace"`
	Kind      string             `json:"kind"`
	Op        string             `json:"op"`
	Type      diff.DiffEntryType `json:"type"`
	Error     string             `json:"error,omitempty"`
	Before    string             `json:"before,omitempty"`
	After     string             `json:"after,omitempty"`
	Entries   []diff.DiffEntry   `json:"entries"`
}

func resourcePlanDiffs(contents, field string) ([]mcpPlanDiffChange, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(contents), &raw); err != nil {
		return nil, errors.Wrap(err, "unable to parse plan")
	}
	payload := raw[field]
	if len(payload) == 0 || string(payload) == "null" {
		return []mcpPlanDiffChange{}, nil
	}
	var items []resourcePlanItem
	if err := json.Unmarshal(payload, &items); err != nil {
		return nil, errors.Wrapf(err, "unable to parse %s", field)
	}
	changes := make([]mcpPlanDiffChange, 0, len(items))
	for _, item := range items {
		if item.Error != "" {
			changes = append(changes, finishPlanDiff("error", item.Kind, resourceAddress(item), item.Error))
			continue
		}
		action := resourceAction(item)
		text := resourceDiffText(item)
		changes = append(changes, finishPlanDiff(action, item.Kind, resourceAddress(item), text))
	}
	return changes, nil
}

func resourceAddress(item resourcePlanItem) string {
	parts := make([]string, 0, 3)
	if item.Namespace != "" {
		parts = append(parts, item.Namespace)
	}
	if item.Kind != "" {
		parts = append(parts, item.Kind)
	}
	if item.Name != "" {
		parts = append(parts, item.Name)
	}
	if len(parts) == 0 {
		return item.Name
	}
	return strings.Join(parts, "/")
}

func resourceAction(item resourcePlanItem) string {
	if strings.EqualFold(item.Op, "delete") || item.Type == diff.EntryRemoved {
		return "delete"
	}
	switch item.Type {
	case diff.EntryAdded:
		return "create"
	case diff.EntryModified:
		return "update"
	case diff.EntryUnchanged:
		return "no-op"
	case diff.EntryError:
		return "error"
	default:
		if item.Op != "" {
			if action := normalizePlanAction(item.Op); action != "" {
				return action
			}
		}
		return "update"
	}
}

func resourceDiffText(item resourcePlanItem) string {
	if len(item.Entries) > 0 {
		return entryDiffText(item.Entries)
	}
	if item.Before != "" || item.After != "" {
		return unifiedLineDiff(item.Before, item.After)
	}
	return ""
}

func entryDiffText(entries []diff.DiffEntry) string {
	var b strings.Builder
	for _, entry := range entries {
		switch entry.Type {
		case diff.EntryUnchanged:
			continue
		case diff.EntryRemoved:
			fmt.Fprintf(&b, "- %s\n", entryLine(entry, true))
		case diff.EntryAdded:
			fmt.Fprintf(&b, "+ %s\n", entryLine(entry, false))
		case diff.EntryModified:
			fmt.Fprintf(&b, "- %s\n+ %s\n", entryLine(entry, true), entryLine(entry, false))
		case diff.EntryError:
			fmt.Fprintf(&b, "! %s\n", entryLine(entry, false))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func entryLine(entry diff.DiffEntry, before bool) string {
	value := entry.Applied
	if before {
		value = entry.Original
	}
	if entry.Path != "" {
		if text := scalarText(value); text != "" {
			return entry.Path + ": " + text
		}
	}
	if entry.Payload != "" {
		return entry.Payload
	}
	return scalarText(value)
}

type appBranchPlanDocument struct {
	Installs []appBranchPlanInstall `json:"installs"`
}

type appBranchPlanInstall struct {
	InstallID   string                 `json:"install_id"`
	InstallName string                 `json:"install_name"`
	Diff        *app.InstallConfigDiff `json:"diff"`
}

func appBranchPlanDiffs(contents string) ([]mcpPlanDiffChange, error) {
	var plan appBranchPlanDocument
	if err := json.Unmarshal([]byte(contents), &plan); err != nil {
		return nil, errors.Wrap(err, "unable to parse app branch plan")
	}
	changes := make([]mcpPlanDiffChange, 0, len(plan.Installs))
	for _, install := range plan.Installs {
		address := install.InstallName
		if address == "" {
			address = install.InstallID
		}
		if install.Diff == nil {
			changes = append(changes, finishPlanDiff("no-op", "install", address, ""))
			continue
		}
		lines := make([]string, 0)
		lines = appendComponentLines(lines, "+", "added", install.Diff.Added)
		lines = appendComponentLines(lines, "-", "removed", install.Diff.Removed)
		lines = appendComponentLines(lines, "~", "changed", install.Diff.Changed)
		if install.Diff.SandboxChanged {
			lines = append(lines, "sandbox changed")
		}
		if install.Diff.StackChanged {
			lines = append(lines, "stack changed")
		}
		action := "update"
		if len(lines) == 0 {
			action = "no-op"
		}
		changes = append(changes, finishPlanDiff(action, "install", address, strings.Join(lines, "\n")))
	}
	return changes, nil
}

func appendComponentLines(lines []string, prefix, label string, entries []app.ComponentDiffEntry) []string {
	for _, entry := range entries {
		name := entry.ComponentName
		if name == "" {
			name = entry.ComponentID
		}
		lines = append(lines, prefix+" "+label+" "+name)
	}
	return lines
}

func applySensitive(value, mask any) any {
	switch typed := mask.(type) {
	case bool:
		if typed {
			return "(sensitive value)"
		}
		return value
	case map[string]any:
		src, _ := value.(map[string]any)
		out := map[string]any{}
		for key, child := range src {
			out[key] = child
		}
		for key, childMask := range typed {
			out[key] = applySensitive(out[key], childMask)
		}
		if len(out) == 0 && value == nil {
			return value
		}
		return out
	case []any:
		src, _ := value.([]any)
		out := append([]any(nil), src...)
		for i, childMask := range typed {
			if i < len(out) {
				out[i] = applySensitive(out[i], childMask)
			}
		}
		return out
	default:
		return value
	}
}

func applyUnknown(value, unknown any) any {
	switch typed := unknown.(type) {
	case bool:
		if typed {
			return "(known after apply)"
		}
		return value
	case map[string]any:
		src, _ := value.(map[string]any)
		if src == nil && len(typed) == 0 {
			return value
		}
		out := map[string]any{}
		for key, child := range src {
			out[key] = child
		}
		for key, child := range typed {
			out[key] = applyUnknown(out[key], child)
		}
		return out
	case []any:
		src, _ := value.([]any)
		out := append([]any(nil), src...)
		for i, child := range typed {
			if i < len(out) {
				out[i] = applyUnknown(out[i], child)
			}
		}
		return out
	default:
		return value
	}
}

func diffValues(before, after any, path string, lines []string, depth int) []string {
	if reflect.DeepEqual(before, after) {
		return lines
	}
	if depth > 8 {
		return append(lines, "~ "+displayPath(path)+": (nested change)")
	}
	beforeMap, beforeIsMap := before.(map[string]any)
	afterMap, afterIsMap := after.(map[string]any)
	if beforeIsMap || afterIsMap {
		keys := unionKeys(beforeMap, afterMap)
		for _, key := range keys {
			next := joinPath(path, key)
			beforeValue, beforeOK := lookup(beforeMap, key)
			afterValue, afterOK := lookup(afterMap, key)
			switch {
			case !beforeOK:
				lines = appendAdded(lines, next, afterValue, depth+1)
			case !afterOK:
				lines = appendRemoved(lines, next, beforeValue, depth+1)
			default:
				lines = diffValues(beforeValue, afterValue, next, lines, depth+1)
			}
		}
		return lines
	}
	beforeSlice, beforeIsSlice := before.([]any)
	afterSlice, afterIsSlice := after.([]any)
	if beforeIsSlice || afterIsSlice {
		n := len(beforeSlice)
		if len(afterSlice) > n {
			n = len(afterSlice)
		}
		for i := 0; i < n; i++ {
			next := joinPath(path, strconv.Itoa(i))
			switch {
			case i >= len(beforeSlice):
				lines = appendAdded(lines, next, afterSlice[i], depth+1)
			case i >= len(afterSlice):
				lines = appendRemoved(lines, next, beforeSlice[i], depth+1)
			default:
				lines = diffValues(beforeSlice[i], afterSlice[i], next, lines, depth+1)
			}
		}
		return lines
	}
	label := displayPath(path)
	if before != nil {
		lines = append(lines, "- "+label+": "+scalarText(before))
	}
	if after != nil {
		lines = append(lines, "+ "+label+": "+scalarText(after))
	}
	return lines
}

func lookup(values map[string]any, key string) (any, bool) {
	if values == nil {
		return nil, false
	}
	value, ok := values[key]
	return value, ok
}

func appendAdded(lines []string, path string, value any, depth int) []string {
	if depth > 8 {
		return append(lines, "+ "+displayPath(path)+": (nested change)")
	}
	if typed, ok := value.(map[string]any); ok {
		keys := unionKeys(typed, nil)
		if len(keys) == 0 {
			return append(lines, "+ "+displayPath(path)+": {}")
		}
		for _, key := range keys {
			lines = appendAdded(lines, joinPath(path, key), typed[key], depth+1)
		}
		return lines
	}
	if typed, ok := value.([]any); ok {
		if len(typed) == 0 {
			return append(lines, "+ "+displayPath(path)+": []")
		}
		for i, child := range typed {
			lines = appendAdded(lines, joinPath(path, strconv.Itoa(i)), child, depth+1)
		}
		return lines
	}
	return append(lines, "+ "+displayPath(path)+": "+scalarText(value))
}

func appendRemoved(lines []string, path string, value any, depth int) []string {
	if depth > 8 {
		return append(lines, "- "+displayPath(path)+": (nested change)")
	}
	if typed, ok := value.(map[string]any); ok {
		keys := unionKeys(typed, nil)
		if len(keys) == 0 {
			return append(lines, "- "+displayPath(path)+": {}")
		}
		for _, key := range keys {
			lines = appendRemoved(lines, joinPath(path, key), typed[key], depth+1)
		}
		return lines
	}
	if typed, ok := value.([]any); ok {
		if len(typed) == 0 {
			return append(lines, "- "+displayPath(path)+": []")
		}
		for i, child := range typed {
			lines = appendRemoved(lines, joinPath(path, strconv.Itoa(i)), child, depth+1)
		}
		return lines
	}
	return append(lines, "- "+displayPath(path)+": "+scalarText(value))
}

func unionKeys(left, right map[string]any) []string {
	seen := map[string]struct{}{}
	for key := range left {
		seen[key] = struct{}{}
	}
	for key := range right {
		seen[key] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func joinPath(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

func displayPath(path string) string {
	if path == "" {
		return "value"
	}
	return path
}

func scalarText(value any) string {
	if value == nil {
		return "null"
	}
	switch typed := value.(type) {
	case string:
		return strconv.Quote(typed)
	case json.Number:
		return typed.String()
	default:
		raw, err := json.Marshal(typed)
		if err != nil {
			return strconv.Quote(fmt.Sprint(typed))
		}
		text := string(raw)
		if len(text) > 180 {
			return text[:180] + "... (clipped)"
		}
		return text
	}
}

func unifiedLineDiff(before, after string) string {
	if before == after {
		return ""
	}
	beforeLines := splitDiffLines(before)
	afterLines := splitDiffLines(after)
	if len(beforeLines) > 120 || len(afterLines) > 120 {
		var b strings.Builder
		for _, line := range beforeLines {
			fmt.Fprintf(&b, "- %s\n", line)
		}
		for _, line := range afterLines {
			fmt.Fprintf(&b, "+ %s\n", line)
		}
		return strings.TrimRight(b.String(), "\n")
	}
	return lcsLineDiff(beforeLines, afterLines)
}

func splitDiffLines(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, "\n")
}

func lcsLineDiff(before, after []string) string {
	n, m := len(before), len(after)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if before[i] == after[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	var b strings.Builder
	i, j := 0, 0
	for i < n && j < m {
		if before[i] == after[j] {
			i++
			j++
			continue
		}
		if dp[i+1][j] >= dp[i][j+1] {
			fmt.Fprintf(&b, "- %s\n", before[i])
			i++
		} else {
			fmt.Fprintf(&b, "+ %s\n", after[j])
			j++
		}
	}
	for ; i < n; i++ {
		fmt.Fprintf(&b, "- %s\n", before[i])
	}
	for ; j < m; j++ {
		fmt.Fprintf(&b, "+ %s\n", after[j])
	}
	return strings.TrimRight(b.String(), "\n")
}
