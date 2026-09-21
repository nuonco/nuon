package syncerr

import (
	"errors"
	"fmt"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func From(resource, description string, err error) error {
	var userErr stderr.ErrUser
	var invalidErr stderr.ErrInvalidRequest
	if errors.As(err, &userErr) || errors.As(err, &invalidErr) {
		detail := userErr.Description
		if detail == "" {
			detail = err.Error()
		}
		return sync.SyncErr{
			Resource:    resource,
			Description: fmt.Sprintf("%s: %s", description, detail),
			Err:         err,
		}
	}
	return sync.SyncInternalErr{Description: description, Err: err}
}
