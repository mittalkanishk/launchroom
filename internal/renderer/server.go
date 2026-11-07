package renderer

import (
	"encoding/json"
	"net/http"
)

// Request is one scene to render.
type Request struct {
	Source string `json:"source"`
}

type errorReply struct {
	Error string `json:"error"`
}

// Handler is the renderer's HTTP API, reachable only inside the stack:
//
//	POST /v1/render   a Request; 200 with a Result, 422 if the scene fails to render
//	GET  /healthz
func (r *Renderer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/render", r.handleRender)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
	return mux
}

func (r *Renderer) handleRender(w http.ResponseWriter, req *http.Request) {
	var in Request
	if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 1<<20)).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, errorReply{Error: err.Error()})
		return
	}
	result, err := r.Render(req.Context(), in.Source)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorReply{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
