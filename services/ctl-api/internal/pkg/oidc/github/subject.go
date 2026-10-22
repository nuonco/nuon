package github

import (
	"strings"
)

const (
	subjectRepoPrefix   = "repo:"
	ownerForbiddenChars = "*?[]{}\\:"
)

type Subject struct {
	Owner   string
	OwnerID string
	Repo    string
	RepoID  string
	Context string
}

func NewSubject(owner, repo, context string) Subject {
	return Subject{Owner: owner, Repo: repo, Context: context}
}

func (s Subject) WithContext(context string) Subject {
	s.Context = context
	return s
}

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

func ParseSubjectOwner(pattern string) (string, bool) {
	rest, ok := strings.CutPrefix(pattern, subjectRepoPrefix)
	if !ok {
		return "", false
	}

	ownerPart, _, ok := strings.Cut(rest, "/")
	if !ok || ownerPart == "" {
		return "", false
	}

	if strings.ContainsAny(ownerPart, ownerForbiddenChars) {
		return "", false
	}

	owner, _ := cutID(ownerPart)

	return owner, owner != ""
}
