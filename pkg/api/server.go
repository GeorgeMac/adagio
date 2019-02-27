package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/repository"
)

const (
	versionPrefix = "/v0"
)

type Server struct {
	*http.ServeMux

	repo repository.Repository
}

func NewServer(repo repository.Repository) *Server {
	s := &Server{
		ServeMux: http.NewServeMux(),
		repo:     repo,
	}

	s.handle("/execute", http.HandlerFunc(s.handleExecute))
	s.handle("/runs", http.HandlerFunc(s.handleListRuns))

	return s
}

func (s *Server) handle(p string, h http.Handler) {
	s.ServeMux.Handle(path.Join(versionPrefix, p), h)
}

func (s *Server) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" && r.Method != "POST" {
		http.Error(w, fmt.Sprintf("%s not supported", r.Method), http.StatusMethodNotAllowed)
		return
	}

	var graph adagio.Graph
	if err := json.NewDecoder(r.Body).Decode(&graph); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	run, err := s.repo.StartRun(graph)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&run); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, fmt.Sprintf("%s not supported", r.Method), http.StatusMethodNotAllowed)
		return
	}

	runs, err := s.repo.ListRuns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(runs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := json.NewEncoder(w).Encode(&runs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
