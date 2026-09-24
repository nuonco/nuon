package lookup

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

func AppBranchID(ctx context.Context, apiClient nuon.Client, appID, branchNameOrID string) (string, error) {
	if branchNameOrID == "" {
		return "", &ui.CLIUserError{
			Msg: "branch is not set, pass the --branch-id flag",
		}
	}

	branches, err := apiClient.GetAppBranches(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("unable to list app branches: %w", err)
	}

	for _, b := range branches {
		if b.ID == branchNameOrID || b.Name == branchNameOrID {
			return b.ID, nil
		}
	}

	return "", &ui.CLIUserError{
		Msg: fmt.Sprintf("branch \"%s\" not found", branchNameOrID),
	}
}
