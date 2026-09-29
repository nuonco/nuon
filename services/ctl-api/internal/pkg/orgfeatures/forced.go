package orgfeatures

import (
	"strings"
	"sync/atomic"
)

var forced atomic.Pointer[map[string]bool]

func SetForced(csv string) {
	set := make(map[string]bool)
	for _, name := range strings.Split(csv, ",") {
		if name = strings.TrimSpace(name); name != "" {
			set[name] = true
		}
	}
	forced.Store(&set)
}

func Forced() map[string]bool {
	set := forced.Load()
	if set == nil {
		return map[string]bool{}
	}
	return *set
}

func IsForced(name string) bool {
	return Forced()[name]
}
