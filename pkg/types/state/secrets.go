package state

import "github.com/powertoolsdev/mono/pkg/types/outputs"

func NewSecretsState() outputs.SyncSecretsOutput {
	return make(outputs.SyncSecretsOutput, 0)
}

type SecretsState = outputs.SyncSecretsOutput
