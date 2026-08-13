package labels

import "strings"

// IsTemplatedValue reports whether a label value is a dynamic label template.
// Mirrors the render package's fast path: only values referencing .nuon inside
// braces are ever rendered; anything else stays a literal.
func IsTemplatedValue(value string) bool {
	return strings.Contains(value, "{{") && strings.Contains(value, ".nuon")
}

// SplitTemplated partitions labels into static literals and templated values.
func (l Labels) SplitTemplated() (static Labels, templated Labels) {
	static = make(Labels)
	templated = make(Labels)
	for k, v := range l {
		if IsTemplatedValue(v) {
			templated[k] = v
		} else {
			static[k] = v
		}
	}
	return static, templated
}
