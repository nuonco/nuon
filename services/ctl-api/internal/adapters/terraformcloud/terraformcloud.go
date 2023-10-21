package terraformcloud

import (
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

func NewTerraformCloud(cfg *internal.Config) (*tfe.Client, error) {
	config := &tfe.Config{
		Token:             cfg.TFEToken,
		RetryServerErrors: true,
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get terraform cloud client: %w", err)
	}

	return client, nil
}
