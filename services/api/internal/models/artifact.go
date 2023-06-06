// build.go
package models

import (
	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
)

type Artifact struct {
	Model
}

func (a *Artifact) NewID() error {
	if a.ID == "" {
		a.ID = domains.NewArtifactID()
	}
	return nil
}
