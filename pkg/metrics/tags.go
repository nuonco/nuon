package metrics

import (
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/pkg/generics"
)

// ToTags is used to accept a "default" tag mapping, and then appened additonal tags into it, and then return it as an
// array of key:value pairs
func ToTags(inputs map[string]string, addtlTags ...string) []string {
	tags := make([]string, 0)
	for k, v := range inputs {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}

	kvs := generics.SliceToGroups(addtlTags, 2)
	for _, kv := range kvs {
		if len(kv) < 2 {
			continue
		}

		tags = append(tags, strings.Join(kv, ":"))
	}

	return tags
}
