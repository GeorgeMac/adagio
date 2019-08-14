package worker

type Option func(*Pool)

type Options []Option

func (o Options) Apply(p *Pool) {
	for _, opt := range o {
		opt(p)
	}
}

func WithWorkerCount(count int) Option {
	return func(p *Pool) {
		p.size = count
	}
}

func WithClaimerFunc(fn func() Claimer) Option {
	return func(p *Pool) {
		p.newClaimer = fn
	}
}
