package job

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
