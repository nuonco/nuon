package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

const APIProxyRoutePattern = "/v1/*path"

type APIProxyHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewAPIProxyHandler(cfg *internal.Config, l *zap.Logger) *APIProxyHandler {
	return &APIProxyHandler{cfg: cfg, l: l}
}

func (h *APIProxyHandler) RegisterRoutes(e *gin.Engine) error {
	target, err := url.Parse(h.cfg.ServerAPIURL())
	if err != nil {
		return err
	}
	public, err := url.Parse(h.cfg.APIUrl)
	if err != nil {
		return err
	}
	// Dial the in-cluster address, but keep the public API host on the request.
	// ctl-api copies Host into the config schema $id, which the browser fetches.
	host := target.Host
	if public.Host != "" {
		host = public.Host
	}

	proxy := &httputil.ReverseProxy{
		ModifyResponse: stripUpstreamCORS,
		Director: func(req *http.Request) {
			token, _ := req.Cookie(authCookie)

			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = host
			req.Header.Del("Cookie")
			req.Header.Del("Accept-Encoding")

			if token != nil && token.Value != "" {
				req.Header.Set("Authorization", "Bearer "+token.Value)
			}
		},
		ErrorLog: zap.NewStdLog(h.l),
	}

	e.Any(APIProxyRoutePattern, gin.WrapH(proxy))
	return nil
}

func stripUpstreamCORS(resp *http.Response) error {
	h := resp.Header
	h.Del("Access-Control-Allow-Origin")
	h.Del("Access-Control-Allow-Credentials")
	h.Del("Access-Control-Allow-Methods")
	h.Del("Access-Control-Allow-Headers")
	h.Del("Access-Control-Expose-Headers")
	h.Del("Access-Control-Max-Age")
	return nil
}
