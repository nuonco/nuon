package dev

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

func (d *devver) Init(ctx context.Context) error {
	type step struct {
		name string
		fn   func(context.Context) error
	}
	steps := []step{
		{"runner-id", d.initRunner},
		{"runner-api-token", d.initToken},
		{"runner-creds", d.initCreds},
	}
	for _, st := range steps {
		if err := st.fn(ctx); err != nil {
			return errors.Wrap(err, fmt.Sprintf("unable to initialize %s", st.name))
		}
	}

	return nil
}
