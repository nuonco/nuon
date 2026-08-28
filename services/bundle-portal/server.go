package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

//go:embed client/dist/*
var portalAssets embed.FS

type portalServer struct {
	connected *connectedClient
	csrf      string
	hosts     map[string]bool
	assets    fs.FS
	index     []byte
	logger    *zap.Logger
	branding  portalBranding
}

func newPortalServer(csrf string, hosts map[string]bool, logger *zap.Logger) (*portalServer, error) {
	assets, err := fs.Sub(portalAssets, "client/dist")
	if err != nil {
		return nil, fmt.Errorf("open embedded client: %w", err)
	}
	index, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		return nil, fmt.Errorf("read embedded client index: %w", err)
	}
	index = []byte(strings.Replace(string(index), "{{CSRF_TOKEN}}", csrf, 1))
	return &portalServer{csrf: csrf, hosts: hosts, assets: assets, index: index, logger: logger, branding: defaultPortalBranding()}, nil
}

func (p *portalServer) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/branding", p.getBranding)
	mux.HandleFunc("GET /api/connected/releases", func(w http.ResponseWriter, r *http.Request) { p.connected.proxy(w, r, "/releases") })
	mux.HandleFunc("GET /api/connected/release-updates", func(w http.ResponseWriter, r *http.Request) { p.connected.proxy(w, r, "/release-updates") })
	mux.HandleFunc("GET /api/connected/releases/{release_id}", func(w http.ResponseWriter, r *http.Request) {
		p.connected.proxy(w, r, "/releases/"+url.PathEscape(r.PathValue("release_id")))
	})
	mux.HandleFunc("GET /api/connected/releases/{release_id}/files/content", func(w http.ResponseWriter, r *http.Request) {
		p.connected.proxy(w, r, "/releases/"+url.PathEscape(r.PathValue("release_id"))+"/files/content")
	})
	mux.HandleFunc("GET /api/connected/release-packages/{package_id}", func(w http.ResponseWriter, r *http.Request) {
		p.connected.proxy(w, r, "/release-packages/"+url.PathEscape(r.PathValue("package_id")))
	})
	mux.HandleFunc("GET /api/connected/workflows", func(w http.ResponseWriter, r *http.Request) { p.connected.proxy(w, r, "/workflows") })
	mux.HandleFunc("GET /api/connected/workflows/{workflow_id}", func(w http.ResponseWriter, r *http.Request) {
		p.connected.proxy(w, r, "/workflows/"+url.PathEscape(r.PathValue("workflow_id")))
	})
	mux.HandleFunc("GET /api/connected/workflows/{workflow_id}/steps/{step_id}/logs", func(w http.ResponseWriter, r *http.Request) {
		p.connected.proxy(w, r, "/workflows/"+url.PathEscape(r.PathValue("workflow_id"))+"/steps/"+url.PathEscape(r.PathValue("step_id"))+"/logs")
	})
	mux.HandleFunc("POST /api/connected/workflows/{workflow_id}/steps/{step_id}/retry", func(w http.ResponseWriter, r *http.Request) {
		p.connected.proxy(w, r, "/workflows/"+url.PathEscape(r.PathValue("workflow_id"))+"/steps/"+url.PathEscape(r.PathValue("step_id"))+"/retry")
	})
	mux.HandleFunc("GET /api/connected/workflows/{workflow_id}/steps/{step_id}/approvals/{approval_id}/contents", p.connectedApprovalContents)
	mux.HandleFunc("POST /api/connected/workflows/{workflow_id}/steps/{step_id}/approvals/{approval_id}/response", p.connectedApprovalResponse)
	mux.HandleFunc("GET /api/", http.NotFound)
	mux.HandleFunc("GET /", p.static)
	return p.middleware(mux)
}

func (p *portalServer) getBranding(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, p.branding)
}

func (p *portalServer) connectedApprovalContents(w http.ResponseWriter, r *http.Request) {
	p.connected.proxy(w, r, connectedApprovalPath(r)+"/contents")
}

func (p *portalServer) connectedApprovalResponse(w http.ResponseWriter, r *http.Request) {
	p.connected.proxy(w, r, connectedApprovalPath(r)+"/response")
}

func connectedApprovalPath(r *http.Request) string {
	return "/workflows/" + url.PathEscape(r.PathValue("workflow_id")) + "/steps/" + url.PathEscape(r.PathValue("step_id")) + "/approvals/" + url.PathEscape(r.PathValue("approval_id"))
}

func (p *portalServer) middleware(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		status := http.StatusOK
		writer := &statusWriter{ResponseWriter: w, status: &status}
		defer func() {
			p.logger.Info("http request", zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.Int("status", status), zap.Duration("duration", time.Since(started)))
		}()
		if !requestHostAllowed(p.hosts, r.Host) {
			writeAPIError(writer, fmt.Errorf("forbidden"), http.StatusForbidden)
			return
		}
		if r.Method != http.MethodGet && r.Header.Get("X-CSRF-Token") != p.csrf {
			writeAPIError(writer, fmt.Errorf("forbidden"), http.StatusForbidden)
			return
		}
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("X-Frame-Options", "DENY")
		writer.Header().Set("Referrer-Policy", "same-origin")
		writer.Header().Set("Content-Security-Policy", "default-src 'self'; connect-src 'self'; img-src 'self' data: https:; style-src 'self'; script-src 'self'; frame-ancestors 'none'")
		mux.ServeHTTP(writer, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status *int
}

func (w *statusWriter) WriteHeader(status int) {
	*w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (p *portalServer) static(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || !strings.Contains(strings.TrimPrefix(r.URL.Path, "/"), ".") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(p.index)
		return
	}
	http.FileServer(http.FS(p.assets)).ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeRawJSON(w http.ResponseWriter, raw []byte) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(raw)
}

func writeAPIError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
