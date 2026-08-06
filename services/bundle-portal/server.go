package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2state"
)

//go:embed client/dist/*
var portalAssets embed.FS

type portalServer struct {
	store       day2state.State
	csrf        string
	requestedBy string
	hosts       map[string]bool
	assets      fs.FS
	index       []byte
	logger      *zap.Logger
}

func newPortalServer(store day2state.State, csrf, requestedBy string, hosts map[string]bool, logger *zap.Logger) (*portalServer, error) {
	assets, err := fs.Sub(portalAssets, "client/dist")
	if err != nil {
		return nil, fmt.Errorf("open embedded client: %w", err)
	}
	index, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		return nil, fmt.Errorf("read embedded client index: %w", err)
	}
	index = []byte(strings.Replace(string(index), "{{CSRF_TOKEN}}", csrf, 1))
	return &portalServer{store: store, csrf: csrf, requestedBy: requestedBy, hosts: hosts, assets: assets, index: index, logger: logger}, nil
}

func (p *portalServer) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/catalog", p.catalog)
	mux.HandleFunc("GET /api/status", p.object("status.json"))
	mux.HandleFunc("GET /api/health", p.health)
	mux.HandleFunc("GET /api/runs", p.runs)
	mux.HandleFunc("GET /api/runs/{id}", p.run)
	mux.HandleFunc("POST /api/dispatch", p.dispatch)
	mux.HandleFunc("GET /api/dispatches", p.dispatches)
	mux.HandleFunc("GET /", p.static)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		status := http.StatusOK
		writer := &statusWriter{ResponseWriter: w, status: &status}
		defer func() {
			p.logger.Info("http request", zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.Int("status", status), zap.Duration("duration", time.Since(started)))
		}()
		if !p.hosts[requestHost(r.Host)] {
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
		writer.Header().Set("Content-Security-Policy", "default-src 'self'; connect-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; frame-ancestors 'none'")
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

func (p *portalServer) object(key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, ok, err := p.store.Get(r.Context(), key)
		if err != nil {
			writeAPIError(w, err, http.StatusInternalServerError)
			return
		}
		if !ok {
			writeAPIError(w, fmt.Errorf("not found"), http.StatusNotFound)
			return
		}
		writeRawJSON(w, raw)
	}
}

func (p *portalServer) catalog(w http.ResponseWriter, r *http.Request) {
	p.object(day2.CatalogKey)(w, r)
}

func (p *portalServer) health(w http.ResponseWriter, r *http.Request) {
	latest, _, err := p.store.Get(r.Context(), "health/latest.json")
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	transitions, ok, err := p.store.Get(r.Context(), "health/transitions.json")
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if !ok {
		transitions = []byte("[]")
	}
	if len(latest) == 0 {
		latest = []byte("null")
	}
	writeRawJSON(w, []byte(fmt.Sprintf(`{"latest":%s,"transitions":%s}`, latest, transitions)))
}

func (p *portalServer) runs(w http.ResponseWriter, r *http.Request) {
	runs, err := p.listRuns(r)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, runs)
}

func (p *portalServer) run(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := day2.ValidateDispatchID(id); err != nil {
		writeAPIError(w, err, http.StatusBadRequest)
		return
	}
	run, err := p.readRun(r, id)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if run == nil {
		writeAPIError(w, fmt.Errorf("run not found"), http.StatusNotFound)
		return
	}
	writeJSON(w, run)
}

func (p *portalServer) readRun(r *http.Request, id string) (*day2.RunStatus, error) {
	raw, ok, err := p.store.Get(r.Context(), day2.RunStatusKey(id))
	if err != nil || !ok {
		return nil, err
	}
	var run day2.RunStatus
	if err := json.Unmarshal(raw, &run); err != nil {
		return nil, fmt.Errorf("decode run %s: %w", id, err)
	}
	return &run, nil
}

func (p *portalServer) listRuns(r *http.Request) ([]day2.RunStatus, error) {
	keys, err := p.store.List(r.Context(), day2.RunsPrefix)
	if err != nil {
		return nil, err
	}
	runs := make([]day2.RunStatus, 0)
	for _, key := range keys {
		if !strings.HasSuffix(key, "/status.json") {
			continue
		}
		raw, ok, err := p.store.Get(r.Context(), key)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		var run day2.RunStatus
		if err := json.Unmarshal(raw, &run); err != nil {
			return nil, fmt.Errorf("decode %s: %w", key, err)
		}
		runs = append(runs, run)
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].StartedAt.After(runs[j].StartedAt) })
	return runs, nil
}

func (p *portalServer) dispatch(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RefID string `json:"ref_id"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeAPIError(w, err, http.StatusBadRequest)
		return
	}
	catalog, err := p.readCatalog(r)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if !catalogHasRef(catalog, input.RefID) {
		writeAPIError(w, fmt.Errorf("unknown ref %q", input.RefID), http.StatusBadRequest)
		return
	}
	id := "portal-" + strings.ToLower(ulid.Make().String())
	request := day2.Request{
		SchemaVersion: day2.SchemaVersion,
		DeploymentID:  catalog.DeploymentID,
		BundleDigest:  catalog.BundleDigest,
		RefID:         input.RefID,
		DispatchID:    id,
		Source:        day2.SourcePortal,
		RequestedBy:   p.requestedBy,
		CreatedAt:     time.Now().UTC(),
	}
	raw, err := json.Marshal(request)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if err := p.store.PutIfAbsent(r.Context(), day2.RequestKey(id), raw); err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	p.logger.Info("dispatch requested", zap.String("dispatch_id", id), zap.String("ref_id", input.RefID), zap.String("requested_by", p.requestedBy))
	writeJSON(w, map[string]string{"dispatch_id": id})
}

func (p *portalServer) readCatalog(r *http.Request) (*day2.Catalog, error) {
	raw, ok, err := p.store.Get(r.Context(), day2.CatalogKey)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("no day-2 catalog found")
	}
	var catalog day2.Catalog
	if err := json.Unmarshal(raw, &catalog); err != nil {
		return nil, fmt.Errorf("decode %s: %w", day2.CatalogKey, err)
	}
	return &catalog, nil
}

func catalogHasRef(catalog *day2.Catalog, id string) bool {
	for _, ref := range catalog.Refs {
		if ref.ID == id {
			return true
		}
	}
	return false
}

func (p *portalServer) dispatches(w http.ResponseWriter, r *http.Request) {
	keys, err := p.store.List(r.Context(), "dispatch/")
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	receipts := make(map[string]bool)
	for _, key := range keys {
		if strings.HasPrefix(key, day2.ReceiptsPrefix) {
			receipts[strings.TrimSuffix(strings.TrimPrefix(key, day2.ReceiptsPrefix), ".json")] = true
		}
	}
	items := make([]json.RawMessage, 0)
	for _, key := range keys {
		if !strings.HasPrefix(key, day2.RequestsPrefix) && !strings.HasPrefix(key, day2.ReceiptsPrefix) {
			continue
		}
		if strings.HasPrefix(key, day2.RequestsPrefix) && receipts[strings.TrimSuffix(strings.TrimPrefix(key, day2.RequestsPrefix), ".json")] {
			continue
		}
		raw, ok, err := p.store.Get(r.Context(), key)
		if err != nil {
			writeAPIError(w, err, http.StatusInternalServerError)
			return
		}
		if ok {
			items = append(items, raw)
		}
	}
	writeJSON(w, items)
}

func writeJSON(w http.ResponseWriter, value any) {
	raw, err := json.Marshal(value)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	writeRawJSON(w, raw)
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
