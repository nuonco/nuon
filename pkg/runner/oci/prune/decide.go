package prune

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

const childTagSeparator = "--"

type Manifest struct {
	Digest   string    `json:"digest"`
	Tags     []string  `json:"tags,omitempty"`
	PushedAt time.Time `json:"pushed_at"`
	Children []string  `json:"children,omitempty"`
	Subject  string    `json:"subject,omitempty"`
}

type Input struct {
	Manifests    []Manifest
	OwnedTags    []string
	OwnedPattern []string
	KeepTags     []string
	InUseTags    []string
	InUseDigests []string
	GracePeriod  time.Duration
	Now          time.Time
	MaxDeletes   int
}

type KeepReason string

const (
	KeepReasonUnowned  KeepReason = "unowned_tag"
	KeepReasonKeepTag  KeepReason = "keep_tag"
	KeepReasonInUse    KeepReason = "in_use"
	KeepReasonGrace    KeepReason = "grace_period"
	KeepReasonParent   KeepReason = "referenced_by_kept_index"
	KeepReasonSubject  KeepReason = "attached_to_kept_manifest"
	KeepReasonDanglers KeepReason = "unresolved_reference"
)

type Deletion struct {
	Digest string   `json:"digest"`
	Tags   []string `json:"tags,omitempty"`
}

type Decision struct {
	Batches   [][]Deletion          `json:"batches"`
	Kept      map[string]KeepReason `json:"kept"`
	Truncated bool                  `json:"truncated"`
	Total     int                   `json:"total"`
}

func (d *Decision) Deletions() []Deletion {
	var out []Deletion
	for _, b := range d.Batches {
		out = append(out, b...)
	}
	return out
}

func Decide(in Input) (*Decision, error) {
	ownedSet := toSet(in.OwnedTags)
	patterns := make([]*regexp.Regexp, 0, len(in.OwnedPattern))
	for _, p := range in.OwnedPattern {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("invalid owned tag pattern %q: %w", p, err)
		}
		patterns = append(patterns, re)
	}
	owned := func(tag string) bool {
		if ownedSet[tag] {
			return true
		}
		for _, re := range patterns {
			if re.MatchString(tag) {
				return true
			}
		}
		return false
	}
	keepTags := toSet(in.KeepTags)
	inUseTags := toSet(in.InUseTags)
	inUseDigests := toSet(in.InUseDigests)

	byDigest := make(map[string]Manifest, len(in.Manifests))
	for _, m := range in.Manifests {
		byDigest[m.Digest] = m
	}

	kept := map[string]KeepReason{}
	keep := func(digest string, reason KeepReason) bool {
		if _, ok := kept[digest]; ok {
			return false
		}
		kept[digest] = reason
		return true
	}

	for _, m := range in.Manifests {
		switch {
		case inUseDigests[m.Digest]:
			keep(m.Digest, KeepReasonInUse)
		case in.GracePeriod > 0 && in.Now.Sub(m.PushedAt) < in.GracePeriod:
			keep(m.Digest, KeepReasonGrace)
		}
		for _, tag := range m.Tags {
			base := baseTag(tag)
			switch {
			case !owned(tag) && !owned(base):
				keep(m.Digest, KeepReasonUnowned)
			case keepTags[tag] || keepTags[base]:
				keep(m.Digest, KeepReasonKeepTag)
			case inUseTags[tag] || inUseTags[base]:
				keep(m.Digest, KeepReasonInUse)
			}
		}
		for _, child := range m.Children {
			if _, ok := byDigest[child]; !ok {
				keep(m.Digest, KeepReasonDanglers)
			}
		}
	}

	for changed := true; changed; {
		changed = false
		for _, m := range in.Manifests {
			if _, ok := kept[m.Digest]; ok {
				for _, child := range m.Children {
					if keep(child, KeepReasonParent) {
						changed = true
					}
				}
				continue
			}
			if m.Subject != "" {
				if _, ok := kept[m.Subject]; ok {
					if keep(m.Digest, KeepReasonSubject) {
						changed = true
					}
				}
			}
		}
	}

	parents := map[string][]string{}
	for _, m := range in.Manifests {
		if _, ok := kept[m.Digest]; ok {
			continue
		}
		for _, child := range m.Children {
			parents[child] = append(parents[child], m.Digest)
		}
		if m.Subject != "" {
			parents[m.Subject] = append(parents[m.Subject], m.Digest)
		}
	}

	var candidates []string
	for _, m := range in.Manifests {
		if _, ok := kept[m.Digest]; !ok {
			candidates = append(candidates, m.Digest)
		}
	}
	sort.Strings(candidates)

	decision := &Decision{Kept: kept, Total: len(candidates)}
	done := map[string]bool{}
	remaining := len(candidates)
	budget := in.MaxDeletes

	for remaining > 0 && budget > 0 {
		var batch []Deletion
		for _, digest := range candidates {
			if done[digest] || !allDone(parents[digest], done) {
				continue
			}
			if budget == 0 {
				break
			}
			batch = append(batch, Deletion{Digest: digest, Tags: byDigest[digest].Tags})
			budget--
		}
		if len(batch) == 0 {
			break
		}
		for _, d := range batch {
			done[d.Digest] = true
		}
		remaining -= len(batch)
		decision.Batches = append(decision.Batches, batch)
	}
	decision.Truncated = remaining > 0

	return decision, nil
}

func baseTag(tag string) string {
	if i := strings.Index(tag, childTagSeparator); i > 0 {
		return tag[:i]
	}
	return tag
}

func allDone(digests []string, done map[string]bool) bool {
	for _, d := range digests {
		if !done[d] {
			return false
		}
	}
	return true
}

func toSet(vals []string) map[string]bool {
	out := make(map[string]bool, len(vals))
	for _, v := range vals {
		if v != "" {
			out[v] = true
		}
	}
	return out
}
