// build.go
package models

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
)

type Artifact struct {
	Model
}

func (a *Artifact) NewID() error {
	if a.ID == "" {
		id, err := shortid.NewNanoID("art")
		if err != nil {
			return fmt.Errorf("unable to make nanoid for artifact: %w", err)
		}
		a.ID = id
	}
	return nil
}
