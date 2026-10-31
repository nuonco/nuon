package helpers

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/google/go-github/v50/github"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) GetSourceFile(ctx context.Context, orgID, repository, commit, filePath string, connection *app.VCSConnection) ([]byte, error) {
	if !sourceidentity.IsCommit(commit) || path.IsAbs(filePath) || path.Clean(filePath) != filePath || filePath == ".." || strings.HasPrefix(filePath, "../") {
		return nil, fmt.Errorf("invalid pinned source path")
	}
	owner, repo, err := parseOwnerRepo(repository)
	if err != nil {
		return nil, err
	}
	var client *github.Client
	if connection != nil {
		if connection.OrgID != orgID {
			return nil, fmt.Errorf("source connection does not belong to org")
		}
		client, err = h.GetVCSConnectionClient(ctx, connection)
	} else {
		client, _, err = h.ResolvePublicRepoGithubClient(ctx, h.l, orgID, owner)
	}
	if err != nil {
		return nil, err
	}
	file, _, _, err := client.Repositories.GetContents(ctx, owner, repo, filePath, &github.RepositoryContentGetOptions{Ref: commit})
	if err != nil {
		return nil, fmt.Errorf("unable to read pinned source file")
	}
	if file == nil || file.GetType() != "file" {
		return nil, fmt.Errorf("source is not a regular file")
	}
	if file.GetSize() > 512*1024 {
		return nil, fmt.Errorf("file exceeds the 512 KiB diff limit")
	}
	if file.GetEncoding() != "base64" {
		return nil, fmt.Errorf("source file content is unavailable")
	}
	contents, err := file.GetContent()
	if err != nil {
		return nil, fmt.Errorf("unable to decode source file")
	}
	return []byte(contents), nil
}
