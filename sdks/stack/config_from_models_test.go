package stack

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
	"github.com/nuonco/nuon/sdks/stack/models"
)

// jsonPaths returns every json tag reachable from t, dotted by nesting. Maps and
// slices are descended into by element type, so a nested role struct contributes
// "aws.break_glass_roles.permissions".
func jsonPaths(t reflect.Type, prefix string, seen map[reflect.Type]bool) []string {
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Map {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct || seen[t] {
		return nil
	}
	seen[t] = true
	defer delete(seen, t)

	var out []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := strings.Split(f.Tag.Get("json"), ",")[0]
		if tag == "" || tag == "-" {
			continue
		}
		path := tag
		if prefix != "" {
			path = prefix + "." + tag
		}
		out = append(out, path)
		out = append(out, jsonPaths(f.Type, path, seen)...)
	}

	return out
}

// configFromModel copies fields by re-encoding rather than by name, so it is only
// total while the wire tags remain a subset of Config's. If ctl-api renames or adds
// a tag Config does not have, the value would be dropped silently; this fails first.
func TestWireFieldsHaveCoreFields(t *testing.T) {
	wire := jsonPaths(reflect.TypeOf(models.AppInstallerSDKConfig{}), "", map[reflect.Type]bool{})
	require.NotEmpty(t, wire)

	core := jsonPaths(reflect.TypeOf(core.Config{}), "", map[reflect.Type]bool{})
	have := make(map[string]bool, len(core))
	for _, p := range core {
		have[p] = true
	}

	var missing []string
	for _, p := range wire {
		if !have[p] {
			missing = append(missing, p)
		}
	}
	sort.Strings(missing)

	assert.Empty(t, missing,
		"wire fields with no core.Config counterpart — they would be dropped silently")
}
