package docs

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/gin-gonic/gin"
)

func (d *Docs) getOAPI3publicSpec(ctx *gin.Context) {
	doc, err := d.loadOAPI2Spec()
	if err != nil {
		ctx.Error(err)
		return
	}

	oapi3Doc, err := openapi2conv.ToV3(doc)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to convert open api spec to v3: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, oapi3Doc)
}

func (d *Docs) getOAPI3AdminSpec(ctx *gin.Context) {
	doc, err := d.loadOAPI2AdminSpec()
	if err != nil {
		ctx.Error(err)
		return
	}

	oapi3Doc, err := openapi2conv.ToV3(doc)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to convert open api spec to v3: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, oapi3Doc)
}
