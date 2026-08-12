package lookup

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// SecretID resolves a secret ID or name to the secret's ID. The API has no
// single-get endpoint for secrets and its delete is idempotent, so without
// this resolution a delete by name (or any bad ID) reports success while
// deleting nothing.
func SecretID(ctx context.Context, apiClient nuon.Client, appID, secretIDOrName string) (string, error) {
	const limit = 100
	offset := 0

	for {
		secrets, hasMore, err := apiClient.GetAppSecrets(ctx, appID, &models.GetPaginatedQuery{
			Offset: offset,
			Limit:  limit,
		})
		if err != nil {
			return "", &ui.CLIUserError{
				Msg: "unable to list app secrets",
			}
		}

		for _, secret := range secrets {
			if secret.ID == secretIDOrName || secret.Name == secretIDOrName {
				return secret.ID, nil
			}
		}

		if !hasMore {
			return "", &ui.CLIUserError{
				Msg: fmt.Sprintf("secret id or name is not valid: %s", secretIDOrName),
			}
		}
		offset += limit
	}
}
