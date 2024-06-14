package authz

import (
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Client struct {
	cfg      *internal.Config
	db       *gorm.DB
	v        *validator.Validate
	evClient eventloop.Client
}

func New(v *validator.Validate,
	cfg *internal.Config,
	db *gorm.DB,
	evClient eventloop.Client,
) *Client {
	return &Client{
		v:        v,
		cfg:      cfg,
		db:       db,
		evClient: evClient,
	}
}
