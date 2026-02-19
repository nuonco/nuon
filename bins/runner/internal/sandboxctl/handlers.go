package sandboxctl

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/sandbox/state", s.handleGetState)
	mux.HandleFunc("POST /api/v1/sandbox/reset", s.handleReset)
	mux.HandleFunc("GET /api/v1/sandbox/job-types", s.handleGetJobTypes)
	mux.HandleFunc("GET /api/v1/sandbox/job-types/", s.handleGetOrUpdateJobType)
	mux.HandleFunc("PUT /api/v1/sandbox/job-types/", s.handleGetOrUpdateJobType)
	mux.HandleFunc("GET /api/v1/sandbox/presets", s.handleGetPresets)
	mux.HandleFunc("POST /api/v1/sandbox/panic", s.handlePanic)
	mux.HandleFunc("POST /api/v1/sandbox/shutdown", s.handleShutdown)
	mux.HandleFunc("GET /api/v1/sandbox/stats", s.handleGetStats)
	mux.HandleFunc("GET /", s.handleUI)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.state.Snapshot())
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	s.state.Reset()
	writeJSON(w, http.StatusOK, map[string]string{"status": "reset"})
}

func (s *Server) handleGetJobTypes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.state.GetAllConfigs())
}

func (s *Server) handleGetOrUpdateJobType(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/sandbox/job-types/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "job type required")
		return
	}

	jobType := parts[0]

	if len(parts) == 3 && parts[1] == "preset" && r.Method == http.MethodPut {
		preset := ResponsePreset(parts[2])
		cfg := ApplyPreset(preset, s.state.DefaultDuration)
		s.state.SetConfig(jobType, cfg)
		writeJSON(w, http.StatusOK, cfg)
		return
	}

	switch r.Method {
	case http.MethodGet:
		cfg := s.state.GetConfig(jobType)
		writeJSON(w, http.StatusOK, cfg)
	case http.MethodPut:
		var cfg JobTypeConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		s.state.SetConfig(jobType, &cfg)
		writeJSON(w, http.StatusOK, &cfg)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleGetPresets(w http.ResponseWriter, r *http.Request) {
	result := make(map[string][]PresetInfo)
	for _, jt := range AllJobTypes() {
		result[jt] = PresetsForJobType(jt)
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handlePanic(w http.ResponseWriter, r *http.Request) {
	s.state.SetPanic()
	writeJSON(w, http.StatusOK, map[string]string{"status": "panic armed for next job"})
}

func (s *Server) handleShutdown(w http.ResponseWriter, r *http.Request) {
	s.state.SetShutdown()
	writeJSON(w, http.StatusOK, map[string]string{"status": "shutdown armed for next job"})
}

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.state.GetStats())
}
