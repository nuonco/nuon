package api

import (
	"fmt"
)

func (a *API) registerServices() error {
	// register services
	for idx, svc := range a.services {
		if err := svc.RegisterPublicRoutes(a.public); err != nil {
			return fmt.Errorf("unable to register routes on svc %d: %w", idx, err)
		}

		if err := svc.RegisterInternalRoutes(a.internal); err != nil {
			return fmt.Errorf("unable to register internal routes on svc %d: %w", idx, err)
		}

		if err := svc.RegisterRunnerRoutes(a.runner); err != nil {
			return fmt.Errorf("unable to register runner routes on svc %d: %w", idx, err)
		}
	}

	return nil
}
