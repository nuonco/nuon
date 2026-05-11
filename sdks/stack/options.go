package stack

// Options configures a stack Installer. Use FromURL when you have a
// create-run URL from the dashboard; use New when you already know the
// install ID and region (e.g. for offline Status inspection).
type Options struct {
	InstallID string
	AWSRegion string

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
