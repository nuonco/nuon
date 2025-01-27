package git

import (
	"path/filepath"
	"strings"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

const (
	gitSuffix string = ".git"
)

// https://powertoolsdev:token@github.com/powertoolsdev/mono.git
func Dir(src *plantypes.GitSource) string {
	if src == nil {
		return "."
	}

	url := src.URL

	url, _ = strings.CutSuffix(url, gitSuffix)
	return filepath.Base(url)
}
