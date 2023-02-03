package install

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	t.Parallel()
	v := validator.New()
	tests := map[string]struct {
		d           string
		l           *log.Logger
		vers        string
		v           *validator.Validate
		assertions  func(*testing.T, *terraformInstaller)
		errExpected error
	}{
		"happy path": {
			d:    "/tmp",
			l:    log.Default(),
			v:    v,
			vers: "v1.2.3",
			assertions: func(t *testing.T, i *terraformInstaller) {
				assert.Equal(t, "/tmp", i.Dir)
				assert.NotEmpty(t, i.Logger)
				assert.Equal(t, product.Terraform.Name, i.installer.(*releases.ExactVersion).Product.Name)
				assert.Equal(t, "v1.2.3", i.Version)
			},
		},
		"uses default version": {
			d: "/tmp",
			l: log.Default(),
			v: v,
			assertions: func(t *testing.T, i *terraformInstaller) {
				assert.Equal(t, "/tmp", i.Dir)
				assert.NotEmpty(t, i.Logger)
				assert.Equal(t, product.Terraform.Name, i.installer.(*releases.ExactVersion).Product.Name)
				assert.Equal(t, defaultTerraformVersion, i.Version)
			},
		},
		"invalid version": {
			d:           "/tmp",
			l:           log.Default(),
			v:           v,
			vers:        "abc123",
			errExpected: fmt.Errorf("Malformed version"),
		},
		"missing directory": {
			d:           "",
			l:           log.Default(),
			v:           v,
			errExpected: fmt.Errorf("Field validation for 'Dir' failed on the 'required' tag"),
		},
		"bad directory": {
			d:           "/doesnotexist",
			l:           log.Default(),
			v:           v,
			errExpected: fmt.Errorf("Field validation for 'Dir' failed on the 'dir' tag"),
		},
		"missing logger": {
			d:           "/tmp",
			l:           nil,
			v:           v,
			errExpected: fmt.Errorf("Field validation for 'Logger' failed on the 'required' tag"),
		},
		"missing validator": {
			d:           "/tmp",
			l:           log.Default(),
			v:           nil,
			errExpected: fmt.Errorf("validator is nil"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfgs := []terraformInstallerOption{WithLogger(test.l), WithInstallDir(test.d)}
			if test.vers != "" {
				cfgs = append(cfgs, WithVersion(test.vers))
			}
			i, err := New(test.v, cfgs...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertions(t, i)
		})
	}
}

type mockInstaller struct{ mock.Mock }

func (m *mockInstaller) Install(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func TestTerraformInstaller_Install(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		installer   func(t *testing.T) *mockInstaller
		errExpected error
	}{
		"happy path": {
			installer: func(t *testing.T) *mockInstaller {
				m := &mockInstaller{}
				m.On("Install", mock.Anything).Return(t.Name(), nil)
				return m
			},
		},
		"errors": {
			installer: func(t *testing.T) *mockInstaller {
				m := &mockInstaller{}
				m.On("Install", mock.Anything).Return("", fmt.Errorf("oops"))
				return m
			},
			errExpected: fmt.Errorf("oops"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mck := test.installer(t)
			i := &terraformInstaller{installer: mck}
			s, err := i.Install(context.Background())

			mck.AssertExpectations(t)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, t.Name(), s)
		})
	}
}
