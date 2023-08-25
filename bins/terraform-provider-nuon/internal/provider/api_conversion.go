package provider

const (
	// orgs, apps & installs:
	statusQueued         = "queued"
	statusProvisioning   = "provisioning"
	statusActive         = "active"
	statusDeleteQueued   = "delete_queued"
	statusDeprovisioning = "deprovisioning"

	// builds:
	// queued
	statusPlanning = "planning"
	statusBuilding = "building"
	// active

	// deploys:
	// queued
	// planning
	statusDeploying = "deploying"
	// active
)
