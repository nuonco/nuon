package instances

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/orgs-api/internal/servers"
	connectv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/instances/v1/instancesv1connect"
)

type server struct {
	*servers.Base
}

var _ connectv1.InstancesServiceHandler = (*server)(nil)

func NewHandler(v *validator.Validate, opts ...servers.BaseOption) (string, http.Handler, error) {
	baseSrv, err := servers.New(v, opts...)
	if err != nil {
		return "", nil, fmt.Errorf("invalid base server: %w", err)
	}

	path, handler := connectv1.NewInstancesServiceHandler(&server{
		Base: baseSrv,
	})
	return path, handler, nil
}
