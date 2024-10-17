package log

import (
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/bins/runner/internal"
)

func NewDev(cfg *internal.Config) (*zap.Logger, error) {
	dev, err := zap.NewDevelopment()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get zap development")
	}

	return dev, nil
}
