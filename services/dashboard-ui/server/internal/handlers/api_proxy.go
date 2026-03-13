package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type APIProxyHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewAPIProxyHandler(cfg *internal.Config, l *zap.Logger) *APIProxyHandler {
	return &APIProxyHandler{cfg: cfg, l: l}
}

func (h *APIProxyHandler) RegisterRoutes(e *gin.Engine) error {
	target, err := url.Parse(h.cfg.APIUrl)
	if err != nil {
		return err
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			token, _ := req.Cookie(authCookie)

			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host
			req.Header.Del("Cookie")
			req.Header.Del("Accept-Encoding")

			if token != nil && token.Value != "" {
				req.Header.Set("Authorization", "Bearer "+token.Value)
			}
		},
		ErrorLog: zap.NewStdLog(h.l),
	}

	e.Any("/v1/*path", gin.WrapH(proxy))
	return nil
}
