package statetelemetry

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"go.opentelemetry.io/otel/attribute"
)

var installLabelKeyPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,63}$`)

func installLabelKeys(csv string) ([]string, error) {
	if strings.TrimSpace(csv) == "" {
		return nil, nil
	}
	keys := strings.Split(csv, ",")
	if len(keys) > 8 {
		return nil, fmt.Errorf("TELEMETRY_INSTALL_LABELS allows at most 8 keys")
	}
	seen := make(map[string]bool, len(keys))
	normalize := strings.NewReplacer(".", "_", "-", "_")
	for i, key := range keys {
		key = strings.TrimSpace(key)
		normalized := normalize.Replace(key)
		prefix, _, _ := strings.Cut(strings.ToLower(normalized), "_")
		if !installLabelKeyPattern.MatchString(key) || prefix == "nuon" || prefix == "otel" || prefix == "service" || seen[normalized] {
			return nil, fmt.Errorf("invalid TELEMETRY_INSTALL_LABELS key at position %d: require unique nonreserved keys, 1–64 ASCII characters, without Prometheus-name collisions", i+1)
		}
		seen[normalized] = true
		keys[i] = key
	}
	return keys, nil
}

func installInfoMetric(labelKeys []string) populationMetric {
	return populationMetric{
		name: "nuon.install.info", unit: "1", detail: true,
		description: "1 per retained install with bounded name, active app-branch identity, sandbox mode and allowlisted materialized labels; metadata status reports omitted invalid values.",
		args:        []any{labelKeys},
		source: `SELECT i.org_id, i.app_id, i.id,
    jsonb_build_object('name', left(record.name, 257), 'branch_id', b.id,
        'branch_name', left(b.name, 257), 'sandbox_mode', i.sandbox_mode,
        'labels', (SELECT jsonb_object_agg(key, CASE WHEN jsonb_typeof(record.labels->key) = 'string'
            THEN to_jsonb(left(record.labels->>key, 257)) ELSE 'null'::jsonb END)
            FROM unnest($1::text[]) AS keys(key) WHERE record.labels ? key))::text,
    1::bigint
FROM live_installs i
JOIN installs record ON record.id = i.id
LEFT JOIN LATERAL (
    SELECT c.app_branch_id FROM install_app_branch_connections c
    WHERE c.install_id = i.id AND c.org_id = i.org_id AND c.active AND c.deleted_at = 0
    ORDER BY c.created_at DESC, c.id DESC LIMIT 1
) connection ON true
LEFT JOIN app_branches b ON b.id = connection.app_branch_id AND b.org_id = i.org_id AND b.app_id = i.app_id AND b.deleted_at = 0`,
		attributes: func(value string) ([]attribute.KeyValue, error) {
			var info struct {
				Name        *string            `json:"name"`
				BranchID    *string            `json:"branch_id"`
				BranchName  *string            `json:"branch_name"`
				SandboxMode bool               `json:"sandbox_mode"`
				Labels      map[string]*string `json:"labels"`
			}
			if err := json.Unmarshal([]byte(value), &info); err != nil {
				return nil, fmt.Errorf("invalid install info metadata")
			}
			attrs := []attribute.KeyValue{attribute.Bool("nuon.install.sandbox_mode", info.SandboxMode)}
			status := "complete"
			add := func(key string, value *string) {
				if value == nil || *value == "" {
					return
				}
				if len(*value) > 256 || !utf8.ValidString(*value) || strings.ContainsFunc(*value, unicode.IsControl) {
					status = "omitted"
					return
				}
				attrs = append(attrs, attribute.String(key, *value))
			}
			add("nuon.install.name", info.Name)
			add("nuon.app_branch.id", info.BranchID)
			add("nuon.app_branch.name", info.BranchName)
			for _, key := range labelKeys {
				if value, ok := info.Labels[key]; ok {
					if value == nil {
						status = "omitted"
					} else {
						add("nuon.install.label."+key, value)
					}
				}
			}
			return append(attrs, attribute.String("nuon.install.metadata.status", status)), nil
		},
	}
}
