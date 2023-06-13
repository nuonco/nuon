package deploy

import (
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/go-playground/validator/v10"
	connectv1 "github.com/powertoolsdev/mono/pkg/types/api/deploy/v1/deployv1connect"
	"github.com/powertoolsdev/mono/services/api/internal/jobs"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"gorm.io/gorm"
)

type server struct {
	v            *validator.Validate
	repo         repos.DeployRepo
	instanceRepo repos.InstanceRepo
	mgr          jobs.Manager

	Interceptors []connect.Interceptor `validate:"required"`
}

var _ connectv1.DeployServiceHandler = (*server)(nil)

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
		path, handler := connectv1.NewDeployServiceHandler(s, connect.WithInterceptors(s.Interceptors...))
		mux.Handle(path, handler)
		return nil
	}
}

func WithDB(db *gorm.DB) serverOption {
	return func(s *server) error {
		s.repo = repos.NewDeployRepo(db)
		s.instanceRepo = repos.NewInstanceRepo(db)
		return nil
	}
}

func WithWorkflowManager(mgr jobs.Manager) serverOption {
	return func(s *server) error {
		s.mgr = mgr
		return nil
	}
}

func WithInterceptors(int ...connect.Interceptor) serverOption {
	return func(s *server) error {
		s.Interceptors = int
		return nil
	}
}
