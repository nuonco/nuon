package nuonrunner

import "context"

type Retryer interface {
	Wrap(func(context.Context) error) func(context.Context) error
}

type defaultRetryer struct{}

func (d *defaultRetryer) Wrap(fn func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		return fn(ctx)
	}
}
