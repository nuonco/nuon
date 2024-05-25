package loops

import (
	"github.com/go-playground/validator/v10"
	loopsclient "github.com/powertoolsdev/mono/pkg/loops"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

func New(v *validator.Validate, cfg *internal.Config) (loopsclient.Client, error) {
	return loopsclient.New(v, loopsclient.WithAPIKey(cfg.LoopsAPIKey))
}
