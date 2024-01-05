package helpers

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"gorm.io/gorm"
)

type Helpers struct {
	cfg      *internal.Config
	ghClient *github.Client
	db       *gorm.DB
}

func New(v *validator.Validate,
	cfg *internal.Config,
	ghClient *github.Client,
	db *gorm.DB,
) *Helpers {
	return &Helpers{
		cfg:      cfg,
		ghClient: ghClient,
		db:       db,
	}
}
