// Package helpers is the cross-domain entry point for working with composite
// errors: recording, hydrating, listing, building cause trees, and resolving.
//
// Mirrors the helpers pattern documented in services/ctl-api/AGENTS.md
// (FX-injected struct with config, DB, validator, logger). Other domains
// access composite errors only through this package.
package helpers

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
	"github.com/nuonco/nuon/pkg/composite_error/catalog"
	"github.com/nuonco/nuon/pkg/composite_error/unknown"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg *internal.Config
	DB  *gorm.DB `name:"psql"`
	L   *zap.Logger
	V   *validator.Validate
}

type Helpers struct {
	cfg      *internal.Config
	db       *gorm.DB
	l        *zap.Logger
	v        *validator.Validate
	pipeline *composite_error.Pipeline
}

func New(params Params) *Helpers {
	l := params.L.With(zap.String("subsystem", "composite_errors"))

	pipeline := composite_error.NewPipeline(
		catalog.ParsersForContext,
		unknown.Build,
		func(name string, val any) {
			l.Warn("composite error parser panicked",
				zap.String("parser", name),
				zap.Any("panic", val),
			)
		},
	)

	return &Helpers{
		cfg:      params.Cfg,
		db:       params.DB,
		l:        l,
		v:        params.V,
		pipeline: pipeline,
	}
}
