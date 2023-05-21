package remote

import (
	"io"
	"log"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/hc-install/product"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func Test_remote_getInstaller(t *testing.T) {
	v := validator.New()
	version := "0.0.1"
	dir := generics.GetFakeObj[string]()
	lg := log.New(io.Discard, "", 1)

	r, err := New(v, WithVersion(version))
	assert.NoError(t, err)

	installer := r.getInstaller(lg, dir)
	assert.Equal(t, installer.Product.Name, product.Terraform.Name)
	assert.Equal(t, installer.Version, r.Version)
	assert.Equal(t, installer.InstallDir, dir)
}
