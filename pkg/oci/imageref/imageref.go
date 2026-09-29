package imageref

import (
	"fmt"
	"strings"
)

type Spec struct {
	Image        string
	Tag          string
	UpdatePolicy string
}

func parseSpec(image, tag string) (repo, literalTag, digest string) {
	repo = image
	switch {
	case strings.Contains(tag, "@sha256:"):
		parts := strings.SplitN(tag, "@sha256:", 2)
		if image == "" {
			repo = parts[0]
		}
		digest = "sha256:" + parts[1]
	case strings.HasPrefix(tag, "sha256:"):
		digest = tag
	case tag != "":
		literalTag = tag
	case strings.Contains(image, "@sha256:"):
		parts := strings.SplitN(image, "@sha256:", 2)
		repo = parts[0]
		digest = "sha256:" + parts[1]
	}
	return repo, literalTag, digest
}

func (sp Spec) PullRef(selectedTag string) string {
	if sp.UpdatePolicy != "" {
		return selectedTag
	}
	_, literalTag, digest := parseSpec(sp.Image, sp.Tag)
	if digest != "" {
		return digest
	}
	return literalTag
}

func (sp Spec) Identity(selectedTag string) Source {
	repo, literalTag, digest := parseSpec(sp.Image, sp.Tag)
	s := Source{SourceImage: repo}
	switch {
	case sp.UpdatePolicy != "":
		s.SourceRef = fmt.Sprintf("%s:%s", repo, sp.UpdatePolicy)
		s.ResolvedTag = selectedTag
	case digest != "":
		s.SourceRef = fmt.Sprintf("%s@%s", repo, digest)
	default:
		s.SourceRef = fmt.Sprintf("%s:%s", repo, literalTag)
		s.ResolvedTag = literalTag
	}
	return s
}

const shortDigestLen = 7

type Source struct {
	SourceImage  string
	SourceRef    string
	ResolvedTag  string
	SourceDigest string
}

func ImageRef(s Source) string {
	if s.SourceDigest == "" || s.SourceImage == "" {
		return ""
	}
	return fmt.Sprintf("%s@%s", s.SourceImage, s.SourceDigest)
}

func DisplayRef(s Source) string {
	tag := s.ResolvedTag
	if tag == "" {
		tag = parseTagFromRef(s.SourceRef)
	}

	short := shortDigest(s.SourceDigest)

	switch {
	case s.SourceImage != "" && tag != "" && short != "":
		return fmt.Sprintf("%s:%s (sha256:%s)", s.SourceImage, tag, short)
	case s.SourceImage != "" && tag != "":
		return fmt.Sprintf("%s:%s", s.SourceImage, tag)
	case s.SourceImage != "" && short != "":
		return fmt.Sprintf("%s@sha256:%s", s.SourceImage, short)
	case s.SourceRef != "":
		return s.SourceRef
	default:
		return s.SourceImage
	}
}

func parseTagFromRef(ref string) string {
	if ref == "" {
		return ""
	}
	if at := strings.Index(ref, "@"); at != -1 {
		ref = ref[:at]
	}
	// why: The tag is everything after the LAST colon, but only if that colon
	// comes after the last "/" (otherwise it's a registry port like
	// "ghcr.io:5000").
	colon := strings.LastIndex(ref, ":")
	if colon == -1 {
		return ""
	}
	if slash := strings.LastIndex(ref, "/"); slash != -1 && slash > colon {
		return ""
	}
	return ref[colon+1:]
}

func shortDigest(digest string) string {
	const prefix = "sha256:"
	if !strings.HasPrefix(digest, prefix) {
		return ""
	}
	hex := digest[len(prefix):]
	if len(hex) < shortDigestLen {
		return hex
	}
	return hex[:shortDigestLen]
}
