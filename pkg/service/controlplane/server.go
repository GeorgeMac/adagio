package controlplane

import (
	"context"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
	"github.com/pkg/errors"
)

var _ controlplane.ControlPlaneServer = (*Service)(nil)

type Repository interface {
	Stats() (*adagio.Stats, error)
	StartRun(*adagio.GraphSpec) (*adagio.Run, error)
	InspectRun(id string) (*adagio.Run, error)
	ListRuns() ([]*adagio.Run, error)
	ListAgents() ([]*adagio.Agent, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	s := &Service{
		repo: repo,
	}

	return s
}

func (s *Service) Stats(_ context.Context, req *controlplane.StatsRequest) (*controlplane.StatsResponse, error) {
	stats, err := s.repo.Stats()
	if err != nil {
		return nil, err
	}

	return &controlplane.StatsResponse{Stats: stats}, nil
}

func (s *Service) Start(_ context.Context, req *controlplane.StartRequest) (*controlplane.StartResponse, error) {
	run, err := s.repo.StartRun(req.Spec)
	if err != nil {
		return nil, errors.Wrap(err, "control plane: starting run")
	}

	return &controlplane.StartResponse{Run: run}, nil
}

func (s *Service) Inspect(_ context.Context, req *controlplane.InspectRequest) (*controlplane.InspectResponse, error) {
	run, err := s.repo.InspectRun(req.Id)
	if err != nil {
		return nil, errors.Wrap(err, "control plane: starting run")
	}

	return &controlplane.InspectResponse{Run: run}, nil
}

func (s *Service) ListRuns(_ context.Context, _ *controlplane.ListRequest) (*controlplane.ListRunsResponse, error) {
	runs, err := s.repo.ListRuns()
	if err != nil {
		return nil, errors.Wrap(err, "control plane: listing runs")
	}

	return &controlplane.ListRunsResponse{Runs: runs}, nil
}

func (s *Service) ListAgents(_ context.Context, _ *controlplane.ListRequest) (*controlplane.ListAgentsResponse, error) {
	agents, err := s.repo.ListAgents()
	if err != nil {
		return nil, errors.Wrap(err, "control place: listing agents")
	}

	return &controlplane.ListAgentsResponse{Agents: agents}, nil
}
