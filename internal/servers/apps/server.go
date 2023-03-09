package apps

import (
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/orgs-api/internal/servers"
	connectv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/apps/v1/appsv1connect"
)

type server struct {
	*servers.Base
}

var _ connectv1.AppsServiceHandler = (*server)(nil)

func NewHandler(v *validator.Validate, opts ...servers.BaseOption) (string, http.Handler, error) {
	baseSrv, err := servers.New(v, opts...)
	if err != nil {
		return "", nil, fmt.Errorf("invalid base server: %w", err)
	}

	path, handler := connectv1.NewAppsServiceHandler(&server{
		Base: baseSrv,
	}, connect.WithInterceptors(connect.UnaryInterceptorFunc(servers.EnsureShortIDInterceptor)))
	return path, handler, nil
}
