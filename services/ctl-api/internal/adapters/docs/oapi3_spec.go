package docs

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/admin"
	"github.com/powertoolsdev/mono/services/ctl-api/docs"
)

// addSpecTags will add tags for each operation into the top level, general section
func addSpecTags(doc *openapi2.T) {
	allTags := make(map[string]struct{}, 0)
	for _, pathItem := range doc.Paths {
		for _, op := range pathItem.Operations() {
			for _, tag := range op.Tags {
				allTags[tag] = struct{}{}
			}
		}
	}

	doc.Tags = make([]*openapi3.Tag, 0, len(allTags))
	for tag := range allTags {
		doc.Tags = append(doc.Tags, &openapi3.Tag{
			Name:        tag,
			Description: tag,
		})
	}
}

func (d *Docs) getOAPI3publicSpec(ctx *gin.Context) {
	spec := docs.SwaggerInfo.ReadDoc()
	byts := []byte(spec)

	var doc openapi2.T
	err := json.Unmarshal(byts, &doc)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to convert open api spec to json: %w", err))
		return
	}
	doc.Info.Version = d.cfg.GitRef
	addSpecTags(&doc)

	oapi3Doc, err := openapi2conv.ToV3(&doc)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to convert open api spec to v3: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, oapi3Doc)
}

func (d *Docs) getOAPI3AdminSpec(ctx *gin.Context) {
	spec := admin.SwaggerInfoadmin.ReadDoc()
	byts := []byte(spec)

	var doc openapi2.T
	err := json.Unmarshal(byts, &doc)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to convert open api spec to json: %w", err))
		return
	}

	doc.Info.Version = d.cfg.GitRef
	addSpecTags(&doc)

	oapi3Doc, err := openapi2conv.ToV3(&doc)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to convert open api spec to v3: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, oapi3Doc)
}
