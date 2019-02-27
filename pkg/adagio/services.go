package adagio

type Worker interface {
	Execute(Node) error
}

type ControlPlane interface {
	Execute(Graph) (Run, error)
	ListRuns() ([]Run, error)
}
