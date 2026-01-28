// Package dockerhub provides utilities for working with Docker Hub references.
package dockerhub

import "strings"

// IsDockerHubRegistry returns true if the host is a Docker Hub registry.
func IsDockerHubRegistry(host string) bool {
	switch host {
	case "docker.io", "registry-1.docker.io", "index.docker.io":
		return true
	default:
		return false
	}
}

// NormalizeReference normalizes Docker Hub references to fully qualified form.
// - "nginx" -> "docker.io/library/nginx" (official images)
// - "username/image" -> "docker.io/username/image" (user images)
// - "docker.io/nginx" -> "docker.io/library/nginx" (official images with explicit host)
// - "registry-1.docker.io/nginx" -> "registry-1.docker.io/library/nginx"
// - Non-Docker Hub refs pass through unchanged
func NormalizeReference(ref string) string {
	ref = strings.TrimPrefix(ref, "https://")
	ref = strings.TrimPrefix(ref, "http://")

	parts := strings.Split(ref, "/")

	// Single name like "nginx" -> official image
	if len(parts) == 1 {
		return "docker.io/library/" + parts[0]
	}

	// Two parts: could be "user/image" or "registry.io/image"
	if len(parts) == 2 {
		first := parts[0]
		// Check if first part looks like a registry host
		if strings.Contains(first, ".") || strings.Contains(first, ":") || first == "localhost" {
			// It's "registry/image" - check if Docker Hub
			if IsDockerHubRegistry(first) {
				// Docker Hub official image needs library/ prefix
				return first + "/library/" + parts[1]
			}
			// Non-Docker Hub registry, return as-is
			return ref
		}
		// No registry, it's "user/image" -> Docker Hub user image
		return "docker.io/" + ref
	}

	// Three or more parts: "registry/namespace/image" or "registry/library/image"
	if len(parts) >= 3 {
		host := parts[0]
		// For Docker Hub, ensure we don't double-add library/
		if IsDockerHubRegistry(host) {
			// Already has namespace (e.g., library/nginx or user/nginx)
			return ref
		}
	}

	return ref
}
