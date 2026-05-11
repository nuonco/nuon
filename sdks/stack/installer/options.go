package installer

// LogStreamConfig points the SDK at a Nuon ctl-api log stream for OTLP push.
// When nil, the SDK logs to stdout.
type LogStreamConfig struct {
	ID           string
	WriteToken   string
	RunnerAPIURL string
}

// StackRunConfig points the SDK at the public ctl-api stack-run endpoints.
// When set, the SDK creates a stack run at provision start and updates it
// with terminal status at the end. nil disables run reporting.
//
// Mirrors the phone-home pattern: PhoneHomeID is the per-stack-version
// secret embedded in the URL path; no API token is required.
type StackRunConfig struct {
	CtlAPIURL   string
	PhoneHomeID string
}

// Options configure an Installer run.
type Options struct {
	InstallID string
	AWSRegion string
	LogStream *LogStreamConfig
	StackRun  *StackRunConfig
}

// CreateRunURL is the single-argument bootstrap shape: the URL of the
// create-run POST endpoint that the dashboard renders. Used by FromCreateRunURL.
type CreateRunURL struct {
	URL  string
	Kind string // "provision" / "reprovision" / "deprovision"
}
