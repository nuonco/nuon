package docs

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/admin"
	"github.com/powertoolsdev/mono/services/ctl-api/docs"
)

func (d *Docs) getOAPI2PublicSpec(ctx *gin.Context) {
	spec := docs.SwaggerInfo.ReadDoc()
	byts := []byte(spec)

	var doc openapi2.T
	err := json.Unmarshal(byts, &doc)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to convert open api spec to json: %w", err))
		return
	}
	doc.Info.Version = d.cfg.GitRef

	ctx.JSON(http.StatusOK, doc)
}

func (d *Docs) getOAPI2AdminSpec(ctx *gin.Context) {
	spec := admin.SwaggerInfoadmin.ReadDoc()
	byts := []byte(spec)

	var doc openapi2.T
	err := json.Unmarshal(byts, &doc)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to convert open api spec to json: %w", err))
		return
	}
	doc.Info.Version = d.cfg.GitRef

	ctx.JSON(http.StatusOK, doc)
}
