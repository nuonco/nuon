package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/invopop/jsonschema"
)

// SchemaDiff represents differences between local and remote schemas
type SchemaDiff struct {
	SchemaType     string
	MissingLocally []string // Properties in remote but not in local
	MissingRemote  []string // Properties in local but not in remote (not an issue for validation)
	TypeMismatches []TypeMismatch
}

// TypeMismatch represents a type difference for a property
type TypeMismatch struct {
	Property   string
	LocalType  string
	RemoteType string
}

// HasMeaningfulDiff returns true if there are differences that would cause validation failures
// We only care about properties missing locally (remote has new fields the CLI doesn't know about)
func (d *SchemaDiff) HasMeaningfulDiff() bool {
	return len(d.MissingLocally) > 0 || len(d.TypeMismatches) > 0
}

func (d *SchemaDiff) String() string {
	if !d.HasMeaningfulDiff() {
		return ""
	}

	var parts []string
	if len(d.MissingLocally) > 0 {
		parts = append(parts, fmt.Sprintf("New fields in API: %s", strings.Join(d.MissingLocally, ", ")))
	}
	if len(d.TypeMismatches) > 0 {
		var mismatches []string
		for _, m := range d.TypeMismatches {
			mismatches = append(mismatches, fmt.Sprintf("%s (local: %s, remote: %s)", m.Property, m.LocalType, m.RemoteType))
		}
		parts = append(parts, fmt.Sprintf("type changes: %s", strings.Join(mismatches, ", ")))
	}
	return strings.Join(parts, "; ")
}

// FetchRemoteSchema fetches a schema from the API
func FetchRemoteSchema(ctx context.Context, apiURL, schemaType string) (*jsonschema.Schema, error) {
	url := fmt.Sprintf("%s/v1/general/config-schema?type=%s", strings.TrimSuffix(apiURL, "/"), schemaType)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching schema: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var schema jsonschema.Schema
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}

	return &schema, nil
}

// CompareSchemas compares a local schema against a remote schema and returns differences
func CompareSchemas(local, remote *jsonschema.Schema) *SchemaDiff {
	diff := &SchemaDiff{}

	localProps := extractProperties(local)
	remoteProps := extractProperties(remote)

	// Find properties missing locally (new in remote/API)
	for prop := range remoteProps {
		if _, exists := localProps[prop]; !exists {
			diff.MissingLocally = append(diff.MissingLocally, prop)
		}
	}

	// Find properties missing in remote (deprecated/removed from API)
	for prop := range localProps {
		if _, exists := remoteProps[prop]; !exists {
			diff.MissingRemote = append(diff.MissingRemote, prop)
		}
	}

	// Check for type mismatches on common properties
	for prop, localType := range localProps {
		if remoteType, exists := remoteProps[prop]; exists {
			if localType != remoteType {
				diff.TypeMismatches = append(diff.TypeMismatches, TypeMismatch{
					Property:   prop,
					LocalType:  localType,
					RemoteType: remoteType,
				})
			}
		}
	}

	// Sort for consistent output
	sort.Strings(diff.MissingLocally)
	sort.Strings(diff.MissingRemote)
	sort.Slice(diff.TypeMismatches, func(i, j int) bool {
		return diff.TypeMismatches[i].Property < diff.TypeMismatches[j].Property
	})

	return diff
}

// extractProperties recursively extracts all property names and their types from a schema
func extractProperties(s *jsonschema.Schema) map[string]string {
	props := make(map[string]string)
	if s == nil {
		return props
	}

	extractPropertiesRecursive(s, "", props)
	return props
}

func extractPropertiesRecursive(s *jsonschema.Schema, prefix string, props map[string]string) {
	if s == nil {
		return
	}

	// Handle properties at this level
	if s.Properties != nil {
		for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
			name := pair.Key
			prop := pair.Value

			fullName := name
			if prefix != "" {
				fullName = prefix + "." + name
			}

			// Determine the type
			propType := determineType(prop)
			props[fullName] = propType

			// Recurse into nested objects
			if prop.Type == "object" || prop.Properties != nil {
				extractPropertiesRecursive(prop, fullName, props)
			}

			// Handle arrays with object items
			if prop.Type == "array" && prop.Items != nil {
				extractPropertiesRecursive(prop.Items, fullName+"[]", props)
			}
		}
	}

	// Handle definitions/schemas referenced
	if s.Definitions != nil {
		for name, def := range s.Definitions {
			defPrefix := "#/definitions/" + name
			if prefix != "" {
				defPrefix = prefix + "." + defPrefix
			}
			extractPropertiesRecursive(def, defPrefix, props)
		}
	}
}

func determineType(s *jsonschema.Schema) string {
	if s == nil {
		return "unknown"
	}

	if s.Type != "" {
		return s.Type
	}

	// Handle references
	if s.Ref != "" {
		return "ref:" + s.Ref
	}

	// Handle oneOf/anyOf/allOf
	if len(s.OneOf) > 0 {
		return "oneOf"
	}
	if len(s.AnyOf) > 0 {
		return "anyOf"
	}
	if len(s.AllOf) > 0 {
		return "allOf"
	}

	return "unknown"
}

// CheckSchemaCompatibility fetches the remote schema and compares it with the local schema
// Returns a diff if there are meaningful differences, nil otherwise
func CheckSchemaCompatibility(ctx context.Context, apiURL, schemaType string) (*SchemaDiff, error) {
	// Get local schema
	localSchemaFn, ok := SchemaMapping[schemaType]
	if !ok {
		return nil, fmt.Errorf("unknown schema type: %s", schemaType)
	}

	localSchema, err := localSchemaFn()
	if err != nil {
		return nil, fmt.Errorf("generating local schema: %w", err)
	}

	// Fetch remote schema
	remoteSchema, err := FetchRemoteSchema(ctx, apiURL, schemaType)
	if err != nil {
		// Don't fail validation if we can't reach the API
		return nil, nil
	}

	diff := CompareSchemas(localSchema, remoteSchema)
	diff.SchemaType = schemaType

	if diff.HasMeaningfulDiff() {
		return diff, nil
	}

	return nil, nil
}
