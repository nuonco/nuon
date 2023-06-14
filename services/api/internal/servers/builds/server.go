package builds

import (
	"fmt"
	"net/http"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/bufbuild/connect-go"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/pkg/types/api/build/v1/buildv1connect"
	"github.com/powertoolsdev/mono/services/api/internal"
	"github.com/powertoolsdev/mono/services/api/internal/clients/github"
	"gorm.io/gorm"
)

type server struct {
	v               *validator.Validate
	temporalClient  temporal.Client
	githubTransport *ghinstallation.AppsTransport
	db              *gorm.DB

	Interceptors []connect.Interceptor `validate:"required"`
}

var _ buildv1connect.BuildsServiceHandler = (*server)(nil)

func New(v *validator.Validate, opts ...serverOption) (*server, error) {
	srv := &server{
		v:            v,
		Interceptors: make([]connect.Interceptor, 0),
	}

	for idx, opt := range opts {
		if err := opt(srv); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := srv.v.Struct(srv); err != nil {
		return nil, fmt.Errorf("unable to validate server: %w", err)
	}
	return srv, nil
}

type serverOption func(*server) error

func WithHTTPMux(mux *http.ServeMux) serverOption {
	return func(s *server) error {
		path, handler := buildv1connect.NewBuildsServiceHandler(s, connect.WithInterceptors(s.Interceptors...))
		mux.Handle(path, handler)
		return nil
	}
}

func WithInterceptors(int ...connect.Interceptor) serverOption {
	return func(s *server) error {
		s.Interceptors = int
		return nil
	}
}

func WithTemporalClient(temporalClient temporal.Client) serverOption {
	return func(s *server) error {
		s.temporalClient = temporalClient
		return nil
	}
}

func WithGithubClient(config *internal.Config) serverOption {
	return func(s *server) error {
		transport, err := github.New(github.WithConfig(config))
		if err != nil {
			return fmt.Errorf("unable to github client: %w", err)
		}
		s.githubTransport = transport

		return nil
	}
}

func WithDBClient(db *gorm.DB) serverOption {
	return func(s *server) error {
		s.db = db
		return nil
	}
}
