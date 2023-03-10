package config

import (
	"net/http"
)

type options struct {
	httpServer *http.Server
}

// Option is a configuration option for the metrics system
type Option func(*options)

// WithHTTPServer passes the http.Server to bind to
func WithHTTPServer(httpServer *http.Server) Option {
	return func(options *options) {
		options.httpServer = httpServer
	}
}
