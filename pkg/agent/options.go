package agent

// Option is a functional option for the Pool type
type Option func(*Pool)

// Options is a slice of Option types
type Options []Option

// Apply calls each option in turn on the provided Pool
func (o Options) Apply(p *Pool) {
	for _, opt := range o {
		opt(p)
	}
}

// WithAgentCount configures the number of agents to be run
func WithAgentCount(count int) Option {
	return func(p *Pool) {
		p.size = count
	}
}

// WithClaimerFunc overrides the claimer which generates a unique
// claim per node claim attempt
func WithClaimerFunc(fn func() Claimer) Option {
	return func(p *Pool) {
		p.newClaimer = fn
	}
}
