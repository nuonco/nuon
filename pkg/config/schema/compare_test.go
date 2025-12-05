package schema

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func TestCompareSchemas_NoChanges(t *testing.T) {
	schema1 := createTestSchema(map[string]string{
		"name":  "string",
		"count": "integer",
	})
	schema2 := createTestSchema(map[string]string{
		"name":  "string",
		"count": "integer",
	})

	diff := CompareSchemas(schema1, schema2)

	if diff.HasMeaningfulDiff() {
		t.Errorf("expected no meaningful diff, got: %s", diff.String())
	}
	if len(diff.MissingLocally) != 0 {
		t.Errorf("expected no missing locally, got: %v", diff.MissingLocally)
	}
	if len(diff.MissingRemote) != 0 {
		t.Errorf("expected no missing remote, got: %v", diff.MissingRemote)
	}
	if len(diff.TypeMismatches) != 0 {
		t.Errorf("expected no type mismatches, got: %v", diff.TypeMismatches)
	}
}

func TestCompareSchemas_MissingLocally(t *testing.T) {
	local := createTestSchema(map[string]string{
		"name": "string",
	})
	remote := createTestSchema(map[string]string{
		"name":      "string",
		"new_field": "string",
	})

	diff := CompareSchemas(local, remote)

	if !diff.HasMeaningfulDiff() {
		t.Error("expected meaningful diff when remote has new fields")
	}
	if len(diff.MissingLocally) != 1 || diff.MissingLocally[0] != "new_field" {
		t.Errorf("expected MissingLocally to contain 'new_field', got: %v", diff.MissingLocally)
	}
}

func TestCompareSchemas_MissingRemote(t *testing.T) {
	local := createTestSchema(map[string]string{
		"name":         "string",
		"deprecated_x": "string",
	})
	remote := createTestSchema(map[string]string{
		"name": "string",
	})

	diff := CompareSchemas(local, remote)

	if diff.HasMeaningfulDiff() {
		t.Error("MissingRemote should NOT be considered a meaningful diff")
	}
	if len(diff.MissingRemote) != 1 || diff.MissingRemote[0] != "deprecated_x" {
		t.Errorf("expected MissingRemote to contain 'deprecated_x', got: %v", diff.MissingRemote)
	}
}

func TestCompareSchemas_TypeMismatch(t *testing.T) {
	local := createTestSchema(map[string]string{
		"count": "string",
	})
	remote := createTestSchema(map[string]string{
		"count": "integer",
	})

	diff := CompareSchemas(local, remote)

	if !diff.HasMeaningfulDiff() {
		t.Error("expected meaningful diff when types differ")
	}
	if len(diff.TypeMismatches) != 1 {
		t.Fatalf("expected 1 type mismatch, got: %d", len(diff.TypeMismatches))
	}
	if diff.TypeMismatches[0].Property != "count" {
		t.Errorf("expected type mismatch for 'count', got: %s", diff.TypeMismatches[0].Property)
	}
	if diff.TypeMismatches[0].LocalType != "string" || diff.TypeMismatches[0].RemoteType != "integer" {
		t.Errorf("expected string->integer mismatch, got: %s->%s",
			diff.TypeMismatches[0].LocalType, diff.TypeMismatches[0].RemoteType)
	}
}

func TestCompareSchemas_NilSchemas(t *testing.T) {
	diff := CompareSchemas(nil, nil)
	if diff.HasMeaningfulDiff() {
		t.Error("nil schemas should not produce meaningful diff")
	}

	schema := createTestSchema(map[string]string{"name": "string"})
	diff = CompareSchemas(nil, schema)
	if len(diff.MissingLocally) != 1 {
		t.Errorf("expected 1 missing locally, got: %d", len(diff.MissingLocally))
	}

	diff = CompareSchemas(schema, nil)
	if len(diff.MissingRemote) != 1 {
		t.Errorf("expected 1 missing remote, got: %d", len(diff.MissingRemote))
	}
}

func TestCompareSchemas_ComplexSchema(t *testing.T) {
	local := &jsonschema.Schema{
		Type: "object",
		Properties: orderedmap.New[string, *jsonschema.Schema](
			orderedmap.WithInitialData(
				orderedmap.Pair[string, *jsonschema.Schema]{
					Key: "metadata",
					Value: &jsonschema.Schema{
						Type: "object",
						Properties: orderedmap.New[string, *jsonschema.Schema](
							orderedmap.WithInitialData(
								orderedmap.Pair[string, *jsonschema.Schema]{
									Key:   "name",
									Value: &jsonschema.Schema{Type: "string"},
								},
							),
						),
					},
				},
				orderedmap.Pair[string, *jsonschema.Schema]{
					Key:   "version",
					Value: &jsonschema.Schema{Type: "string"},
				},
			),
		),
	}

	remote := &jsonschema.Schema{
		Type: "object",
		Properties: orderedmap.New[string, *jsonschema.Schema](
			orderedmap.WithInitialData(
				orderedmap.Pair[string, *jsonschema.Schema]{
					Key: "metadata",
					Value: &jsonschema.Schema{
						Type: "object",
						Properties: orderedmap.New[string, *jsonschema.Schema](
							orderedmap.WithInitialData(
								orderedmap.Pair[string, *jsonschema.Schema]{
									Key:   "name",
									Value: &jsonschema.Schema{Type: "string"},
								},
								orderedmap.Pair[string, *jsonschema.Schema]{
									Key:   "labels",
									Value: &jsonschema.Schema{Type: "object"},
								},
							),
						),
					},
				},
				orderedmap.Pair[string, *jsonschema.Schema]{
					Key:   "version",
					Value: &jsonschema.Schema{Type: "string"},
				},
			),
		),
	}

	diff := CompareSchemas(local, remote)

	if !diff.HasMeaningfulDiff() {
		t.Error("expected meaningful diff for nested schema change")
	}

	found := false
	for _, m := range diff.MissingLocally {
		if m == "metadata.labels" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'metadata.labels' in MissingLocally, got: %v", diff.MissingLocally)
	}
}

func TestCompareSchemas_ArrayItems(t *testing.T) {
	local := &jsonschema.Schema{
		Type: "object",
		Properties: orderedmap.New[string, *jsonschema.Schema](
			orderedmap.WithInitialData(
				orderedmap.Pair[string, *jsonschema.Schema]{
					Key: "items",
					Value: &jsonschema.Schema{
						Type: "array",
						Items: &jsonschema.Schema{
							Type: "object",
							Properties: orderedmap.New[string, *jsonschema.Schema](
								orderedmap.WithInitialData(
									orderedmap.Pair[string, *jsonschema.Schema]{
										Key:   "id",
										Value: &jsonschema.Schema{Type: "string"},
									},
								),
							),
						},
					},
				},
			),
		),
	}

	remote := &jsonschema.Schema{
		Type: "object",
		Properties: orderedmap.New[string, *jsonschema.Schema](
			orderedmap.WithInitialData(
				orderedmap.Pair[string, *jsonschema.Schema]{
					Key: "items",
					Value: &jsonschema.Schema{
						Type: "array",
						Items: &jsonschema.Schema{
							Type: "object",
							Properties: orderedmap.New[string, *jsonschema.Schema](
								orderedmap.WithInitialData(
									orderedmap.Pair[string, *jsonschema.Schema]{
										Key:   "id",
										Value: &jsonschema.Schema{Type: "string"},
									},
									orderedmap.Pair[string, *jsonschema.Schema]{
										Key:   "status",
										Value: &jsonschema.Schema{Type: "string"},
									},
								),
							),
						},
					},
				},
			),
		),
	}

	diff := CompareSchemas(local, remote)

	if !diff.HasMeaningfulDiff() {
		t.Error("expected meaningful diff for array item schema change")
	}

	found := false
	for _, m := range diff.MissingLocally {
		if m == "items[].status" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'items[].status' in MissingLocally, got: %v", diff.MissingLocally)
	}
}

func TestSchemaDiff_String(t *testing.T) {
	diff := &SchemaDiff{
		MissingLocally: []string{"field1", "field2"},
		TypeMismatches: []TypeMismatch{
			{Property: "count", LocalType: "string", RemoteType: "integer"},
		},
	}

	str := diff.String()

	if str == "" {
		t.Error("expected non-empty string for diff with changes")
	}
	if !contains(str, "field1") || !contains(str, "field2") {
		t.Errorf("expected string to contain missing fields, got: %s", str)
	}
	if !contains(str, "count") {
		t.Errorf("expected string to contain type mismatch info, got: %s", str)
	}
}

func TestSchemaDiff_String_Empty(t *testing.T) {
	diff := &SchemaDiff{}
	str := diff.String()
	if str != "" {
		t.Errorf("expected empty string for no diff, got: %s", str)
	}
}

func TestFetchRemoteSchema(t *testing.T) {
	testSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/general/config-schema" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "full" {
			t.Errorf("unexpected type param: %s", r.URL.Query().Get("type"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testSchema)
	}))
	defer server.Close()

	ctx := context.Background()
	schema, err := FetchRemoteSchema(ctx, server.URL, "full")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}
	if schema.Type != "object" {
		t.Errorf("expected type 'object', got: %s", schema.Type)
	}
}

func TestFetchRemoteSchema_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx := context.Background()
	_, err := FetchRemoteSchema(ctx, server.URL, "full")

	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestFetchRemoteSchema_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	ctx := context.Background()
	_, err := FetchRemoteSchema(ctx, server.URL, "full")

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestCheckSchemaCompatibility_NoAPI(t *testing.T) {
	ctx := context.Background()
	diff, err := CheckSchemaCompatibility(ctx, "http://localhost:99999", "full")

	if err != nil {
		t.Errorf("expected nil error when API is unreachable, got: %v", err)
	}
	if diff != nil {
		t.Errorf("expected nil diff when API is unreachable, got: %v", diff)
	}
}

func TestCheckSchemaCompatibility_UnknownSchemaType(t *testing.T) {
	ctx := context.Background()
	_, err := CheckSchemaCompatibility(ctx, "http://localhost", "unknown-type")

	if err == nil {
		t.Error("expected error for unknown schema type")
	}
}

func createTestSchema(props map[string]string) *jsonschema.Schema {
	properties := orderedmap.New[string, *jsonschema.Schema]()
	for name, typ := range props {
		properties.Set(name, &jsonschema.Schema{Type: typ})
	}
	return &jsonschema.Schema{
		Type:       "object",
		Properties: properties,
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
