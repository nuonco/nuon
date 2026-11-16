package nuonauth

// First-party public client for `nuon auth login`. The CLI binds 127.0.0.1 on
// a random port and the auth server accepts that port against CLILoopbackRedirectURI.
const (
	CLIClientID            = "oacnuoncli0000000000000000"
	CLIClientName          = "Nuon CLI"
	CLICallbackPath        = "/callback"
	CLILoopbackRedirectURI = "http://127.0.0.1" + CLICallbackPath
	CLICompletePath        = "/cli/complete"

	CLIResultSuccess = "success"
	CLIResultDenied  = "denied"
	CLIResultFailed  = "failed"
)
