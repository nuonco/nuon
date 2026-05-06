package sandboxmode

// AllRunnerJobOperationTypes returns all known runner job operation types.
func AllRunnerJobOperationTypes() []string {
	return []string{
		"exec",
		"build",
		"create-apply-plan",
		"create-teardown-plan",
		"apply-plan",
	}
}

// AllRunnerJobTypes returns all known runner job types.
func AllRunnerJobTypes() []string {
	return []string{
		"terraform-deploy",
		"helm-chart-deploy",
		"kubernetes-manifest-deploy",
		"job-deploy",
		"noop-deploy",

		"docker-build",
		"container-image-build",
		"terraform-module-build",
		"helm-chart-build",
		"kubernetes-manifest-build",
		"noop-build",

		"oci-sync",
		"noop-sync",
		"fetch-image-metadata",

		"actions-workflow",

		"sandbox-terraform",
		"sandbox-sync-secrets",

		"runner-helm",
		"runner-terraform",
	}
}
