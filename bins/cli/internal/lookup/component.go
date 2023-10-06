package lookup

import (
	"context"

	"github.com/nuonco/nuon-go"
)

func ComponentID(ctx context.Context, apiClient nuon.Client, compIDOrName string) (string, error) {
	comp, err := apiClient.GetComponent(ctx, compIDOrName)
	if err != nil {
		return "", err
	}

	return comp.ID, nil
}
