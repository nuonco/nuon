package ecr

import (
	"fmt"
	"strings"
)

func parseImageURL(url string) (string, error) {
	// TODO(jm): parse this with actual regex, or something less brittle.
	pieces := strings.SplitN(url, ".dkr.ecr", 3)
	if len(pieces) != 2 {
		return "", fmt.Errorf("invalid ecr image url")
	}

	return pieces[0], nil
}

func TrimRepositoryName(repoName, serverAddress string) (string, error) {
	addrSubs := strings.SplitN(serverAddress, "https://", 2)
	if len(addrSubs) != 2 {
		return "", fmt.Errorf("malformed server address - no https:// prefix")
	}

	prefix := addrSubs[1]
	repoName = strings.TrimPrefix(repoName, prefix)

	repoName = strings.TrimPrefix(repoName, "/")

	return repoName, nil
}
