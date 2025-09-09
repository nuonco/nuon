package migrations

import (
	"go.uber.org/fx"
)

type Params struct {
	fx.In
}

type Migrations struct{}

func New(params Params) *Migrations {
	return &Migrations{}
}
