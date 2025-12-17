package loops

import (
	"github.com/go-playground/validator/v10"
	loopsclient "github.com/nuonco/nuon/pkg/loops"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func New(v *validator.Validate, cfg *internal.Config) (loopsclient.Client, error) {
	return loopsclient.New(v, loopsclient.WithAPIKey(cfg.LoopsAPIKey))
}
