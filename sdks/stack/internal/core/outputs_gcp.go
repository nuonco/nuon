package core

// GCPOutputs is the GCP-specific result of a successful provision. The field
// set mirrors install-stacks/gcp/outputs.tf and phone_home.tf (phone_home_type
// "gcp") so app templates resolving `nuon.install_stack.outputs.*` see the
// module's keys. Service-account values are fully resolved (emails + numeric
// uniqueIds).
type GCPOutputs struct {
	ProjectID string
	Region    string

	NetworkName string
	NetworkID   string

	PublicSubnetName  string
	PrivateSubnetName string
	RunnerSubnetName  string

	RunnerSAEmail    string
	RunnerSAUniqueID string

	GKENodePoolSAEmail    string
	GKENodePoolSAUniqueID string

	ProvisionSAEmail      string
	ProvisionSAUniqueID   string
	MaintenanceSAEmail    string
	MaintenanceSAUniqueID string
	DeprovisionSAEmail    string
	DeprovisionSAUniqueID string

	// Keyed by role name.
	BreakGlassSAEmails    map[string]string
	BreakGlassSAUniqueIDs map[string]string
	CustomSAEmails        map[string]string
	CustomSAUniqueIDs     map[string]string

	// SecretNames maps `<name>_secret_name` to the fully-qualified Secret
	// Manager resource name.
	SecretNames map[string]string
}
