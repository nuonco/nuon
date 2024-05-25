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

func (d *Docs) loadOAPI2Spec() (*openapi2.T, error) {
	spec := docs.SwaggerInfo.ReadDoc()
	byts := []byte(spec)

	var doc openapi2.T
	err := json.Unmarshal(byts, &doc)
	if err != nil {
		return nil, fmt.Errorf("unable to convert open api spec to json: %w", err)
	}

	addSpecTags(&doc)
	return &doc, nil
}

func (d *Docs) getOAPI2PublicSpec(ctx *gin.Context) {
	doc, err := d.loadOAPI2Spec()
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, doc)
}

func (d *Docs) loadOAPI2AdminSpec() (*openapi2.T, error) {
	spec := admin.SwaggerInfoadmin.ReadDoc()
	byts := []byte(spec)

	var doc openapi2.T
	err := json.Unmarshal(byts, &doc)
	if err != nil {
		return nil, fmt.Errorf("unable to convert open api spec to json: %w", err)
	}
	addSpecTags(&doc)
	removeSecurity(&doc)

	return &doc, nil
}

func (d *Docs) getOAPI2AdminSpec(ctx *gin.Context) {
	doc, err := d.loadOAPI2AdminSpec()
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, doc)
}
