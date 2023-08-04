package string

import (
	"context"
	"io"
	"strings"
)

type stringFetcher struct {
	io.Reader
}

func New(s string) *stringFetcher {
	return &stringFetcher{Reader: strings.NewReader(s)}
}

func (sf *stringFetcher) Fetch(ctx context.Context) (io.ReadCloser, error) {
	return sf, nil
}

func (sf *stringFetcher) Close() error { return nil }
