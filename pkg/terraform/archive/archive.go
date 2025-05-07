package archive

import (
	"context"
	"io"
)

// Archive package exposes methods for loading a workspace archive
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=archive_mock.go -source=archive.go -package=archive
type Archive interface {
	// Init should be used for fetching things from s3, or setting up credentials
	Init(context.Context) error

	// Unpack is used to unpack an archive, and should call the unpackFn with each source file
	Unpack(context.Context, Callback) error

	Cleanup(context.Context) error
}

// Callback is passed in to the call back function to allow the implementer to pass a function in
type Callback func(context.Context, string, io.ReadCloser) error

type Callbacker interface {
	Callback(context.Context, string, io.ReadCloser) error
}
