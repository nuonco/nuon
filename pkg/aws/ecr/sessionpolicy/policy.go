package sessionpolicy

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var (
	repositoryURIPattern = regexp.MustCompile(`^(\d{12})\.dkr\.ecr\.([a-z0-9-]+)\.amazonaws\.com/([a-z0-9]+(?:[._/-][a-z0-9]+)*)$`)
	regionPattern        = regexp.MustCompile(`^[a-z0-9-]+$`)
	registryIDPattern    = regexp.MustCompile(`^\d{12}$`)
	namePattern          = regexp.MustCompile(`^[a-z0-9]+(?:[._/-][a-z0-9]+)*$`)
)

var pullActions = []string{
	"ecr:BatchCheckLayerAvailability",
	"ecr:BatchGetImage",
	"ecr:GetDownloadUrlForLayer",
}

var pushActions = []string{
	"ecr:InitiateLayerUpload",
	"ecr:UploadLayerPart",
	"ecr:CompleteLayerUpload",
	"ecr:PutImage",
}

// RepositoryARN parses an ECR repository URI
// (111122223333.dkr.ecr.us-west-2.amazonaws.com/org/app) into its repository ARN.
func RepositoryARN(repositoryURI string) (string, error) {
	match := repositoryURIPattern.FindStringSubmatch(repositoryURI)
	if match == nil {
		return "", fmt.Errorf("unable to parse ecr repository uri %q", repositoryURI)
	}
	return repositoryARN(match[2], match[1], match[3])
}

// RepositoryARNFromParts builds the repository ARN for a name in a known registry.
func RepositoryARNFromParts(region, registryID, name string) (string, error) {
	if !regionPattern.MatchString(region) || !registryIDPattern.MatchString(registryID) || !namePattern.MatchString(name) {
		return "", fmt.Errorf("invalid ecr repository arn parts")
	}
	return repositoryARN(region, registryID, name)
}

func repositoryARN(region, registryID, name string) (string, error) {
	return fmt.Sprintf("arn:aws:ecr:%s:%s:repository/%s", region, registryID, name), nil
}

// Document is an IAM session policy that can mint an ECR token and use one
// repository. push adds the layer-upload actions the runner needs when this
// same token is the destination of an image copy.
func Document(repositoryARN string, push bool) (string, error) {
	if repositoryARN == "" || repositoryARN == "*" {
		return "", fmt.Errorf("ecr session policy requires one repository arn")
	}
	actions := append([]string{}, pullActions...)
	if push {
		actions = append(actions, pushActions...)
	}
	doc := struct {
		Version   string `json:"Version"`
		Statement []struct {
			Effect   string `json:"Effect"`
			Action   any    `json:"Action"`
			Resource string `json:"Resource"`
		} `json:"Statement"`
	}{
		Version: "2012-10-17",
		Statement: []struct {
			Effect   string `json:"Effect"`
			Action   any    `json:"Action"`
			Resource string `json:"Resource"`
		}{
			{Effect: "Allow", Action: "ecr:GetAuthorizationToken", Resource: "*"},
			{Effect: "Allow", Action: actions, Resource: repositoryARN},
		},
	}
	byts, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	return string(byts), nil
}
