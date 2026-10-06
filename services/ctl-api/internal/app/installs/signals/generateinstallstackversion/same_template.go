package generateinstallstackversion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// SameTemplate reports whether two stack templates match once each version's own phone-home ID and version ID are
// blanked out. Templates are compared as decoded JSON because stored contents are jsonb, which reorders keys. Tags and
// parameter group lists are compared as sets, since versions rendered before they were sorted list them in random order.
func SameTemplate(a []byte, aVersion *app.InstallStackVersion, b []byte, bVersion *app.InstallStackVersion) bool {
	if len(a) == 0 || len(b) == 0 || aVersion == nil || bVersion == nil {
		return false
	}
	ca, ok := canonicalTemplate(a, aVersion)
	if !ok {
		return false
	}
	cb, ok := canonicalTemplate(b, bVersion)
	if !ok {
		return false
	}
	return bytes.Equal(ca, cb)
}

func canonicalTemplate(tmpl []byte, version *app.InstallStackVersion) ([]byte, bool) {
	for _, id := range []string{version.PhoneHomeID, version.ID} {
		if id != "" {
			tmpl = bytes.ReplaceAll(tmpl, []byte(id), []byte("<version>"))
		}
	}
	var decoded any
	if err := json.Unmarshal(tmpl, &decoded); err != nil {
		return nil, false
	}
	sortUnorderedLists(decoded)
	out, err := json.Marshal(decoded)
	if err != nil {
		return nil, false
	}
	return out, true
}

func sortUnorderedLists(node any) {
	switch v := node.(type) {
	case map[string]any:
		for key, child := range v {
			switch key {
			case "Tags":
				if tags, ok := child.([]any); ok {
					sort.SliceStable(tags, func(i, j int) bool { return tagKey(tags[i]) < tagKey(tags[j]) })
				}
			case "ParameterGroups":
				if groups, ok := child.([]any); ok {
					for _, group := range groups {
						if g, ok := group.(map[string]any); ok {
							if params, ok := g["Parameters"].([]any); ok {
								sort.SliceStable(params, func(i, j int) bool { return fmt.Sprint(params[i]) < fmt.Sprint(params[j]) })
							}
						}
					}
				}
			}
			sortUnorderedLists(child)
		}
	case []any:
		for _, child := range v {
			sortUnorderedLists(child)
		}
	}
}

func tagKey(tag any) string {
	if t, ok := tag.(map[string]any); ok {
		return fmt.Sprint(t["Key"])
	}
	return fmt.Sprint(tag)
}
