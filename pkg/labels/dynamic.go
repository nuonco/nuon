package labels

import "strings"

// IsTemplatedValue reports whether a label value is a dynamic label template.
// It mirrors the render package's fast path: a value is only ever rendered
// when it references the .nuon namespace inside an interpolation expression,
// so anything else — including braces without .nuon — stays a literal.
func IsTemplatedValue(value string) bool {
	return strings.Contains(value, "{{") && strings.Contains(value, ".nuon")
}

// SplitTemplated partitions labels into static literal values and templated
// values that must be rendered against install state before use.
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
