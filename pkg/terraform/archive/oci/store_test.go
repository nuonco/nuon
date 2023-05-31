package oci

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
)

// testFile is a test file that is used for writing into an archive
type testFile struct {
	Name      string
	Bytes     []byte
	MediaType string
}

// testStore creates a store with different images in the store, based on their tag
func testStore(t *testing.T, artifacts map[string][]testFile) *file.Store {
	ctx := context.Background()
	tmpDir := t.TempDir()
	fs, err := file.New(tmpDir)
	assert.NoError(t, err)
	fs.AllowPathTraversalOnWrite = true

	for tag, files := range artifacts {
		// fetch file descriptors
		descriptors := make([]v1.Descriptor, 0, len(files))
		for _, file := range files {
			fp := filepath.Join(tmpDir, file.Name)
			err := os.MkdirAll(filepath.Dir(fp), 0744)
			assert.NoError(t, err)

			err = os.WriteFile(fp, file.Bytes, 0600)
			assert.NoError(t, err)

			desc, err := fs.Add(ctx, file.Name, file.MediaType, file.Name)
			assert.NoError(t, err)
			descriptors = append(descriptors, desc)
		}

		// pack files
		manifest, err := oras.Pack(ctx, fs, defaultArtifactType, descriptors, oras.PackOptions{
			PackImageManifest: true,
		})
		assert.NoError(t, err)

		err = fs.Tag(ctx, manifest, tag)
		assert.NoError(t, err)
	}

	return fs
}
