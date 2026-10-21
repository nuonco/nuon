package github

import (
	"errors"
	"fmt"
	"sort"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc"
)

// IssuerURL is GitHub's multi-tenant Actions OIDC issuer: any GitHub user can
// mint tokens from it.
const IssuerURL = "https://token.actions.githubusercontent.com"

var errOwnerNotPinned = errors.New("GitHub Actions tokens can come from any account, so the `sub` must pin an owner: " +
	"start it with a literal owner like `repo:acme/`, or add a numeric `repository_owner_id` condition")

var errOwnerIDNotNumber = errors.New(`claim "repository_owner_id" must be a number`)

var claimNames = oidc.ClaimNames(Claims{})

// Provider implements oidc.Provider for GitHub Actions.
type Provider struct{}

var _ oidc.Provider = Provider{}

func (Provider) IssuerURL() string { return IssuerURL }

// Probes returns repository subjects of an owner no policy should match.
func (Provider) Probes() []string {
	repo := NewSubject(oidc.ProbeA, oidc.ProbeB, "")

	return []string{
		repo.WithContext("ref:refs/heads/main").String(),
		repo.WithContext("environment:production").String(),
		repo.WithContext("pull_request").String(),
		repo.WithIDs("1", "2").WithContext("ref:refs/heads/main").String(),
	}
}

// ValidateConditions rejects claims GitHub does not issue and requires the
// policy to be pinned to an owner, either by a literal owner in `sub` or a
// numeric `repository_owner_id` condition.
func (Provider) ValidateConditions(v *validator.Validate, c oidc.Conditions) error {
	names := make([]string, 0, len(c.Extra))
	for name := range c.Extra {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if _, ok := claimNames[name]; !ok {
			return fmt.Errorf("claim %q is not a GitHub Actions token claim", name)
		}
	}

	ownerID, hasOwnerID := c.Extra["repository_owner_id"]
	if hasOwnerID && v.Var(ownerID, "number") != nil {
		return errOwnerIDNotNumber
	}

	if _, ok := ParseSubjectOwner(c.Sub); ok || hasOwnerID {
		return nil
	}

	return errOwnerNotPinned
}
