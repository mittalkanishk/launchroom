package studio

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Handler is studio's HTTP API, reachable only inside the stack:
//
//	POST /v1/jobs                     submit a Request; 202 with the Job
//	GET  /v1/jobs?video=<name>        the video's jobs
//	GET  /v1/jobs/{id}                one job
//	GET  /v1/files/{video}/{path...}  a file a job wrote
//	GET  /healthz
func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/jobs", s.handleSubmit)
	mux.HandleFunc("GET /v1/jobs", s.handleJobs)
	mux.HandleFunc("GET /v1/jobs/{id}", s.handleJob)
	mux.HandleFunc("GET /v1/files/{video}/{path...}", s.handleFile)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	return mux
}

func (s *Service) handleSubmit(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	job, err := s.Submit(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSONResponse(w, http.StatusAccepted, job)
}

func (s *Service) handleJobs(w http.ResponseWriter, r *http.Request) {
	writeJSONResponse(w, http.StatusOK, s.Jobs(r.URL.Query().Get("video")))
}

func (s *Service) handleJob(w http.ResponseWriter, r *http.Request) {
	job, ok := s.Job(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, errors.New("no such job"))
		return
	}
	writeJSONResponse(w, http.StatusOK, job)
}

func (s *Service) handleFile(w http.ResponseWriter, r *http.Request) {
	path, err := s.File(r.PathValue("video"), r.PathValue("path"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	http.ServeFile(w, r, path)
}

type errorReply struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSONResponse(w, status, errorReply{Error: err.Error()})
}

func writeJSONResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
