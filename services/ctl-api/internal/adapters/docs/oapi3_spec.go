package docs

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
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
