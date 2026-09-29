package patcher

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/patcher"
)

type Params struct {
	fx.In
	L  *zap.Logger
	DB *gorm.DB `name:"psql"`
}

type middleware struct {
	l  *zap.Logger
	db *gorm.DB
}

func (m middleware) Name() string {
	return "patcher"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method != http.MethodPatch {
			ctx.Next()
			return
		}

		if ctx.Request.Body == nil || ctx.Request.ContentLength == 0 {
			m.l.Warn("Request body is nil")
			ctx.Next()
			return
		}

		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			m.l.Error("failed to read request body", zap.Error(err))
			ctx.Next()
			return
		}

		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var jsonData map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
			m.l.Error("failed to parse JSON", zap.Error(err))
			ctx.Next()
			return
		}

		properties := m.extractPropertiesRecursively(jsonData, "")

		cctx.SetPatcherGinCtx(ctx, patcher.Patcher{
			SelectFields: properties,
		})

		ctx.Next()
	}
}

func (m middleware) extractPropertiesRecursively(data map[string]interface{}, prefix string) []string {
	var properties []string

	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		properties = append(properties, fullKey)

		switch v := value.(type) {
		case map[string]interface{}:
			nestedProps := m.extractPropertiesRecursively(v, fullKey)
			properties = append(properties, nestedProps...)
		case []interface{}:
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					arrayPrefix := fullKey + "[" + strconv.Itoa(i) + "]"
					nestedProps := m.extractPropertiesRecursively(itemMap, arrayPrefix)
					properties = append(properties, nestedProps...)
				}
			}
		}
	}

	return properties
}

func New(params Params) *middleware {
	return &middleware{
		l:  params.L,
		db: params.DB,
	}
}
