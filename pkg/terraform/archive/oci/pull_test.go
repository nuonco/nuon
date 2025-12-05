package oci

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func Test_oci_pull(t *testing.T) {
	tests := map[string]struct {
		filesFn     func(*testing.T) map[string][]testFile
		tag         string
		assertFn    func(*testing.T, context.Context, *oci, map[string][]testFile)
		errExpected error
	}{
		"happy path": {
			filesFn: func(t *testing.T) map[string][]testFile {
				return map[string][]testFile{
					"basic": {
						generics.GetFakeObj[testFile](),
					},
				}
			},
			tag: "basic",
			assertFn: func(t *testing.T, ctx context.Context, obj *oci, artifacts map[string][]testFile) {
				desc, err := obj.store.Resolve(ctx, "basic")
				assert.NoError(t, err)
				assert.NotNil(t, desc)
				assert.Equal(t, defaultArtifactType, desc.ArtifactType)
				assert.NotEmpty(t, desc.Digest)

				// make sure all files exist in the local file path
				for _, file := range artifacts["basic"] {
					expectedFp := filepath.Join(obj.tmpDir, file.Name)
					_, err := os.Stat(expectedFp)
					assert.NoError(t, err)

					byts, err := os.ReadFile(expectedFp)
					assert.NoError(t, err)
					assert.Equal(t, file.Bytes, byts)
				}
			},
		},
		"can pull one tag, when multiple exist": {
			filesFn: func(t *testing.T) map[string][]testFile {
				return map[string][]testFile{
					"basic": {
						generics.GetFakeObj[testFile](),
					},
					"other": {
						{
							Name:      "test.txt",
							Bytes:     []byte("hello world"),
							MediaType: generics.GetFakeObj[string](),
						},
					},
				}
			},
			tag: "basic",
			assertFn: func(t *testing.T, _ context.Context, obj *oci, _ map[string][]testFile) {
				expectedFp := filepath.Join(obj.tmpDir, "test.txt")
				_, err := os.Stat(expectedFp)
				assert.Error(t, err)
				assert.True(t, os.IsNotExist(err))
			},
		},
		"supports nested directories": {
			filesFn: func(t *testing.T) map[string][]testFile {
				return map[string][]testFile{
					"basic": {
						testFile{
							Name:      "test/data.txt",
							Bytes:     []byte("hello world"),
							MediaType: generics.GetFakeObj[string](),
						},
					},
					"other": {
						generics.GetFakeObj[testFile](),
					},
				}
			},
			tag: "basic",
			assertFn: func(t *testing.T, ctx context.Context, obj *oci, _ map[string][]testFile) {
				expectedFp := filepath.Join(obj.tmpDir, "test/data.txt")
				byts, err := os.ReadFile(expectedFp)
				assert.NoError(t, err)
				assert.Equal(t, []byte("hello world"), byts)
			},
		},
		"tag does not exist": {
			filesFn: func(t *testing.T) map[string][]testFile {
				return map[string][]testFile{
					"basic": {
						generics.GetFakeObj[testFile](),
					},
				}
			},
			tag:         "not-found",
			errExpected: fmt.Errorf("unable to copy"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancelFn := context.WithCancel(ctx)
			defer cancelFn()

			testFiles := test.filesFn(t)
			fs := testStore(t, testFiles)

			obj := &oci{
				v:       validator.New(),
				testSrc: fs,
				tmpDir:  t.TempDir(),
				Image: &Image{
					Registry: generics.GetFakeObj[string](),
					Repo:     generics.GetFakeObj[string](),
					Tag:      test.tag,
				},
				Auth: generics.GetFakeObj[*Auth](),
			}
			err := obj.Init(ctx)
			assert.NoError(t, err)

			err = obj.pull(ctx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, ctx, obj, testFiles)
		})
	}
}
