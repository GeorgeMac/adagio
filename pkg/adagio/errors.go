package adagio

import "errors"

var (
	// ErrMissingNode is returned when a node is not found
	ErrMissingNode = errors.New("node not found")
	// ErrNodeNotReady is returned when a claim is made on a node that is not ready
	ErrNodeNotReady = errors.New("node not ready")
	// ErrRunDoesNotExist is returned when a run is referenced which does not exist
	ErrRunDoesNotExist = errors.New("run does not exist")
)
