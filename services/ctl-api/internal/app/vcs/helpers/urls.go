package helpers

import (
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// SplitRepoSlug splits a url that is of the format owner/name into proper values.
func (h *Helpers) SplitRepoSlug(val string) (string, string, error) {
	pieces := strings.SplitN(val, "/", 2)
	if len(pieces) != 2 {
		return "", "", stderr.ErrUser{
			Err:         fmt.Errorf("invalid repo, must be of the format <user-name>/<repo-name>"),
			Description: "please correct format and try again",
		}
	}

	return pieces[0], pieces[1], nil
}
