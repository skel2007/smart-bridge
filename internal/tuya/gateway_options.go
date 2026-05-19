package tuya

import "log/slog"

type Option func(*gatewayOptions)

type gatewayOptions struct {
	logger             *slog.Logger
	specificationCache bool
}

func WithLogger(logger *slog.Logger) Option {
	return func(options *gatewayOptions) {
		if logger != nil {
			options.logger = logger
		}
	}
}

func WithSpecificationCache() Option {
	return func(options *gatewayOptions) {
		options.specificationCache = true
	}
}

func newGatewayOptions(options ...Option) gatewayOptions {
	var gatewayOptions gatewayOptions
	for _, option := range options {
		option(&gatewayOptions)
	}

	return gatewayOptions
}
