package admin

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/powertoolsdev/mono/services/api/internal/services"
	connectv1 "github.com/powertoolsdev/mono/pkg/protos/api/generated/types/admin/v1/adminv1connect"
)

type server struct {
	Svc services.AdminService
}

var _ connectv1.AdminServiceHandler = (*server)(nil)

func NewHandler() (string, http.Handler) {
	srv := &server{}
	return connectv1.NewAdminServiceHandler(srv)
}

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
		path, handler := connectv1.NewAdminServiceHandler(s)
		mux.Handle(path, handler)
		return nil
	}
}

func WithService(svc services.AdminService) serverOption {
	return func(s *server) error {
		s.Svc = svc
		return nil
	}
}
