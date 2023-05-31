package oci

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2/content/file"
)

const (
	defaultStorePath string = "/tmp"
)

func (o *oci) Init(ctx context.Context) error {
	fs, err := file.New(o.tmpDir)
	if err != nil {
		return fmt.Errorf("unable to initialize store: %w", err)
	}
	o.store = fs

	return nil
}
