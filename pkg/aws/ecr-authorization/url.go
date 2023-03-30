package ecr

import (
	"fmt"
	"strings"
)

// parseImageURL returns the registryID from an oci image url
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
