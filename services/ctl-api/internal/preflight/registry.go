package preflight

// registry is an ordered slice rather than a map so the results table and
// --list output keep a stable order between runs.
var registry = []Check{
	rdsCheck,
	clickhouseCheck,
	temporalCheck,
	nuonAuthCheck,
	githubCheck,
	awsCheck,
	kafkaCheck,
	blobstoreCheck,
	slackCheck,
}

func All() []Check {
	return registry
}

func Lookup(name string) (Check, bool) {
	for _, check := range registry {
		if check.Name == name {
			return check, true
		}
	}

	return Check{}, false
}

func Names() []string {
	names := make([]string, 0, len(registry))
	for _, check := range registry {
		names = append(names, check.Name)
	}

	return names
}
