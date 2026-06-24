package stack

// Options configures a stack Installer. Use FromURL when you have a
// create-run URL from the dashboard; use New when you already know the
// install ID and region (e.g. for offline Status inspection).
type Options struct {
	InstallID string
	AWSRegion string

	// GCP carries the customer-supplied GCP location/sizing inputs (project,
	// region, machine type, GKE node-pool). These are NOT part of the ctl-api
	// Config — the caller supplies them, mirroring AWSRegion. Ignored unless
	// the run targets GCP.
	GCP GCPOptions

	// Cloud overrides the target cloud provider. When empty, the cloud from
	// the ctl-api Config is used, falling back to the default (aws). Lets
	// stack-cli force a cloud from the CLI.
	Cloud Cloud

	// Method overrides the provisioning method. When empty, the method from
	// the ctl-api Config is used, falling back to the per-cloud default
	// (aws→sdk, others→terraform). Lets stack-cli force a method from the CLI.
	Method Method

	// logStream / stackRun are populated by FromURL. They aren't part of the
	// public construction API because callers using New only need the
	// install ID + region; the URL flow is the only one that needs ctl-api
	// wiring, and it fills these in itself.
	logStream *logStreamConfig
	stackRun  *stackRunConfig
}

// URLOptions is the single-argument bootstrap shape: the URL of the
// create-run POST endpoint that the dashboard renders. Used by FromURL.
type URLOptions struct {
	URL  string
	Kind Kind

	// GCP carries the customer-supplied GCP location/sizing inputs, forwarded
	// onto Options.GCP. Required (project + region) when the run targets GCP.
	GCP GCPOptions

	// Cloud overrides the target cloud provider. Empty uses the cloud the
	// ctl-api Config specifies, falling back to the default (aws).
	Cloud Cloud

	// Method overrides the provisioning method. Empty uses the method the
	// ctl-api Config specifies, falling back to the per-cloud default
	// (aws→sdk, others→terraform).
	Method Method
}

// GCPOptions holds the customer-supplied GCP inputs the module requires that
// ctl-api does not provide: the target project and region (required), plus
// optional runner sizing and GKE node-pool service-account controls. Empty
// optional fields fall back to the module's defaults. They are overlaid onto
// the ctl-api-provided Config.GCP at provision time.
type GCPOptions struct {
	ProjectID         string
	Region            string
	RunnerMachineType string
	// HasGKENodePool is a pointer so "unset" (nil) can fall back to the
	// module default (true) rather than forcing false.
	HasGKENodePool     *bool
	GKENodePoolSAEmail string
}

// logStreamConfig points the SDK at a Nuon ctl-api log stream for OTLP push.
// When nil, the SDK logs to stdout.
type logStreamConfig struct {
	ID           string
	WriteToken   string
	RunnerAPIURL string
}

// stackRunConfig points the SDK at the public ctl-api stack-run endpoints.
// Mirrors the phone-home pattern: PhoneHomeID is the per-stack-version
// secret embedded in the URL path; no API token is required.
type stackRunConfig struct {
	CtlAPIURL   string
	PhoneHomeID string
}
