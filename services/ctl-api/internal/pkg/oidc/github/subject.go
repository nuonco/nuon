package github

import (
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidc"
)

const subjectRepoPrefix = "repo:"

// Subject is the `sub` claim of a GitHub Actions token:
// `repo:<owner>/<repo>:<context>`. When the repository has immutable
// identifiers enabled, owner and repo carry their numeric IDs as
// `<name>@<id>`.
type Subject struct {
	Owner   string
	OwnerID string
	Repo    string
	RepoID  string
	// Context is everything after the repository, e.g. `ref:refs/heads/main`
	// or `environment:production`.
	Context string
}

// NewSubject builds a subject with mutable (name-only) repository identity.
func NewSubject(owner, repo, context string) Subject {
	return Subject{Owner: owner, Repo: repo, Context: context}
}

// WithContext returns the subject with the given context.
func (s Subject) WithContext(context string) Subject {
	s.Context = context
	return s
}

// WithIDs returns the subject with immutable owner and repository IDs.
func (s Subject) WithIDs(ownerID, repoID string) Subject {
	s.OwnerID = ownerID
	s.RepoID = repoID
	return s
}

func (s Subject) String() string {
	var b strings.Builder

	b.WriteString(subjectRepoPrefix)
	b.WriteString(withID(s.Owner, s.OwnerID))
	b.WriteString("/")
	b.WriteString(withID(s.Repo, s.RepoID))
	if s.Context != "" {
		b.WriteString(":")
		b.WriteString(s.Context)
	}

	return b.String()
}

func withID(name, id string) string {
	if id == "" {
		return name
	}

	return name + "@" + id
}

func cutID(part string) (name, id string) {
	name, id, _ = strings.Cut(part, "@")
	return name, id
}

// ParseSubject parses a concrete `sub` claim value.
func ParseSubject(sub string) (Subject, bool) {
	rest, ok := strings.CutPrefix(sub, subjectRepoPrefix)
	if !ok {
		return Subject{}, false
	}

	ownerPart, rest, ok := strings.Cut(rest, "/")
	if !ok || ownerPart == "" {
		return Subject{}, false
	}

	repoPart, context, _ := strings.Cut(rest, ":")
	if repoPart == "" {
		return Subject{}, false
	}

	var s Subject
	s.Owner, s.OwnerID = cutID(ownerPart)
	s.Repo, s.RepoID = cutID(repoPart)
	s.Context = context

	return s, s.Owner != "" && s.Repo != ""
}

// ParseSubjectOwner returns the literal owner a `sub` pattern is pinned to,
// for patterns like `repo:acme/app:*` or `repo:acme@123/app@456:*`. It
// reports false when the pattern does not start with `repo:<owner>/` or the
// owner is empty or contains glob metacharacters.
func ParseSubjectOwner(pattern string) (string, bool) {
	rest, ok := strings.CutPrefix(pattern, subjectRepoPrefix)
	if !ok {
		return "", false
	}

	ownerPart, _, ok := strings.Cut(rest, "/")
	if !ok || ownerPart == "" {
		return "", false
	}

	if oidc.IsGlob(ownerPart) || strings.Contains(ownerPart, ":") {
		return "", false
	}

	owner, _ := cutID(ownerPart)

	return owner, owner != ""
}
