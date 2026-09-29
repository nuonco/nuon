package dockerhub

import "strings"

func IsDockerHubRegistry(host string) bool {
	switch host {
	case "docker.io", "registry-1.docker.io", "index.docker.io":
		return true
	default:
		return false
	}
}

func NormalizeReference(ref string) string {
	ref = strings.TrimPrefix(ref, "https://")
	ref = strings.TrimPrefix(ref, "http://")

	parts := strings.Split(ref, "/")

	if len(parts) == 1 {
		return "docker.io/library/" + parts[0]
	}

	if len(parts) == 2 {
		first := parts[0]
		if strings.Contains(first, ".") || strings.Contains(first, ":") || first == "localhost" {
			if IsDockerHubRegistry(first) {
				return first + "/library/" + parts[1]
			}
			return ref
		}
		return "docker.io/" + ref
	}

	if len(parts) >= 3 {
		host := parts[0]
		if IsDockerHubRegistry(host) {
			return ref
		}
	}

	return ref
}
