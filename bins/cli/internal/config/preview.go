package config

import "os"

const PreviewEnvVar string = "NUON_PREVIEW"

func Preview() bool {
	return os.Getenv(PreviewEnvVar) == "true"
}
