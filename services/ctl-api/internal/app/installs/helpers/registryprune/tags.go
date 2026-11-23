package registryprune

import "strings"

func ImageNameSegment(componentName string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(componentName) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case !lastDash:
			b.WriteRune('-')
			lastDash = true
		}
	}
	segment := strings.Trim(b.String(), "-")
	if segment == "" {
		return "app"
	}
	return segment
}

func InstallRegistryTag(componentName, buildID, sourceDigest, resolvedTag string, sharedRepo bool) string {
	if sourceDigest == "" {
		return buildID
	}
	if resolvedTag == "" {
		return buildID
	}
	if sharedRepo {
		return ImageNameSegment(componentName) + "-" + resolvedTag
	}
	return resolvedTag
}
