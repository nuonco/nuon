package hooks

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type Hooks struct {
	l *zap.Logger
}

func New(v *validator.Validate, l *zap.Logger) *Hooks {
	return &Hooks{
		l: l,
	}
}
