package daemon

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// APIServer manages standard HTTP API endpoints.
type APIServer struct {
	daemon *Daemon
	mux    *http.ServeMux
}

// NewAPIServer creates and configures the HTTP multiplexer.
func NewAPIServer(d *Daemon) *APIServer {
	s := &APIServer{
		daemon: d,
		mux:    http.NewServeMux(),
	}
	s.routes()
	return s
}

// ServeHTTP implements http.Handler interface.
func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS - Allow origin from CODEFORGE_CORS_ORIGIN env var (defaults to * for local dev)
	corsOrigin := os.Getenv("CODEFORGE_CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}

func (s *APIServer) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/status", s.handleStatus)
	s.mux.HandleFunc("/logs/", s.handleLogs)
	s.mux.HandleFunc("/trigger/", s.handleTrigger)
	s.mux.HandleFunc("/rollback/", s.handleRollback)
}

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *APIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	s.daemon.mu.RLock()
	defer s.daemon.mu.RUnlock()

	type PipelineStatus struct {
		Project      string `json:"project"`
		Status       string `json:"status"`
		LastRun      string `json:"last_run"`
		LastDuration string `json:"last_duration"`
		ConfigPath   string `json:"config_path"`
	}

	res := []PipelineStatus{}
	for name, p := range s.daemon.pipelines {
		var lastRunStr string
		if !p.LastRun.IsZero() {
			lastRunStr = p.LastRun.Format("2006-01-02T15:04:05Z07:00")
		}

		res = append(res, PipelineStatus{
			Project:      name,
			Status:       p.LastStatus,
			LastRun:      lastRunStr,
			LastDuration: p.LastDuration.String(),
			ConfigPath:   p.Path,
		})
	}

	writeJSON(w, http.StatusOK, res)
}

func (s *APIServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	project := strings.TrimPrefix(r.URL.Path, "/logs/")
	if project == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing project parameter"})
		return
	}

	// Read optional query parameter for limit
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	lines, err := s.daemon.logger.TailLines(project, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project": project,
		"logs":    lines,
	})
}

func (s *APIServer) handleTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	project := strings.TrimPrefix(r.URL.Path, "/trigger/")
	if project == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing project parameter"})
		return
	}

	err := s.daemon.Trigger(project, "manual API call")
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"project": project,
		"status":  "triggered",
	})
}

func (s *APIServer) handleRollback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	project := strings.TrimPrefix(r.URL.Path, "/rollback/")
	if project == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing project parameter"})
		return
	}

	err := s.daemon.TriggerRollback(project)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"project": project,
		"status":  "rollback_triggered",
	})
}

func writeJSON(w http.ResponseWriter, code int, val interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, _ := json.Marshal(val)
	_, _ = w.Write(data)
}
