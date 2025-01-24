package git

import (
	"path/filepath"
	"strings"
)

const (
	gitSuffix string = ".git"
)

// https://powertoolsdev:token@github.com/powertoolsdev/mono.git
func Dir(src *Source) string {
	url := src.URL

	url, _ = strings.CutSuffix(url, gitSuffix)
	return filepath.Base(url)
}
