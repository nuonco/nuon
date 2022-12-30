package orgs

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/orgs-api/internal/orgcontext"
	orgsservice "github.com/powertoolsdev/orgs-api/internal/services/orgs"
	connectv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1/orgsv1connect"
)

type server struct {
	CtxProvider orgcontext.Provider `validate:"required"`
	Svc         orgsservice.Service `validate:"required"`
}

var _ connectv1.OrgsServiceHandler = (*server)(nil)

func NewHandler() (string, http.Handler) {
	srv := &server{}
	return connectv1.NewOrgsServiceHandler(srv)
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

// WithContextProvider sets the context provider on the object
func WithContextProvider(ctxProvider orgcontext.Provider) serverOption {
	return func(s *server) error {
		s.CtxProvider = ctxProvider
		return nil
	}
}

// WithService sets the provided service object
func WithService(svc orgsservice.Service) serverOption {
	return func(s *server) error {
		s.Svc = svc
		return nil
	}
}
