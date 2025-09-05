package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	generalsig "github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
	installsig "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

type AdminPromotionRequest struct {
	Tag string `json:"tag"`
}

// @ID						AdminPromotion
// @Summary				promotion callback.
// @Description.markdown	promotion.md
// @Param					req	body	AdminPromotionRequest	true	"Input"
// @Tags					general/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/general/promotion [POST]
func (s *service) AdminPromotion(ctx *gin.Context) {
	var req AdminPromotionRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(errors.Wrap(err, "unable to promote"))
		return
	}

	s.evClient.Send(ctx, "general", &generalsig.Signal{
		Type: generalsig.OperationRestart,
		Tag:  req.Tag,
	})
	s.evClient.Send(ctx, "general", &generalsig.Signal{
		Type: generalsig.OperationPromotion,
		Tag:  req.Tag,
	})

	// TODO: remove this when the install state initialization has already ran in the promotion workflow
	s.initializeInstallStates(ctx)
	ctx.JSON(http.StatusCreated, "ok")
}

func (s *service) initializeInstallStates(ctx *gin.Context) {
	batchSize := 20
	offset := 0
	hasMore := true

	for hasMore {
		var installs []app.Install

		res := s.db.
			Scopes(scopes.WithDisableViews).
			Model(&app.Install{}).
			Select("installs.id").
			Joins("LEFT JOIN install_states ON installs.id = install_states.install_id").
			Where("install_states.install_id IS NULL").
			Limit(batchSize).
			Offset(offset).
			Find(&installs)

		if res.Error != nil {
			ctx.Error(errors.Wrap(res.Error, "unable to get installs without state"))
			return
		}

		if len(installs) < batchSize {
			hasMore = false
		}

		for _, install := range installs {
			s.evClient.Send(ctx, install.ID, &installsig.Signal{
				Type: installsig.OperationGenerateState,
			})
		}
		offset += batchSize

	}
}
