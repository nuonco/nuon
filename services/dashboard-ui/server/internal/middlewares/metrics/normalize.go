package metrics

import (
	"regexp"
	"strings"

	"github.com/nuonco/nuon/pkg/shortid"
)

var (
	uuidRe   = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	digitsRe = regexp.MustCompile(`^[0-9]+$`)
)

func normalizeAPIPath(p string) string {
	trimmed := strings.Trim(p, "/")
	if trimmed == "" {
		return "/"
	}

	segs := strings.Split(trimmed, "/")
	for i, s := range segs {
		if prefix, err := shortid.GetPrefix(s); err == nil {
			segs[i] = "{" + prefix + "_id}"
			continue
		}
		switch {
		case uuidRe.MatchString(s):
			segs[i] = "{uuid}"
		case digitsRe.MatchString(s):
			segs[i] = "{id}"
		}
	}

	return "/" + strings.Join(segs, "/")
}
