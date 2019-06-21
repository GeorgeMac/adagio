package worker

type Option func(*Pool)

type Options []Option

func (o Options) Apply(p *Pool) {
	for _, opt := range o {
		opt(p)
	}
}

func WorkerCount(count int) Option {
	return func(p *Pool) {
		p.size = count
	}
}
