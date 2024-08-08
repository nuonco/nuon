package activities

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

//go:generate go run ../../../../../../../pkg/bins/temporal-gen

type Activities struct {
	db       *gorm.DB
	protos   *protos.Adapter
	helpers  *helpers.Helpers
	evClient eventloop.Client
}

func New(prt *protos.Adapter,
	db *gorm.DB,
	helpers *helpers.Helpers,
	evClient eventloop.Client,
) (*Activities, error) {
	return &Activities{
		db:       db,
		protos:   prt,
		helpers:  helpers,
		evClient: evClient,
	}, nil
}
