package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The json tags on this package restate the Go field names so swag can read them. They must stay a
// serialization no-op: the CLI writes intermediate config blobs with stdlib json.Marshal and the
// server reads them back, so any key rename would make every stored blob unreadable.
func TestIntermediateConfigRoundTripsByteForByte(t *testing.T) {
	for _, name := range []string{"app_config_all_fields.json", "app_config_realistic.json"} {
		t.Run(name, func(t *testing.T) {
			stored, err := os.ReadFile(filepath.Join("testdata", name))
			require.NoError(t, err)

			var cfg AppConfig
			require.NoError(t, json.Unmarshal(stored, &cfg))

			remarshalled, err := json.Marshal(&cfg)
			require.NoError(t, err)

			require.Equal(t, string(stored), string(remarshalled))
		})
	}
}

func TestJSONTagNamesMatchGoFieldNames(t *testing.T) {
	preTagged := map[string]string{
		"AppBranchConfig.PostDeployRunbooks":   "post_deploy_runbooks",
		"AppBranchConfig.IgnoreChangesRegex":   "ignore_changes_regex",
		"AppBranchConfig.SendStatusesOnIgnore": "send_statuses_on_ignore",
		"AzureCustomStack.Schema":              "$schema",
		"AzureCustomStack.ContentVersion":      "contentVersion",
		"AzureCustomStack.Resources":           "resources",
		"CustomNestedStack.Name":               "name",
		"CustomNestedStack.Index":              "index",
		"CustomNestedStack.Contents":           "contents",
		"CustomNestedStack.ContentsHash":       "contents_hash",
		"CustomNestedStack.TemplateURL":        "template_url",
		"CustomNestedStack.TemplateSourceURL":  "template_source_url",
		"CustomNestedStack.Status":             "status",
		"CustomNestedStack.Parameters":         "parameters",
		"SourceArchive.SchemaVersion":          "schema_version",
		"SourceArchive.Members":                "members",
		"SourceArchive.Files":                  "files",
	}

	seen := map[string]bool{}
	var walk func(t *testing.T, typ reflect.Type, path map[reflect.Type]bool)
	walk = func(t *testing.T, typ reflect.Type, path map[reflect.Type]bool) {
		for typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map {
			typ = typ.Elem()
		}
		if typ.Kind() != reflect.Struct || path[typ] || typ.PkgPath() != reflect.TypeFor[AppConfig]().PkgPath() {
			return
		}
		path[typ] = true
		defer delete(path, typ)

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if !field.IsExported() {
				continue
			}

			key := typ.Name() + "." + field.Name
			seen[key] = true
			tag, ok := field.Tag.Lookup("json")
			require.Truef(t, ok, "%s has no json tag", key)

			name := strings.Split(tag, ",")[0]
			if name == "-" {
				continue
			}
			if want, isPreTagged := preTagged[key]; isPreTagged {
				require.Equalf(t, want, name, "%s changed its pre-existing json name", key)
				continue
			}
			require.Equalf(t, field.Name, name, "%s must serialize under its Go field name", key)

			walk(t, field.Type, path)
		}
	}

	walk(t, reflect.TypeFor[AppConfig](), map[reflect.Type]bool{})
	require.Greater(t, len(seen), 200)
}
