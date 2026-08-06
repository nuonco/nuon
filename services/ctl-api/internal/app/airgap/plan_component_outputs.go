package airgap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	runnerairgap "github.com/nuonco/nuon/pkg/runner/airgap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// componentOutputRefKey identifies one cross-component output reference in
// the app config: the producing component's name plus the dotted path into
// its outputs.
type componentOutputRefKey struct {
	Component string
	Path      string
}

func (r componentOutputRefKey) String() string { return r.Component + "." + r.Path }

func (r componentOutputRefKey) token() string {
	return runnerairgap.ComponentOutputPlaceholder(r.Component, r.Path)
}

func extractComponentOutputRefs(datas ...[]byte) []componentOutputRefKey {
	seen := map[componentOutputRefKey]bool{}
	for _, data := range datas {
		for _, match := range componentOutputRef.FindAllSubmatch(withoutDocFields(data), -1) {
			ref := componentOutputRefKey{Component: string(match[1]), Path: strings.Trim(string(match[2]), ".")}
			if ref.Path == "" {
				continue
			}
			seen[ref] = true
		}
	}
	refs := make([]componentOutputRefKey, 0, len(seen))
	for ref := range seen {
		refs = append(refs, ref)
	}
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Component != refs[j].Component {
			return refs[i].Component < refs[j].Component
		}
		return refs[i].Path < refs[j].Path
	})
	return refs
}

func validateComponentOutputRefs(refs []componentOutputRefKey, connections []app.ComponentConfigConnection, report *QualificationReport) error {
	known := map[string]bool{}
	for _, connection := range connections {
		known[connection.ComponentName] = true
	}
	var unknown []string
	for _, ref := range refs {
		if known[ref.Component] {
			continue
		}
		unknown = append(unknown, ref.String())
		addQualificationViolation(report, "template.component_output_unknown_component", "component:"+ref.Component, fmt.Sprintf("template references outputs of unknown component %q", ref.Component))
	}
	if len(unknown) > 0 {
		return fmt.Errorf("templates reference outputs of unknown component(s): %s", strings.Join(unknown, ", "))
	}
	return nil
}

// normalizeComponentTokenPadding collapses string values that are one
// component-output token surrounded only by whitespace to the bare token.
// Accidental padding inside app config templates (e.g. "{{...}} ") would
// otherwise survive substitution and corrupt bound values such as ARNs.
func normalizeComponentTokenPadding(data json.RawMessage, refs []componentOutputRefKey) (json.RawMessage, error) {
	if len(refs) == 0 {
		return data, nil
	}
	tokens := map[string]bool{}
	for _, ref := range refs {
		tokens[ref.token()] = true
	}
	var decoded any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, fmt.Errorf("decode composite plan for token normalization: %w", err)
	}
	changed := false
	decoded = rewriteJSONStrings(decoded, func(s string) string {
		trimmed := strings.TrimSpace(s)
		if trimmed != s && tokens[trimmed] {
			changed = true
			return trimmed
		}
		return s
	})
	if !changed {
		return data, nil
	}
	normalized, err := json.Marshal(decoded)
	if err != nil {
		return nil, fmt.Errorf("encode composite plan after token normalization: %w", err)
	}
	return normalized, nil
}

func rewriteJSONStrings(node any, rewrite func(string) string) any {
	switch value := node.(type) {
	case string:
		return rewrite(value)
	case map[string]any:
		for key, item := range value {
			value[key] = rewriteJSONStrings(item, rewrite)
		}
		return value
	case []any:
		for i, item := range value {
			value[i] = rewriteJSONStrings(item, rewrite)
		}
		return value
	default:
		return node
	}
}

// componentTokenRefs returns the references whose placeholder tokens occur in
// a rendered plan, i.e. the component outputs that plan actually consumes.
func componentTokenRefs(data []byte, refs []componentOutputRefKey) []componentOutputRefKey {
	var present []componentOutputRefKey
	for _, ref := range refs {
		if bytes.Contains(data, []byte(ref.token())) {
			present = append(present, ref)
		}
	}
	return present
}

func componentRefStrings(refs []componentOutputRefKey) []string {
	names := make([]string, 0, len(refs))
	for _, ref := range refs {
		names = append(names, ref.String())
	}
	return names
}

// seedComponentOutputs builds the .nuon.components state subtree so the
// online renderer (missingkey=error) resolves cross-component references to
// placeholder tokens that the offline runner binds after producers apply.
func seedComponentOutputs(refs []componentOutputRefKey) (map[string]any, error) {
	components := map[string]any{}
	for _, ref := range refs {
		entry, ok := components[ref.Component].(map[string]any)
		if !ok {
			entry = map[string]any{"populated": true, "name": ref.Component, "outputs": map[string]any{}}
			components[ref.Component] = entry
		}
		if err := seedPlaceholderPath(entry["outputs"].(map[string]any), ref.Path, ref.token()); err != nil {
			return nil, fmt.Errorf("component %s output references: %w", ref.Component, err)
		}
	}
	return components, nil
}

// seedPlaceholderPath sets a placeholder token at a dotted path inside a
// nested map, creating intermediate maps as needed. Conflicting references
// (a path that is both a leaf and a prefix of a deeper path) are rejected.
func seedPlaceholderPath(root map[string]any, path, token string) error {
	node := root
	segments := strings.Split(path, ".")
	for i, segment := range segments {
		last := i == len(segments)-1
		existing, exists := node[segment]
		if last {
			if exists {
				if existing == token {
					return nil
				}
				return fmt.Errorf("conflicting references under %s", strings.Join(segments[:i+1], "."))
			}
			node[segment] = token
			return nil
		}
		if !exists {
			next := map[string]any{}
			node[segment] = next
			node = next
			continue
		}
		next, ok := existing.(map[string]any)
		if !ok {
			return fmt.Errorf("conflicting references under %s", strings.Join(segments[:i+1], "."))
		}
		node = next
	}
	return nil
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
