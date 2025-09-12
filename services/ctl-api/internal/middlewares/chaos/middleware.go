package chaos

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Params struct {
	fx.In
	L   *zap.Logger
	Cfg *internal.Config
}

type middleware struct {
	l             *zap.Logger
	cfg           *internal.Config
	errors        map[string]func(*gin.Context, *zap.Logger)
	allowedRoutes map[string]struct{}
}

func (m middleware) Name() string {
	return "chaos"
}

var skipRoutes = map[string]struct{}{
	"/livez":   {},
	"/version": {},
	"/docs/":   {},
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if m.cfg.ChaosErrorRate < 1 {
			ctx.Next()
		}

		if _, ok := skipRoutes[ctx.FullPath()]; ok {
			ctx.Next()
			return
		}

		for route := range skipRoutes {
			if strings.HasPrefix(ctx.FullPath(), route) {
				ctx.Next()
				return
			}
		}

		if len(m.allowedRoutes) > 0 {
			route := strings.ToLower(ctx.Request.Method) + "_" + ctx.FullPath()
			if _, ok := m.allowedRoutes[route]; !ok {
				// route not in allowed list
				ctx.Next()
				return
			}
		}

		if rand.IntN(m.cfg.ChaosErrorRate)+1 != 1 {
			ctx.Next()
			return
		}

		allowedErrors := m.GetAllowedErrors()

		// Chaos triggered! Select random error type
		errorTypes := make([]string, 0, len(allowedErrors))
		for errorType := range allowedErrors {
			errorTypes = append(errorTypes, errorType)
		}

		selectedError := errorTypes[rand.IntN(len(errorTypes))]
		fmt.Printf("Chaos thanos triggered: %s\n", selectedError)

		// Execute the selected error
		m.errors[selectedError](ctx, m.l)
	}
}

func (m *middleware) GetAllowedErrors() map[string]func(*gin.Context, *zap.Logger) {
	// get a new map from errors where only allowed errors are present
	if len(m.cfg.ChaosErrors) == 0 {
		return m.errors
	}

	newErrors := make(map[string]func(*gin.Context, *zap.Logger))
	for _, err := range m.cfg.ChaosErrors {
		if fn, ok := m.errors[err]; ok {
			newErrors[err] = fn
		}
	}
	return newErrors
}

func New(params Params) *middleware {
	errors := map[string]func(*gin.Context, *zap.Logger){
		"panic":                 panicHandler,
		"internal_server_error": internalServerErrorHandler,
		"service_unavailable":   serviceUnavailableHandler,
		"bad_gateway":           badGatewayHandler,
		"timeout":               timeoutHandler,
	}

	var allowedRoutes map[string]struct{}
	if len(params.Cfg.ChaosRoutes) > 0 {
		allowedRoutes = make(map[string]struct{}, len(params.Cfg.ChaosRoutes))
		for _, route := range params.Cfg.ChaosRoutes {
			allowedRoutes[route] = struct{}{}
		}
	}

	return &middleware{
		l:             params.L,
		cfg:           params.Cfg,
		errors:        errors,
		allowedRoutes: allowedRoutes,
	}
}

