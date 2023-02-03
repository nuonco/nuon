package runner

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-multierror"
)

type runner struct {
	Bucket string `validate:"required"`
	Key    string `validate:"required"`
	Region string `validate:"required"`

	// internal state
	validator        *validator.Validate
	cleanupFns       []func() error
	moduleFetcher    moduleFetcher
	requestParser    requestParser
	workspaceSetuper workspaceSetuper
}

type runnerOption func(*runner) error

// New instantiates a new runner
func New(v *validator.Validate, opts ...runnerOption) (*runner, error) {
	r := &runner{}

	if v == nil {
		return nil, fmt.Errorf("error instantiating workspace: validator is nil")
	}
	r.validator = v

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	if err := r.validator.Struct(r); err != nil {
		return nil, err
	}

	r.moduleFetcher = r
	r.requestParser = r
	r.workspaceSetuper = r

	return r, nil
}

// WithBucket specifies the bucket for the request
func WithBucket(s string) runnerOption {
	return func(r *runner) error {
		r.Bucket = s
		return nil
	}
}

// WithKey specifies the bucket key (object name) for the request
func WithKey(s string) runnerOption {
	return func(r *runner) error {
		r.Key = s
		return nil
	}
}

// WithRegion specifies the bucket region for the request
func WithRegion(s string) runnerOption {
	return func(r *runner) error {
		r.Region = s
		return nil
	}
}

// cleanup runs the cleanup functions for the runner and returns the consolidated errors
// safe to run even if there are no cleanupFns
func (r *runner) cleanup() error {
	var err error
	for _, fn := range r.cleanupFns {
		e := fn()
		if e != nil {
			err = multierror.Append(err, e)
		}
	}
	return err
}
