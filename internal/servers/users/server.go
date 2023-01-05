package users

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/api/internal"
	connectv1 "github.com/powertoolsdev/protos/api/generated/types/user/v1/userv1connect"
)

type server struct {
}

var _ connectv1.UsersServiceHandler = (*server)(nil)

func NewHandler() (string, http.Handler) {
	srv := &server{}
	return connectv1.NewUsersServiceHandler(srv)
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
		return nil
	}
}
