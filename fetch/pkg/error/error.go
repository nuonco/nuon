package error

import (
	"context"
	"io"
)

type errorFetcher struct {
	error
}

func New(e error) *errorFetcher {
	return &errorFetcher{error: e}
}

func (ef *errorFetcher) Fetch(ctx context.Context) (io.ReadCloser, error) {
	return nil, ef
}

func (ef *errorFetcher) Close() error { return ef }
