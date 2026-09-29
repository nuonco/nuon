package labels

import "strings"

func IsTemplatedValue(value string) bool {
	return strings.Contains(value, "{{") && strings.Contains(value, ".nuon")
}

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
