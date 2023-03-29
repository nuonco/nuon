package github

import (
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/go-playground/validator/v10"
	connectv1 "github.com/powertoolsdev/mono/pkg/types/api/github/v1/githubv1connect"
	"github.com/powertoolsdev/mono/services/api/internal/servers"
	"github.com/powertoolsdev/mono/services/api/internal/services"
)

type server struct {
	Svc services.GithubService `validate:"required"`
}

var _ connectv1.GithubServiceHandler = (*server)(nil)

func New(opts ...serverOption) (*server, error) {
	srv := &server{}
	validate := validator.New()

	for idx, opt := range opts {
		if err := opt(srv); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := validate.Struct(srv); err != nil {
		return nil, fmt.Errorf("unable to validate server: %w", err)
	}
	return srv, nil
}

type serverOption func(*server) error

func WithHTTPMux(mux *http.ServeMux) serverOption {
	return func(s *server) error {
		path, handler := connectv1.NewGithubServiceHandler(s,
			connect.WithInterceptors(connect.UnaryInterceptorFunc(servers.MetricsInterceptor)))
		mux.Handle(path, handler)
		return nil
	}
}

func WithService(svc services.GithubService) serverOption {
	return func(s *server) error {
		s.Svc = svc
		return nil
	}
}
