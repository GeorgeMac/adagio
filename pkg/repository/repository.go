package repository

import (
	"errors"

	"github.com/georgemac/adagio/pkg/adagio"
)

var (
	// ErrRunDoesNotExist is returned when an attempt is made to interface
	// with a non existent run
	ErrRunDoesNotExist = errors.New("run does not exist")
	// ErrNodeNotReady is returned when an attempt is made to claim a node in a waiting
	// state
	ErrNodeNotReady = errors.New("node not ready")
)

type Repository interface {
	StartRun(adagio.Graph) (*adagio.Run, error)
	ListRuns() ([]*adagio.Run, error)
	NodeRepository
	NodeWatcher
}

type NodeRepository interface {
	ClaimNode(*adagio.Run, *adagio.Node) (bool, error)
	FinishNode(*adagio.Run, *adagio.Node) error
	RecoverNode(*adagio.Run, *adagio.Node) (bool, error)
	BuryNode(*adagio.Run, *adagio.Node) error
}

type Event struct {
	Run      *adagio.Run
	Node     *adagio.Node
	From, To adagio.NodeState
}

type NodeWatcher interface {
	Subscribe(events chan<- Event, states ...adagio.NodeState) error
}
