package archive

import (
	"context"
	"io"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=archive_mock.go -source=archive.go -package=archive
type Archive interface {
	Init(context.Context) error

	Unpack(context.Context, Callback) error

	Cleanup(context.Context) error
}

type Callback func(context.Context, string, io.ReadCloser) error

type Callbacker interface {
	Callback(context.Context, string, io.ReadCloser) error
}
