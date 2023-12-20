package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

const (
	defaultSpecFP string = "docs/swagger.json"
)

func (c *service) GetOpenAPI3Spec(ctx *gin.Context) {
	byts, err := os.ReadFile(defaultSpecFP)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to read spec: %w", err))
		return
	}

	var doc openapi2.T
	err = json.Unmarshal(byts, &doc)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to convert open api spec to json: %w", err))
		return
	}

	oapi3Doc, err := openapi2conv.ToV3(&doc)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to convert open api spec to v3: %w", err))
		return
	}
	oapi3Doc.Servers = openapi3.Servers{
		{
			URL: "http://localhost:8081",
		},
	}

	ctx.JSON(http.StatusOK, oapi3Doc)
}
