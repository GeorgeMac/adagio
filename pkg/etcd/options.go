package etcd

type Option func(*Repository)

type Options []Option

func (o Options) Apply(r *Repository) {
	for _, opt := range o {
		opt(r)
	}
}

func WithNamespace(ns string) Option {
	return func(r *Repository) {
		r.ns = namespace(ns)
	}
}
