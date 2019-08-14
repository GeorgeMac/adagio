package worker

import "github.com/georgemac/adagio/pkg/adagio"

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

func ClaimFunc(fn func() *adagio.Claim) Option {
	return func(p *Pool) {
		p.newClaim = fn
	}
}
