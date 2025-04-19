package dev

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	smithytime "github.com/aws/smithy-go/time"
	"github.com/pkg/errors"
)

func (d *devver) Init(ctx context.Context) error {
	shouldMonitor := true
	if os.Getenv("RUNNER_ID") != "" {
		fmt.Println("disabling monitoring and restarting for new runners")
		shouldMonitor = false
	}

	disabled := d.Disabled()
	if disabled {
		fmt.Println("disabling and returning because of DISABLE_ORG_RUNNER or DISABLE_INSTALL_RUNNER in env")
		for {
			if err := smithytime.SleepWithContext(ctx, time.Second*5); err != nil {
				return err
			}
		}
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
		fmt.Println("monitoring for new runners")
		if err := d.monitorRunners(); err != nil {
			log.Fatalf("unable to monitor runners")
		}
	}()

	return nil
}
