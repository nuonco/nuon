package config

import (
	"net/http"
)

type options struct {
	httpServer *http.Server
}

type Option func(*options)

func WithHTTPServer(httpServer *http.Server) Option {
	return func(options *options) {
		options.httpServer = httpServer
	}
}
