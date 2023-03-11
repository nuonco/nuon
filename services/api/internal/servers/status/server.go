package status

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/api/internal"
	connectv1 "github.com/powertoolsdev/mono/pkg/protos/shared/generated/types/status/v1/statusv1connect"
)

type server struct {
	GitRef string `validate:"required"`
}

var _ connectv1.StatusServiceHandler = (*server)(nil)

func NewHandler() (string, http.Handler) {
	srv := &server{}
	return connectv1.NewStatusServiceHandler(srv)
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

// WithConfig is a helper which allows us to encapsulate what fields are needed from the config here, while also not
// _just_ attaching it to the server struct.
func WithConfig(cfg *internal.Config) serverOption {
	return func(s *server) error {
		s.GitRef = cfg.GitRef
		return nil
	}
}

func WithHTTPMux(mux *http.ServeMux) serverOption {
	return func(s *server) error {
		path, handler := connectv1.NewStatusServiceHandler(s)
		mux.Handle(path, handler)
		return nil
	}
}
