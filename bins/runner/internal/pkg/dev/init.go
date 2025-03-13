package dev

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
)

func (d *devver) Init(ctx context.Context) error {
	if d.Disabled() {
		fmt.Println("disabling and returning because of DISABLE_ORG_RUNNER or DISABLE_INSTALL_RUNNER in env")
		return nil
	}

	shouldMonitor := true
	if os.Getenv("RUNNER_ID") != "" {
		fmt.Println("disabling monitoring and restarting for new runners")
		shouldMonitor = false
	}

	type step struct {
		name string
		fn   func(context.Context) error
	}
	steps := []step{
		{"runner-id", d.initRunner},
		{"runner-api-token", d.initToken},
		{"runner-creds", d.initCreds},
		{"runner-env", d.initEnv},
	}
	for _, st := range steps {
		if err := st.fn(ctx); err != nil {
			return errors.Wrap(err, fmt.Sprintf("unable to initialize %s", st.name))
		}
	}

	if !shouldMonitor {
		return nil
	}
	go func() {
		if err := d.monitorRunners(); err != nil {
			log.Fatalf("unable to monitor runners")
		}
	}()

	return nil
}
