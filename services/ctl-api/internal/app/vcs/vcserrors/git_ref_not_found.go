package vcserrors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/v50/github"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const (
	GitRefNotFoundType         compositeerrors.Type = "vcs.git_ref_not_found"
	GitRefNotFoundTemporalType                      = "vcs_git_ref_not_found"
)

type GitRefNotFoundError struct {
	Repo string `json:"repo,omitempty"`
	Ref  string `json:"ref,omitempty"`
}

var (
	_ compositeerrors.CompositeError = (*GitRefNotFoundError)(nil)
	_ compositeerrors.HintsProvider  = (*GitRefNotFoundError)(nil)
)

func (e *GitRefNotFoundError) Error() string {
	return GitRefNotFoundMessage(e.Repo, e.Ref)
}

func (e *GitRefNotFoundError) Type() compositeerrors.Type {
	return GitRefNotFoundType
}

func (e *GitRefNotFoundError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityFatal
}

func (e *GitRefNotFoundError) Sections() []compositeerrors.Section {
	return []compositeerrors.Section{
		compositeerrors.MarkdownSection("Why", "The configured git ref could not be read from the repository. It may have been renamed or deleted, or the connection may not have access to it."),
		compositeerrors.MarkdownSection("How to fix", "Point the config at a ref that exists in the repository, or grant the Nuon GitHub app access to it, then run again."),
	}
}

func (e *GitRefNotFoundError) Hints() compositeerrors.Hints {
	return compositeerrors.NewHints().WithTerminal()
}

func GitRefNotFoundMessage(repo, ref string) string {
	switch {
	case ref == "" && repo == "":
		return "The configured git ref was not found"
	case ref == "":
		return fmt.Sprintf("The configured git ref was not found in %q", repo)
	case repo == "":
		return fmt.Sprintf("Git ref %q was not found", ref)
	default:
		return fmt.Sprintf("Git ref %q was not found in %q", ref, repo)
	}
}

func NewGitRefNotFound(repo, ref string, cause error) error {
	return temporal.NewNonRetryableApplicationError(
		GitRefNotFoundMessage(repo, ref),
		GitRefNotFoundTemporalType,
		cause,
	)
}

func IsGitRefNotFound(err error) bool {
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		switch appErr.Type() {
		case GitRefNotFoundTemporalType,
			githubErrorType(http.StatusNotFound),
			githubErrorType(http.StatusUnprocessableEntity):
			return true
		}
	}

	var ghErr *github.ErrorResponse
	if errors.As(err, &ghErr) && ghErr.Response != nil {
		return ghErr.Response.StatusCode == http.StatusNotFound ||
			ghErr.Response.StatusCode == http.StatusUnprocessableEntity
	}

	return false
}

func NonRetryableGitHubError(err error) error {
	var ghErr *github.ErrorResponse
	if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode >= 400 && ghErr.Response.StatusCode < 500 {
		return temporal.NewNonRetryableApplicationError(
			ghErr.Message,
			githubErrorType(ghErr.Response.StatusCode),
			err,
		)
	}
	return nil
}

func githubErrorType(status int) string {
	return fmt.Sprintf("github_%d", status)
}
