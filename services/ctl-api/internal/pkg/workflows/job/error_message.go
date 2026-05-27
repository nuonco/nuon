package job

// JobErrorMessage extracts a user-facing message from an ExecuteJob error.
// The ExecuteJob workflow returns errors like "job did not succeed: <description>"
// which contain the actual runner job StatusDescription. This helper surfaces
// that message for status updates instead of using hardcoded fallback strings.
func JobErrorMessage(err error, fallback string) string {
	if err == nil {
		return fallback
	}
	msg := err.Error()
	if msg != "" {
		return msg
	}
	return fallback
}
