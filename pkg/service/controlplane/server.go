package controlplane

import (
	"context"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/georgemac/adagio/pkg/rpc/controlplane"
	"github.com/pkg/errors"
)

type Repository interface {
	StartRun(adagio.Graph) (*adagio.Run, error)
	ListRuns() ([]*adagio.Run, error)
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

func (s *Service) Start(_ context.Context, pbgraph *controlplane.Graph) (*controlplane.Run, error) {
	graph, err := toAdagioGraph(pbgraph)
	if err != nil {
		return nil, errors.Wrap(err, "control plane: starting run")
	}

	run, err := s.repo.StartRun(graph)
	if err != nil {
		return nil, errors.Wrap(err, "control plane: starting run")
	}

	return toPBRun(run), nil
}

func (s *Service) List(_ context.Context, _ *controlplane.ListRequest) (*controlplane.ListResponse, error) {
	runs, err := s.repo.ListRuns()
	if err != nil {
		return nil, errors.Wrap(err, "control plane: listing runs")
	}

	return &controlplane.ListResponse{Runs: toPBRuns(runs...)}, nil
}
