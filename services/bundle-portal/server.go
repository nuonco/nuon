package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/bundleupgrade"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operationstate"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
)

//go:embed client/dist/*
var portalAssets embed.FS

type portalServer struct {
	mode                    string
	connected               *connectedClient
	store                   operationstate.State
	controlStore            operationstate.State
	stackStore              operationstate.State
	csrf                    string
	requestedBy             string
	hosts                   map[string]bool
	assets                  fs.FS
	index                   []byte
	logger                  *zap.Logger
	installStackName        string
	installStackReader      installStackReader
	deploymentID            string
	cloudProvider           string
	cloudAccountID          string
	cloudRegion             string
	stackPlanner            stackPlanAPI
	stackPlanMu             sync.Mutex
	bundleUploadMu          sync.Mutex
	bundleUploadStatusMu    sync.RWMutex
	bundleUploadStatus      bundleUploadStatus
	stageBundle             func(context.Context, string, string, func(bundleupgrade.Progress)) (*bundleupgrade.Result, error)
	bundleActionDefinitions bundleActionDefinitions
	branding                portalBranding
}

type bundleUploadStatus struct {
	State     string    `json:"state"`
	Phase     string    `json:"phase"`
	Detail    string    `json:"detail"`
	UpdatedAt time.Time `json:"updated_at"`
}

func newPortalServer(store, stackStore operationstate.State, csrf, requestedBy string, hosts map[string]bool, logger *zap.Logger) (*portalServer, error) {
	assets, err := fs.Sub(portalAssets, "client/dist")
	if err != nil {
		return nil, fmt.Errorf("open embedded client: %w", err)
	}
	index, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		return nil, fmt.Errorf("read embedded client index: %w", err)
	}
	index = []byte(strings.Replace(string(index), "{{CSRF_TOKEN}}", csrf, 1))
	if stackStore == nil {
		stackStore = store
	}
	return &portalServer{store: store, controlStore: store, stackStore: stackStore, csrf: csrf, requestedBy: requestedBy, hosts: hosts, assets: assets, index: index, logger: logger, branding: defaultPortalBranding()}, nil
}

func (p *portalServer) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/branding", p.getBranding)
	mux.HandleFunc("GET /api/mode/capabilities", p.modeCapabilities)
	if p.connected != nil {
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
	mux.HandleFunc("GET /api/catalog", p.catalog)
	mux.HandleFunc("GET /api/bundle", p.bundle)
	mux.HandleFunc("POST /api/bundle-candidate", p.uploadBundleCandidate)
	mux.HandleFunc("GET /api/bundle-candidate/upload-status", p.getBundleUploadStatus)
	mux.HandleFunc("POST /api/bundle-candidate/clear", p.clearBundleCandidate)
	mux.HandleFunc("POST /api/bundle-candidate/plan-stack", p.planBundleCandidateStack)
	mux.HandleFunc("POST /api/bundle-candidate/approve", p.approveBundleCandidate)
	mux.HandleFunc("GET /api/status", p.object("status.json"))
	mux.HandleFunc("GET /api/report", p.object("report.json"))
	mux.HandleFunc("GET /api/stack-outputs", p.stackOutputs)
	mux.HandleFunc("GET /api/install-stack", p.installStack)
	mux.HandleFunc("GET /api/installation-registration", p.installationRegistration)
	mux.HandleFunc("GET /api/support-snapshot", p.supportSnapshot)
	mux.HandleFunc("GET /api/health", p.health)
	mux.HandleFunc("GET /api/runner-heartbeat", p.runnerHeartbeat)
	mux.HandleFunc("GET /api/runs", p.runs)
	mux.HandleFunc("GET /api/runs/{id}", p.run)
	mux.HandleFunc("POST /api/runs/{id}/control", p.runControl)
	mux.HandleFunc("GET /api/logs", p.logs)
	mux.HandleFunc("GET /api/logs/{id}", p.log)
	mux.HandleFunc("GET /api/plans/{id}", p.plan)
	mux.HandleFunc("GET /api/step-plans/{id}", p.stepPlan)
	mux.HandleFunc("GET /api/step-results/{id}", p.stepResult)
	mux.HandleFunc("POST /api/dispatch", p.dispatch)
	mux.HandleFunc("GET /api/dispatches", p.dispatches)
	mux.HandleFunc("GET /", p.static)
	return p.middleware(mux)
}

func (p *portalServer) getBranding(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, p.branding)
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

func (p *portalServer) modeCapabilities(w http.ResponseWriter, _ *http.Request) {
	mode := p.mode
	if mode == "" {
		mode = "offline"
	}
	writeJSON(w, map[string]any{
		"mode": mode,
		"capabilities": map[string]bool{
			"vendor_releases": mode == "connected", "workflows": mode == "connected", "approvals": mode == "connected",
			"bundle_upload": mode == "offline", "support_snapshot": mode == "offline", "local_dispatch": mode == "offline",
		},
	})
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

func (p *portalServer) uploadBundleCandidate(w http.ResponseWriter, r *http.Request) {
	if p.stageBundle == nil {
		writeAPIError(w, fmt.Errorf("bundle upload is unavailable for this portal deployment"), http.StatusConflict)
		return
	}
	if !p.bundleUploadMu.TryLock() {
		writeAPIError(w, fmt.Errorf("another bundle upload is already being processed"), http.StatusConflict)
		return
	}
	defer p.bundleUploadMu.Unlock()
	p.setBundleUploadStatus("uploading", "receiving", "Receiving the bundle archive")
	tmp, err := os.CreateTemp("", "nuon-bundle-upload-*.tar.zst")
	if err != nil {
		p.setBundleUploadStatus("failed", "failed", err.Error())
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	path := tmp.Name()
	defer os.Remove(path)
	const maxBundleUpload = int64(5 << 30)
	reader := http.MaxBytesReader(w, r.Body, maxBundleUpload)
	if _, err := io.Copy(tmp, reader); err != nil {
		tmp.Close()
		p.setBundleUploadStatus("failed", "failed", err.Error())
		status := http.StatusInternalServerError
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			status = http.StatusRequestEntityTooLarge
		}
		writeAPIError(w, fmt.Errorf("receive bundle archive: %w", err), status)
		return
	}
	if err := tmp.Close(); err != nil {
		p.setBundleUploadStatus("failed", "failed", err.Error())
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	archiveName, err := url.PathUnescape(r.Header.Get("X-Nuon-Bundle-Filename"))
	if err != nil {
		p.setBundleUploadStatus("failed", "failed", err.Error())
		writeAPIError(w, fmt.Errorf("decode bundle filename: %w", err), http.StatusBadRequest)
		return
	}
	result, err := p.stageBundle(r.Context(), path, archiveName, func(progress bundleupgrade.Progress) {
		p.setBundleUploadStatus("processing", progress.Phase, progress.Detail)
	})
	if err != nil {
		p.setBundleUploadStatus("failed", "failed", err.Error())
		writeAPIError(w, err, http.StatusUnprocessableEntity)
		return
	}
	p.setBundleUploadStatus("complete", "complete", "Bundle staged and ready for review")
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, result.Candidate)
}

func (p *portalServer) getBundleUploadStatus(w http.ResponseWriter, _ *http.Request) {
	p.bundleUploadStatusMu.RLock()
	status := p.bundleUploadStatus
	p.bundleUploadStatusMu.RUnlock()
	writeJSON(w, status)
}

func (p *portalServer) setBundleUploadStatus(state, phase, detail string) {
	p.bundleUploadStatusMu.Lock()
	p.bundleUploadStatus = bundleUploadStatus{State: state, Phase: phase, Detail: detail, UpdatedAt: time.Now().UTC()}
	p.bundleUploadStatusMu.Unlock()
}

func (p *portalServer) clearBundleCandidate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		BundleDigest string `json:"bundle_digest"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeAPIError(w, fmt.Errorf("decode request: %w", err), http.StatusBadRequest)
		return
	}
	p.bundleUploadMu.Lock()
	defer p.bundleUploadMu.Unlock()
	candidate, _, found, err := p.latestBundleCandidateRecord(r.Context())
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if !found {
		writeAPIError(w, fmt.Errorf("bundle candidate not found"), http.StatusNotFound)
		return
	}
	if input.BundleDigest == "" || input.BundleDigest != candidate.Bundle.BundleDigest {
		writeAPIError(w, fmt.Errorf("candidate changed; reload before clearing"), http.StatusConflict)
		return
	}
	dismissedAt := time.Now().UTC()
	dismissal := operation.BundleCandidateDismissal{
		SchemaVersion: operation.SchemaVersion,
		BundleDigest:  input.BundleDigest,
		DismissedAt:   dismissedAt,
		RequestedBy:   p.requestedBy,
	}
	raw, err := json.Marshal(dismissal)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if err := p.controlStore.PutIfAbsent(r.Context(), operation.CandidateDismissalKey(input.BundleDigest, dismissedAt), raw); err != nil {
		writeAPIError(w, fmt.Errorf("persist candidate dismissal: %w", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"bundle_digest": input.BundleDigest, "status": "cleared"})
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
	p.object(operation.CatalogKey)(w, r)
}

// stackOutputs reads the phone-home Lambda's output document, which the stack
// bootstrap writes as a sibling of the runner state prefix.
func (p *portalServer) stackOutputs(w http.ResponseWriter, r *http.Request) {
	raw, ok, err := p.stackStore.Get(r.Context(), "stack-outputs/outputs.json")
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

type bundleHistoryComparison struct {
	PreviousDigest string                   `json:"previous_digest"`
	BundleDigest   string                   `json:"bundle_digest"`
	Available      bool                     `json:"available"`
	Changes        []operation.BundleChange `json:"changes"`
}

func (p *portalServer) bundle(w http.ResponseWriter, r *http.Request) {
	active, ok, err := p.store.Get(r.Context(), operation.BundleKey)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if !ok {
		active = []byte("null")
	}
	keys, err := p.store.List(r.Context(), operation.BundlesPrefix)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	history := make([]operation.BundleInfo, 0)
	for _, key := range keys {
		if !strings.HasSuffix(key, ".json") {
			continue
		}
		raw, ok, err := p.store.Get(r.Context(), key)
		if err != nil {
			writeAPIError(w, err, http.StatusInternalServerError)
			return
		}
		if !ok {
			continue
		}
		var info operation.BundleInfo
		if err := json.Unmarshal(raw, &info); err != nil {
			writeAPIError(w, fmt.Errorf("decode %s: %w", key, err), http.StatusInternalServerError)
			return
		}
		history = append(history, info)
	}
	sort.Slice(history, func(i, j int) bool { return history[i].ActivatedAt.After(history[j].ActivatedAt) })
	addHistoricalActionDefinitions(history, p.bundleActionDefinitions)
	comparisons := make([]bundleHistoryComparison, 0)
	if len(history) > 1 {
		comparisons = make([]bundleHistoryComparison, 0, len(history)-1)
	}
	for i := 0; i+1 < len(history); i++ {
		comparison := bundleHistoryComparison{
			PreviousDigest: history[i+1].BundleDigest,
			BundleDigest:   history[i].BundleDigest,
			Available:      history[i+1].Contents != nil && history[i].Contents != nil,
		}
		if comparison.Available {
			comparison.Changes = operation.CompareBundleContents(history[i+1], history[i])
		}
		comparisons = append(comparisons, comparison)
	}
	candidateInfo, candidateRecordKey, candidateFound, err := p.latestBundleCandidateRecord(r.Context())
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	candidate := []byte("null")
	if candidateFound {
		candidate, err = json.Marshal(candidateInfo)
		if err != nil {
			writeAPIError(w, err, http.StatusInternalServerError)
			return
		}
	}
	stackCandidate, stackCandidateFound, err := p.store.Get(r.Context(), operation.StackCandidateKey)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if !stackCandidateFound {
		stackCandidate = []byte("null")
	} else {
		var decoded operation.StackCandidate
		if err := json.Unmarshal(stackCandidate, &decoded); err != nil {
			writeAPIError(w, fmt.Errorf("decode stack candidate: %w", err), http.StatusInternalServerError)
			return
		}
		if !stackPropertyChangesCaptured(decoded) {
			deployedTemplate, deployedFound, deployedErr := p.stackStore.Get(r.Context(), "stack/root-template.json")
			candidateTemplate, candidateTemplateFound, candidateErr := p.stackStore.Get(r.Context(), stackCandidateTemplateKey(decoded.BundleDigest))
			if deployedErr == nil && candidateErr == nil && deployedFound && candidateTemplateFound {
				if err := addStackPropertyChanges(&decoded, deployedTemplate, candidateTemplate); err != nil {
					p.logger.Warn("unable to derive stack property changes", zap.Error(err))
				} else if enriched, err := json.Marshal(decoded); err == nil {
					stackCandidate = enriched
				}
			}
		}
	}
	writeJSON(w, struct {
		Active             json.RawMessage           `json:"active"`
		Candidate          json.RawMessage           `json:"candidate"`
		CandidateRecordKey string                    `json:"candidate_record_key,omitempty"`
		StackCandidate     json.RawMessage           `json:"stack_candidate"`
		History            []operation.BundleInfo    `json:"history"`
		Comparisons        []bundleHistoryComparison `json:"comparisons"`
	}{Active: active, Candidate: candidate, CandidateRecordKey: candidateRecordKey, StackCandidate: stackCandidate, History: history, Comparisons: comparisons})
}

func (p *portalServer) approveBundleCandidate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		BundleDigest string `json:"bundle_digest"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeAPIError(w, err, http.StatusBadRequest)
		return
	}
	candidate, _, found, err := p.latestBundleCandidateRecord(r.Context())
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if !found {
		writeAPIError(w, fmt.Errorf("bundle candidate not found"), http.StatusNotFound)
		return
	}
	if input.BundleDigest == "" || input.BundleDigest != candidate.Bundle.BundleDigest {
		writeAPIError(w, fmt.Errorf("candidate changed; reload before approving"), http.StatusConflict)
		return
	}
	activeRaw, found, err := p.store.Get(r.Context(), operation.BundleKey)
	if err != nil || !found {
		writeAPIError(w, errors.Join(err, fmt.Errorf("active bundle not found")), http.StatusConflict)
		return
	}
	var active operation.BundleInfo
	if err := json.Unmarshal(activeRaw, &active); err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if active.BundleDigest != candidate.PreviousDigest || active.BundleDigest == candidate.Bundle.BundleDigest {
		writeAPIError(w, fmt.Errorf("candidate is not based on the active bundle"), http.StatusConflict)
		return
	}
	statusRaw, found, err := p.store.Get(r.Context(), "status.json")
	if err != nil || !found {
		writeAPIError(w, errors.Join(err, fmt.Errorf("candidate plan status not found")), http.StatusConflict)
		return
	}
	var status statestore.Status
	if err := json.Unmarshal(statusRaw, &status); err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	approvalPhase := status.ApprovalPhase
	approvalKeys := []string{}
	if status.BundleDigest == candidate.Bundle.BundleDigest && status.ApprovalRequired {
		stepStatuses := make(map[string]string, len(status.Steps))
		for _, step := range status.Steps {
			stepStatuses[step.ID] = step.Status
		}
		approvalKey, err := candidateApprovalKey(candidate, status.ApprovalPhase, stepStatuses)
		if err != nil {
			writeAPIError(w, err, http.StatusConflict)
			return
		}
		approvalKeys = append(approvalKeys, approvalKey)
	} else {
		approvalKeys, err = p.candidatePlanApprovalKeys(r, candidate)
		if err != nil {
			writeAPIError(w, err, http.StatusConflict)
			return
		}
		approvalPhase = "deployment"
	}
	approval := []byte(fmt.Sprintf(`{"approved_at":%q,"approved_by":%q}`, time.Now().UTC().Format(time.RFC3339Nano), p.requestedBy))
	for _, approvalKey := range approvalKeys {
		if err := p.controlStore.PutIfAbsent(r.Context(), approvalKey, approval); err != nil && !errors.Is(err, operationstate.ErrObjectExists) {
			writeAPIError(w, err, http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, map[string]string{"bundle_digest": candidate.Bundle.BundleDigest, "status": "approved", "phase": approvalPhase})
}

func candidateApprovalKey(candidate operation.BundleCandidate, phase string, stepStatuses map[string]string) (string, error) {
	if phase == "sandbox" {
		for _, change := range candidate.Changes {
			if change.Kind != operation.BundleContentKindSandbox || change.Change == operation.BundleChangeUnchanged {
				continue
			}
			if change.Change != operation.BundleChangeChanged || change.Name != "terraform" || change.PlanStepID == "" || stepStatuses[change.PlanStepID] != operation.RunStatusFinished {
				return "", fmt.Errorf("terraform sandbox plan %q has not finished", change.PlanStepID)
			}
			return operation.CandidateSandboxApprovalKey(candidate.Bundle.BundleDigest), nil
		}
		return "", fmt.Errorf("terraform sandbox plan not found")
	}
	for _, change := range candidate.Changes {
		if change.Kind != operation.BundleContentKindComponent || change.Change == operation.BundleChangeUnchanged {
			continue
		}
		if change.Change == operation.BundleChangeRemoved {
			return "", fmt.Errorf("component removals are not supported")
		}
		if change.PlanStepID == "" || stepStatuses[change.PlanStepID] != operation.RunStatusFinished {
			return "", fmt.Errorf("component plan %q has not finished", change.PlanStepID)
		}
	}
	return operation.CandidateApprovalKey(candidate.Bundle.BundleDigest), nil
}

func (p *portalServer) candidatePlanApprovalKeys(r *http.Request, candidate operation.BundleCandidate) ([]string, error) {
	runs, err := p.listRuns(r)
	if err != nil {
		return nil, err
	}
	for _, run := range runs {
		if run.RefKind != operation.RefKindBundlePlan || run.BundleDigest != candidate.Bundle.BundleDigest || run.Status != operation.RunStatusFinished {
			continue
		}
		stepStatuses := make(map[string]string, len(run.Steps))
		for _, step := range run.Steps {
			stepStatuses[step.ID] = step.Status
		}
		keys := []string{operation.CandidateApprovalKey(candidate.Bundle.BundleDigest)}
		for _, change := range candidate.Changes {
			if change.Change == operation.BundleChangeUnchanged || (change.Kind != operation.BundleContentKindComponent && change.Kind != operation.BundleContentKindSandbox) {
				continue
			}
			if change.PlanStepID == "" || stepStatuses[change.PlanStepID] != operation.RunStatusFinished {
				return nil, fmt.Errorf("candidate plan %q has not finished", change.PlanStepID)
			}
			if change.Kind == operation.BundleContentKindSandbox {
				keys = append(keys, operation.CandidateSandboxApprovalKey(candidate.Bundle.BundleDigest))
			}
		}
		return keys, nil
	}
	return nil, fmt.Errorf("completed candidate deployment plan not found")
}

func (p *portalServer) planBundleCandidateStack(w http.ResponseWriter, r *http.Request) {
	if p.stackPlanner == nil || p.installStackName == "" {
		writeAPIError(w, fmt.Errorf("install stack planning is unavailable in local mode"), http.StatusConflict)
		return
	}
	var input struct {
		BundleDigest string `json:"bundle_digest"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeAPIError(w, err, http.StatusBadRequest)
		return
	}
	candidate, candidateRecordKey, found, err := p.latestBundleCandidateRecord(r.Context())
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if !found {
		writeAPIError(w, fmt.Errorf("bundle candidate not found"), http.StatusNotFound)
		return
	}
	if input.BundleDigest == "" || input.BundleDigest != candidate.Bundle.BundleDigest {
		writeAPIError(w, fmt.Errorf("candidate changed; reload before planning"), http.StatusConflict)
		return
	}
	if candidate.Deployment == nil {
		writeAPIError(w, fmt.Errorf("candidate was staged without deployment assets; stage it with a current bundle CLI"), http.StatusConflict)
		return
	}
	if err := p.requireRunnerCapability(r.Context(), customermanaged.RunnerCapabilityCandidateArtifactPlans); err != nil {
		writeAPIError(w, err, http.StatusConflict)
		return
	}
	template, found, err := p.stackStore.Get(r.Context(), stackCandidateTemplateKey(candidate.Bundle.BundleDigest))
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if !found {
		writeAPIError(w, fmt.Errorf("candidate install stack template not found"), http.StatusConflict)
		return
	}
	if !p.stackPlanMu.TryLock() {
		writeAPIError(w, fmt.Errorf("another bundle plan is already running"), http.StatusConflict)
		return
	}
	planStarted := false
	defer func() {
		if !planStarted {
			p.stackPlanMu.Unlock()
		}
	}()
	if raw, ok, err := p.store.Get(r.Context(), operation.StackCandidateKey); err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	} else if ok {
		var existing operation.StackCandidate
		if err := json.Unmarshal(raw, &existing); err != nil {
			writeAPIError(w, err, http.StatusInternalServerError)
			return
		}
		if existing.CandidateRecordKey == candidateRecordKey {
			if existing.Status == "CREATE_PENDING" {
				writeAPIError(w, fmt.Errorf("install stack plan is already running"), http.StatusConflict)
				return
			}
			if existing.Status != "FAILED" {
				writeAPIError(w, fmt.Errorf("install stack plan already exists"), http.StatusConflict)
				return
			}
		}
	}
	pending := &operation.StackCandidate{
		SchemaVersion: 1, BundleDigest: candidate.Bundle.BundleDigest, StackName: p.installStackName,
		TemplateURL: candidate.Deployment.StackTemplateURL, CandidateBundleKey: candidate.Deployment.CandidateBundleKey,
		TargetBundleKey: candidate.Deployment.TargetBundleKey, CandidateRecordKey: candidateRecordKey,
		Status: "CREATE_PENDING", ExecutionStatus: "UNAVAILABLE", CreatedAt: time.Now().UTC(),
	}
	raw, err := json.Marshal(pending)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	startedAt := time.Now().UTC()
	planStepIDs, err := candidatePlanStepIDs(candidate)
	if err != nil {
		writeAPIError(w, err, http.StatusConflict)
		return
	}
	steps := []operation.RunStep{{ID: "install-stack-plan", Name: "Plan install stack", Kind: "cloudformation-plan", Status: operation.RunStatusInProgress, StartedAt: &startedAt}}
	for _, id := range planStepIDs {
		name, kind := candidatePlanStep(candidate, id)
		steps = append(steps, operation.RunStep{ID: id, Name: name, Kind: kind, Status: "available"})
	}
	planRun := operation.RunStatus{
		RunID: "bundle-plan-" + ulid.Make().String(), RefID: candidateRecordKey, RefKind: "bundle-plan",
		RefName: "Bundle deployment plan", Source: operation.SourcePortal, Status: operation.RunStatusInProgress,
		BundleDigest: candidate.Bundle.BundleDigest, StartedAt: startedAt,
		Steps: steps,
	}
	planRunRaw, err := json.Marshal(planRun)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if err := p.controlStore.Put(r.Context(), operation.RunStatusKey(planRun.RunID), planRunRaw); err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if err := p.controlStore.Put(r.Context(), operation.StackCandidateKey, raw); err != nil {
		finishedAt := time.Now().UTC()
		planRun.Status, planRun.Error, planRun.FinishedAt = operation.RunStatusFailed, err.Error(), &finishedAt
		planRun.Steps[0].Status, planRun.Steps[0].Error, planRun.Steps[0].FinishedAt = operation.RunStatusFailed, err.Error(), &finishedAt
		if failedRaw, marshalErr := json.Marshal(planRun); marshalErr == nil {
			_ = p.controlStore.Put(r.Context(), operation.RunStatusKey(planRun.RunID), failedRaw)
		}
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	planStarted = true
	go func() {
		defer p.stackPlanMu.Unlock()
		p.generateStackPlan(candidate, candidateRecordKey, template, *pending, planRun, planStepIDs)
	}()
	writeJSON(w, pending)
}

func (p *portalServer) requireRunnerCapability(ctx context.Context, capability string) error {
	raw, found, err := p.store.Get(ctx, customermanaged.RunnerHeartbeatKey)
	if err != nil {
		return fmt.Errorf("read runner heartbeat: %w", err)
	}
	if !found {
		return fmt.Errorf("runner heartbeat is unavailable; wait for a compatible runner before planning")
	}
	var heartbeat customermanaged.RunnerHeartbeat
	if err := json.Unmarshal(raw, &heartbeat); err != nil {
		return fmt.Errorf("decode runner heartbeat: %w", err)
	}
	if !heartbeat.Supports(capability) {
		return fmt.Errorf("runner does not support candidate artifact plans; update the runner before planning")
	}
	return nil
}

func candidatePlanStepIDs(candidate operation.BundleCandidate) ([]string, error) {
	ids := make([]string, 0)
	for _, change := range candidate.Changes {
		switch {
		case change.Kind == operation.BundleContentKindSandbox && change.Change != operation.BundleChangeUnchanged:
			if change.Change != operation.BundleChangeChanged || change.PlanStepID == "" {
				return nil, fmt.Errorf("sandbox transition %q for %q cannot be planned", change.Change, change.Name)
			}
			ids = append(ids, change.PlanStepID)
		case change.Kind == operation.BundleContentKindComponent && change.Change == operation.BundleChangeRemoved:
			return nil, fmt.Errorf("component removal for %q cannot be planned", change.Name)
		case change.Kind == operation.BundleContentKindComponent && change.Change != operation.BundleChangeUnchanged:
			if change.PlanStepID == "" {
				return nil, fmt.Errorf("component %q has no plan step", change.Name)
			}
			ids = append(ids, change.PlanStepID)
		}
	}
	return ids, nil
}

func candidatePlanStep(candidate operation.BundleCandidate, id string) (string, string) {
	for _, change := range candidate.Changes {
		if change.PlanStepID != id {
			continue
		}
		if change.Kind == operation.BundleContentKindSandbox {
			return "Plan sandbox", "terraform-plan"
		}
		return "Plan " + change.Name, "component-plan"
	}
	return id, "component-plan"
}

func (p *portalServer) generateStackPlan(candidate operation.BundleCandidate, candidateRecordKey string, template []byte, pending operation.StackCandidate, planRun operation.RunStatus, planStepIDs []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	planned, err := createStackPlan(ctx, p.stackPlanner, p.installStackName, candidate, template)
	finishedAt := time.Now().UTC()
	planRun.Steps[0].FinishedAt = &finishedAt
	if err != nil {
		pending.Status = "FAILED"
		pending.StatusReason = err.Error()
		planned = &pending
		planRun.Status = operation.RunStatusFailed
		planRun.FinishedAt = &finishedAt
		planRun.Error = err.Error()
		planRun.Steps[0].Status = operation.RunStatusFailed
		planRun.Steps[0].Error = err.Error()
		for i := 1; i < len(planRun.Steps); i++ {
			planRun.Steps[i].Status = operation.StepStatusDiscarded
		}
		p.logger.Error("install stack plan failed", zap.String("bundle_digest", candidate.Bundle.BundleDigest), zap.Error(err))
	} else {
		planned.CandidateRecordKey = candidateRecordKey
		planRun.Steps[0].Status = operation.RunStatusFinished
		if len(planStepIDs) == 0 {
			planRun.Status = operation.RunStatusFinished
			planRun.FinishedAt = &finishedAt
		}
	}
	raw, marshalErr := json.Marshal(planned)
	if marshalErr != nil {
		p.logger.Error("encode install stack plan", zap.Error(marshalErr))
		return
	}
	persistCtx := context.Background()
	if err := p.controlStore.Put(persistCtx, operation.StackCandidateKey, raw); err != nil {
		p.logger.Error("persist install stack plan", zap.Error(err))
	}
	if err == nil {
		if resultErr := p.controlStore.Put(persistCtx, operation.RunStepResultKey(planRun.RunID, "install-stack-plan"), raw); resultErr != nil {
			p.logger.Error("persist immutable install stack plan result", zap.Error(resultErr))
		}
	}
	planRunRaw, marshalErr := json.Marshal(planRun)
	if marshalErr != nil {
		p.logger.Error("encode bundle plan run", zap.Error(marshalErr))
		return
	}
	if err := p.controlStore.Put(persistCtx, operation.RunStatusKey(planRun.RunID), planRunRaw); err != nil {
		p.logger.Error("persist bundle plan run", zap.Error(err))
	}
	if err != nil || len(planStepIDs) == 0 {
		return
	}
	request := operation.Request{
		SchemaVersion: operation.SchemaVersion, DeploymentID: candidate.Bundle.DeploymentID,
		BundleDigest: candidate.Bundle.BundleDigest, RefID: candidateRecordKey,
		RefKind: operation.RefKindBundlePlan, RunID: planRun.RunID,
		CandidateArchiveKey: candidate.Deployment.CandidateBundleKey, CandidateRecordKey: candidateRecordKey,
		PlanStepIDs: planStepIDs, DispatchID: planRun.RunID, Source: operation.SourcePortal,
		RequestedBy: p.requestedBy, CreatedAt: time.Now().UTC(),
	}
	requestRaw, marshalErr := json.Marshal(request)
	if marshalErr != nil {
		p.failBundlePlanRun(planRun, marshalErr)
		return
	}
	if putErr := p.controlStore.PutIfAbsent(persistCtx, operation.RequestKey(request.DispatchID), requestRaw); putErr != nil && !errors.Is(putErr, operationstate.ErrObjectExists) {
		p.failBundlePlanRun(planRun, putErr)
	}
}

func (p *portalServer) failBundlePlanRun(run operation.RunStatus, err error) {
	finishedAt := time.Now().UTC()
	run.Status, run.Error, run.FinishedAt = operation.RunStatusFailed, err.Error(), &finishedAt
	for i := 1; i < len(run.Steps); i++ {
		if run.Steps[i].Status == "available" {
			run.Steps[i].Status = operation.StepStatusDiscarded
		}
	}
	raw, marshalErr := json.Marshal(run)
	if marshalErr == nil {
		_ = p.controlStore.Put(context.Background(), operation.RunStatusKey(run.RunID), raw)
	}
	p.logger.Error("dispatch candidate bundle plan", zap.Error(err))
}

func (p *portalServer) latestBundleCandidate(ctx context.Context) (operation.BundleCandidate, bool, error) {
	candidate, _, found, err := p.latestBundleCandidateRecord(ctx)
	return candidate, found, err
}

func (p *portalServer) latestBundleCandidateRecord(ctx context.Context) (operation.BundleCandidate, string, bool, error) {
	var latest operation.BundleCandidate
	latestDismissals := make(map[string]time.Time)
	latestKey := ""
	found := false
	activeDigest := ""
	if raw, ok, err := p.store.Get(ctx, operation.BundleKey); err != nil {
		return operation.BundleCandidate{}, "", false, err
	} else if ok {
		var active operation.BundleInfo
		if err := json.Unmarshal(raw, &active); err != nil {
			return operation.BundleCandidate{}, "", false, fmt.Errorf("decode active bundle: %w", err)
		}
		activeDigest = active.BundleDigest
	}
	keys := []string{operation.CandidateKey, operation.StagedCandidateKey}
	stageKeys, err := p.store.List(ctx, operation.CandidateStagesPrefix)
	if err != nil {
		return operation.BundleCandidate{}, "", false, err
	}
	for _, key := range stageKeys {
		if strings.Contains(key, "/staged/") && strings.HasSuffix(key, ".json") {
			keys = append(keys, key)
			continue
		}
		if strings.Contains(key, "/dismissed/") && strings.HasSuffix(key, ".json") {
			raw, ok, err := p.store.Get(ctx, key)
			if err != nil {
				return operation.BundleCandidate{}, "", false, err
			}
			if !ok {
				continue
			}
			var dismissal operation.BundleCandidateDismissal
			if err := json.Unmarshal(raw, &dismissal); err != nil {
				return operation.BundleCandidate{}, "", false, fmt.Errorf("decode %s: %w", key, err)
			}
			if dismissal.DismissedAt.After(latestDismissals[dismissal.BundleDigest]) {
				latestDismissals[dismissal.BundleDigest] = dismissal.DismissedAt
			}
		}
	}
	for _, key := range keys {
		raw, ok, err := p.store.Get(ctx, key)
		if err != nil {
			return operation.BundleCandidate{}, "", false, err
		}
		if !ok {
			continue
		}
		var candidate operation.BundleCandidate
		if err := json.Unmarshal(raw, &candidate); err != nil {
			return operation.BundleCandidate{}, "", false, fmt.Errorf("decode %s: %w", key, err)
		}
		if candidate.Bundle.BundleDigest == activeDigest {
			continue
		}
		if !found || candidate.StagedAt.After(latest.StagedAt) || (candidate.StagedAt.Equal(latest.StagedAt) && strings.Contains(key, "/staged/")) {
			latest = candidate
			latestKey = key
			found = true
		}
	}
	if found {
		if dismissedAt, dismissed := latestDismissals[latest.Bundle.BundleDigest]; dismissed && !dismissedAt.Before(latest.StagedAt) {
			return operation.BundleCandidate{}, "", false, nil
		}
	}
	return latest, latestKey, found, nil
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

func (p *portalServer) runnerHeartbeat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	raw, ok, err := p.store.Get(r.Context(), customermanaged.RunnerHeartbeatKey)
	if err == nil && !ok {
		raw, ok, err = p.store.Get(r.Context(), customermanaged.LegacyRunnerHeartbeatKey)
	}
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if !ok {
		raw = []byte("null")
	}
	writeRawJSON(w, raw)
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
	if !jobIDPattern.MatchString(id) {
		writeAPIError(w, fmt.Errorf("invalid run ID"), http.StatusBadRequest)
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

func (p *portalServer) runControl(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !jobIDPattern.MatchString(id) {
		writeAPIError(w, fmt.Errorf("invalid run ID"), http.StatusBadRequest)
		return
	}
	var input struct {
		Action string `json:"action"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeAPIError(w, err, http.StatusBadRequest)
		return
	}
	if input.Action != statestore.ControlActionRetry && input.Action != statestore.ControlActionUserSkip && input.Action != statestore.ControlActionCancel {
		writeAPIError(w, fmt.Errorf("invalid control action"), http.StatusBadRequest)
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
	if (input.Action == statestore.ControlActionRetry || input.Action == statestore.ControlActionUserSkip) && run.Status != statestore.RunStatusFailedPendingRetry {
		writeAPIError(w, fmt.Errorf("run is not awaiting a retry decision"), http.StatusConflict)
		return
	}
	if input.Action == statestore.ControlActionCancel && run.Status != statestore.RunStatusInProgress && run.Status != statestore.RunStatusFailedPendingRetry {
		writeAPIError(w, fmt.Errorf("run is not cancellable"), http.StatusConflict)
		return
	}
	request := statestore.ControlRequest{RunID: id, Action: input.Action, RequestedBy: p.requestedBy, CreatedAt: time.Now().UTC()}
	raw, err := json.Marshal(request)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if err := p.controlStore.PutIfAbsent(r.Context(), statestore.InstallControlKey(id, input.Action), raw); err != nil {
		if errors.Is(err, operationstate.ErrObjectExists) {
			writeJSON(w, request)
			return
		}
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, request)
}

func (p *portalServer) plan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !jobIDPattern.MatchString(id) {
		writeAPIError(w, fmt.Errorf("invalid job ID"), http.StatusBadRequest)
		return
	}
	p.object(operation.JobPlanKey(id))(w, r)
}

// stepPlan serves the late-bound rendered plan the customer-managed runner persisted
// for a bootstrap install step, so the portal can preview what the step will
// apply before and while it executes.
func (p *portalServer) stepPlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !jobIDPattern.MatchString(id) {
		writeAPIError(w, fmt.Errorf("invalid step ID"), http.StatusBadRequest)
		return
	}
	p.object(statestore.StepPlanKey(id))(w, r)
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
	request := operation.Request{
		SchemaVersion: operation.SchemaVersion,
		DeploymentID:  catalog.DeploymentID,
		BundleDigest:  catalog.BundleDigest,
		RefID:         input.RefID,
		DispatchID:    id,
		Source:        operation.SourcePortal,
		RequestedBy:   p.requestedBy,
		CreatedAt:     time.Now().UTC(),
	}
	raw, err := json.Marshal(request)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if err := p.controlStore.PutIfAbsent(r.Context(), operation.RequestKey(id), raw); err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	p.logger.Info("dispatch requested", zap.String("dispatch_id", id), zap.String("ref_id", input.RefID), zap.String("requested_by", p.requestedBy))
	writeJSON(w, map[string]string{"dispatch_id": id})
}

func (p *portalServer) readCatalog(r *http.Request) (*operation.Catalog, error) {
	raw, ok, err := p.store.Get(r.Context(), operation.CatalogKey)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("no operation catalog found")
	}
	var catalog operation.Catalog
	if err := json.Unmarshal(raw, &catalog); err != nil {
		return nil, fmt.Errorf("decode %s: %w", operation.CatalogKey, err)
	}
	return &catalog, nil
}

func catalogHasRef(catalog *operation.Catalog, id string) bool {
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
		if strings.HasPrefix(key, operation.ReceiptsPrefix) {
			receipts[strings.TrimSuffix(strings.TrimPrefix(key, operation.ReceiptsPrefix), ".json")] = true
		}
	}
	items := make([]json.RawMessage, 0)
	for _, key := range keys {
		if !strings.HasPrefix(key, operation.RequestsPrefix) && !strings.HasPrefix(key, operation.ReceiptsPrefix) {
			continue
		}
		if strings.HasPrefix(key, operation.RequestsPrefix) && receipts[strings.TrimSuffix(strings.TrimPrefix(key, operation.RequestsPrefix), ".json")] {
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
