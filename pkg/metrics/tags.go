package metrics

import (
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/pkg/generics"
)

// ToTags is a flexible tag creator, which can accept a map (for default tags), and either partial tags or full tags.
//
// For instance, both the following string parameter sets are equivalent:
// ToTags(defaultTags, "status", "ok", "step:step-2")
// ToTags(defaultTags, "status", "ok", "step", "step-2")
// ToTags(defaultTags, "status:ok", "step:step-2")
func ToTags(inputs map[string]string, addtlTags ...string) []string {
	tags := make([]string, 0)
	for k, v := range inputs {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}

	partialTags := make([]string, 0, len(addtlTags))
	for _, tag := range addtlTags {
		if strings.Contains(tag, ":") {
			tags = append(tags, tag)
			continue
		}
		partialTags = append(partialTags, tag)
	}

	kvs := generics.SliceToGroups(partialTags, 2)
	for _, kv := range kvs {
		if len(kv) < 2 {
			continue
		}

		tags = append(tags, strings.Join(kv, ":"))
	}

	return tags
}
