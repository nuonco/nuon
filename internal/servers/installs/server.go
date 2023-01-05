package installs

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	connectv1 "github.com/powertoolsdev/protos/api/generated/types/install/v1/installv1connect"
)

type server struct {
}

var _ connectv1.InstallsServiceHandler = (*server)(nil)

func NewHandler() (string, http.Handler) {
	srv := &server{}
	return connectv1.NewInstallsServiceHandler(srv)
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
		path, handler := connectv1.NewInstallsServiceHandler(s)
		mux.Handle(path, handler)
		return nil
	}
}
