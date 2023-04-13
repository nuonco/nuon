package status

import (
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/go-playground/validator/v10"
	connectv1 "github.com/powertoolsdev/mono/pkg/types/shared/status/v1/statusv1connect"
)

type server struct {
	v *validator.Validate

	GitRef       string                `validate:"required"`
	Interceptors []connect.Interceptor `validate:"required"`
}

var _ connectv1.StatusServiceHandler = (*server)(nil)

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
		path, handler := connectv1.NewStatusServiceHandler(s, connect.WithInterceptors(s.Interceptors...))
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

func WithGitRef(gitRef string) serverOption {
	return func(s *server) error {
		s.GitRef = gitRef
		return nil
	}
}
