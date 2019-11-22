package etcd

// Option is a functional option for repository
type Option func(*Repository)

// Options is a slice of Option
type Options []Option

// Apply calls each option on r in turn
func (o Options) Apply(r *Repository) {
	for _, opt := range o {
		opt(r)
	}
}

// ForList constructs a repository client for a specific
// list (other than the default one)
func ForList(name string) Option {
	return func(r *Repository) {
		r.list = name
	}
}

// WithNamespace configures the etcd client to use a particular
// top-level prefix
func WithNamespace(ns string) Option {
	return func(r *Repository) {
		r.namespace = ns
	}
}
