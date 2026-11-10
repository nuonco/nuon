package nuon

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func encodeLabelsQuery(labels map[string]string) *string {
	if len(labels) == 0 {
		return nil
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, labels[k]))
	}

	encoded := strings.Join(parts, ",")
	return &encoded
}

// GetInstallCreateTarget returns the app config, and any app branch and
// group, that a new install with these labels and appBranchID would start
// on. It does not create anything.
func (c *client) GetInstallCreateTarget(ctx context.Context, appID string, labels map[string]string, appBranchID string) (*models.ServiceInstallCreateTargetResponse, error) {
	params := &operations.GetInstallCreateTargetParams{
		Context: ctx,
		AppID:   appID,
		Labels:  encodeLabelsQuery(labels),
	}
	if appBranchID != "" {
		params.AppBranchID = &appBranchID
	}

	resp, err := c.genClient.Operations.GetInstallCreateTarget(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
