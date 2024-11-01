package local

import (
	"fmt"
)

// When we are building and pushing images locally, we have to do things _slightly_ different, to ensure we interoperate
// with the docker vm correctly.
//
// When using `nctl run-local`, we need to tag and push the image as `host.containers.internal`, because the docker
// build is happening _inside_ the vm. This will push it to our local registry running on port 5001 (ultimately, on the
// host)
//
// When we copy the image from the local registry to ECR, we can always access it as localhost:5001/runner.
//
// For running builds inside kaniko, we can simply use localhost:5001 for everything.

// Return a tag that can be used to build+push from the run-local environment
func GetLocalTag(version string) string {
	return fmt.Sprintf("host.containers.internal:5001/runner:%s", version)
}

// Return a tag that can be used inside kaniko
func GetKanikoTag(version string) string {
	return fmt.Sprintf("localhost:5001/runner:%s", version)
}

// Return the tag we _always_ use for copying
func GetCopyTag(version string) string {
	return fmt.Sprintf("localhost:5001/runner:%s", version)
}

// Return the repo we _always_ use for copying
func GetCopyRepo() string {
	return "localhost:5001/runner"
}
