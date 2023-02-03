//go:build integration

package install

import (
	"context"
	"log"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestTerraformInstaller_Install_Int(t *testing.T) {
	t.Parallel()
	tmpdir := t.TempDir()
	i, err := New(validator.New(), WithInstallDir(tmpdir), WithLogger(log.Default()))
	assert.NoError(t, err)
	defer i.Cleanup()

	p, err := i.Install(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, tmpdir+"/terraform", p)
}

func TestTerraformInstaller_Cleanup_Int(t *testing.T) {
	t.Parallel()
	tmpdir := t.TempDir()
	i, err := New(validator.New(), WithInstallDir(tmpdir), WithLogger(log.Default()))
	assert.NoError(t, err)
	assert.NoError(t, i.Cleanup())
}
