package dir

import (
	"strings"
)

func (p *parser) hasExtension(name string) bool {
	return strings.HasSuffix(name, p.opts.Ext)
}

func (p *parser) ensureExtension(name string) string {
	if p.hasExtension(name) {
		return name
	}

	return name + p.opts.Ext
}
