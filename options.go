package schema

type config struct {
	debug bool
}

type Option func(*config)

// WithDebug sets the debug mode for the schema package.
func WithDebug(debug ...bool) Option {
	return func(c *config) {
		c.debug = optional(true, debug...)
	}
}

func applyOptions(opts ...Option) *config {
	cfg := &config{
		debug: false, // default value
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}
