package validate

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
)

func New(v *validator.Validate,
	writer metrics.Writer,
	l *zap.Logger,
) *workerInterceptor {
	return &workerInterceptor{
		v:  v,
		mw: writer,
		l:  l,
	}
}
