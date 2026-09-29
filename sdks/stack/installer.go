package stack

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/nuonco/nuon/sdks/stack/models"
)

type Options struct {
	APIURL string

	InstallID string

	APIToken string

	OrgID string

	HTTPClient *http.Client
}

func (o Options) validate() error {
	if strings.TrimSpace(o.APIURL) == "" {
		return fmt.Errorf("api_url is required")
	}
	if strings.TrimSpace(o.InstallID) == "" {
		return fmt.Errorf("install_id is required")
	}

	return nil
}

func FetchConfig(ctx context.Context, opts Options) (*models.AppInstallerSDKConfig, error) {
	c, err := newClient(ctx, opts)
	if err != nil {
		return nil, err
	}

	return c.FetchConfig(ctx)
}
