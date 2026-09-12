package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/workflow"
	"go.uber.org/zap"
)

type Server struct {
	workflows *workflow.Service
	logger    *zap.Logger
}

func NewServer(workflows *workflow.Service, logger *zap.Logger) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Server{workflows: workflows, logger: logger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/v1/workflows", s.listWorkflows)
	mux.HandleFunc("POST /api/v1/workflows/bulk", s.createBulkWorkflows)
	mux.HandleFunc("POST /api/v1/workflows/{id}/transition", s.transitionWorkflow)
	mux.HandleFunc("POST /api/v1/workflows/{id}/metrics", s.applyMetrics)
	return s.withRequestLogging(withCORS(mux))
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "time": time.Now().UTC()})
}

func (s *Server) listWorkflows(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"data": s.workflows.List()})
}

func (s *Server) createBulkWorkflows(w http.ResponseWriter, r *http.Request) {
	var req domain.BulkWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	created, err := s.workflows.CreateBulk(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": created, "count": len(created)})
}

func (s *Server) transitionWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		State domain.WorkflowState `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(string(req.State)) == "" {
		writeError(w, http.StatusBadRequest, "state is required")
		return
	}
	item, err := s.workflows.Transition(id, req.State)
	if errors.Is(err, workflow.ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (s *Server) applyMetrics(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var metrics domain.CampaignMetrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		writeError(w, http.StatusBadRequest, "invalid metrics body")
		return
	}
	metrics.CapturedAt = time.Now().UTC()
	item, err := s.workflows.ApplyMetrics(id, metrics)
	if errors.Is(err, workflow.ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (s *Server) withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(recorder, r)

		s.logger.Info("http request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", recorder.status),
			zap.Duration("duration", time.Since(started)),
			zap.String("remote_addr", r.RemoteAddr),
		)
	})
}
