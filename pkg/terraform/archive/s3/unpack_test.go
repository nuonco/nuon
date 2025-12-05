package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mholt/archiver/v4"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
	"github.com/stretchr/testify/assert"
)

func Test_s3_extract(t *testing.T) {
	ctx := context.Background()

	tests := map[string]struct {
		archiveFn   func(*testing.T) io.Reader
		callbackFn  func(*gomock.Controller) archive.Callback
		errExpected error
	}{
		"happy path": {
			archiveFn: func(t *testing.T) io.Reader {
				buffer := new(bytes.Buffer)
				format := archiver.CompressedArchive{
					Compression: archiver.Gz{},
					Archival:    archiver.Tar{},
				}

				files, err := archiver.FilesFromDisk(nil, map[string]string{
					"./testdata/test.txt": "test.txt",
				})
				assert.NoError(t, err)
				err = format.Archive(ctx, buffer, files)
				assert.NoError(t, err)
				return buffer
			},
			callbackFn: func(mockCtl *gomock.Controller) archive.Callback {
				mock := archive.NewMockCallbacker(mockCtl)
				mock.EXPECT().Callback(gomock.Any(), "test.txt", gomock.Any()).Return(nil)
				return mock.Callback
			},
			errExpected: nil,
		},
		"not a valid gz compressed file": {
			archiveFn: func(t *testing.T) io.Reader {
				buffer := new(bytes.Buffer)
				format := archiver.CompressedArchive{
					Archival: archiver.Tar{},
				}

				files, err := archiver.FilesFromDisk(nil, map[string]string{
					"./testdata/test.txt": "test.txt",
				})
				assert.NoError(t, err)
				err = format.Archive(ctx, buffer, files)
				assert.NoError(t, err)
				return buffer
			},
			callbackFn: func(mockCtl *gomock.Controller) archive.Callback {
				mock := archive.NewMockCallbacker(mockCtl)
				return mock.Callback
			},
			errExpected: fmt.Errorf("unable to decompress gz"),
		},
		"empty archive": {
			archiveFn: func(t *testing.T) io.Reader {
				buffer := new(bytes.Buffer)

				format := archiver.CompressedArchive{
					Compression: archiver.Gz{},
					Archival:    archiver.Zip{},
				}
				files, err := archiver.FilesFromDisk(nil, map[string]string{})
				assert.NoError(t, err)
				err = format.Archive(ctx, buffer, files)
				assert.NoError(t, err)
				return buffer
			},
			callbackFn: func(mockCtl *gomock.Controller) archive.Callback {
				mock := archive.NewMockCallbacker(mockCtl)
				return mock.Callback
			},
		},
		"not a valid tar file": {
			archiveFn: func(t *testing.T) io.Reader {
				buffer := new(bytes.Buffer)

				format := archiver.CompressedArchive{
					Compression: archiver.Gz{},
					Archival:    archiver.Zip{},
				}
				files, err := archiver.FilesFromDisk(nil, map[string]string{
					"./testdata/test.txt": "test.txt",
				})
				assert.NoError(t, err)
				err = format.Archive(ctx, buffer, files)
				assert.NoError(t, err)
				return buffer
			},
			callbackFn: func(mockCtl *gomock.Controller) archive.Callback {
				mock := archive.NewMockCallbacker(mockCtl)
				return mock.Callback
			},
			errExpected: fmt.Errorf("unexpected EOF"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)

			reader := test.archiveFn(t)
			cb := test.callbackFn(mockCtl)

			obj := &s3{}
			err := obj.unpack(ctx, reader, cb)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
		})
	}
}
