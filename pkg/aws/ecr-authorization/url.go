package ecr

import (
	"fmt"
	"strings"
)

// parseImageURL returns the registryID from an oci image url or repository url
//
// eg: the following string will return  766121324316
// 766121324316.dkr.ecr.us-west-2.amazonaws.com/<repo-name>:latest
func parseImageURL(url string) (string, error) {
	// TODO(jm): parse this with actual regex, or something less brittle.
	pieces := strings.SplitN(url, ".dkr.ecr", 3)
	if len(pieces) != 2 {
		return "", fmt.Errorf("invalid ecr image url")
	}

	return pieces[0], nil
}

// TrimRepositoryName trims off the serverAddress from a repo name, so when something like a full URI (with accountID)
// is passed in, we can use the same code paths for it.
func TrimRepositoryName(repoName, serverAddress string) (string, error) {
	addrSubs := strings.SplitN(serverAddress, "https://", 2)
	if len(addrSubs) != 2 {
		return "", fmt.Errorf("malformed server address - no https:// prefix")
	}

	// trim the account prefix
	prefix := addrSubs[1]
	repoName = strings.TrimPrefix(repoName, prefix)

	// there can be a leading `/`, which we trim
	repoName = strings.TrimPrefix(repoName, "/")

	return repoName, nil
}
