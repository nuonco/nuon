package helpers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v50/github"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"golang.org/x/oauth2"
)

func (H *Helpers) GetVCSConnectionClient(ctx context.Context, vcsConn *app.VCSConnection) (*github.Client, error) {
	installID, err := strconv.ParseInt(vcsConn.GithubInstallID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to get install ID: %w", err)
	}

	resp, _, err := H.ghClient.Apps.CreateInstallationToken(ctx, installID, &github.InstallationTokenOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get installation token: %w", err)
	}

	// get a client with the github install token
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *resp.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return client, nil
}
