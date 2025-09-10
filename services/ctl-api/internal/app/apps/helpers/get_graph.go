package helpers

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func componentHash(a *app.Component) string {
	return a.ID
}

type errComponentVertex struct {
	err    error
	compID string
}

func (e *errComponentVertex) Error() string {
	return fmt.Errorf("unable to add component to graph: %w", e.err).Error()
}

func (e *errComponentVertex) Unwrap() error {
	return e.err
}

type errComponentEdge struct {
	err          error
	compID       string
	dependencyID string
}

func (e *errComponentEdge) Error() string {
	return fmt.Errorf("unable to add component edge to graph: %w", e.err).Error()
}

func (e *errComponentEdge) Unwrap() error {
	return e.err
}
