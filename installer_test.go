package terraform

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockHcTerraformInstaller struct {
	mock.Mock
}

func (m *mockHcTerraformInstaller) SetLogger(l *log.Logger) {
	m.Called(l)
}

func (m *mockHcTerraformInstaller) Install(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *mockHcTerraformInstaller) Remove(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

var _ hcTerraformInstaller = (*mockHcTerraformInstaller)(nil)

type mockInstaller struct {
	mock.Mock
}

func (m *mockInstaller) initTerraformInstaller(l *log.Logger) error {
	args := m.Called(l)
	return args.Error(0)
}

func (m *mockInstaller) installTerraform(ctx context.Context) (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *mockInstaller) removeTerraform(ctx context.Context) error {
	args := m.Called()
	return args.Error(0)
}

var _ terraformInstaller = (*mockInstaller)(nil)

func TestInstaller_init(t *testing.T) {
	i := &tfInstaller{}

	err := i.initTerraformInstaller(nil)
	assert.NoError(t, err)
	assert.NotNil(t, i.installer)
}

func TestInstaller_install(t *testing.T) {
	mockInstall := new(mockHcTerraformInstaller)
	mockInstall.On("Install", mock.Anything).Return("test-path", nil)

	install := &tfInstaller{mockInstall}
	path, err := install.installTerraform(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "test-path", path)
	mockInstall.AssertNumberOfCalls(t, "Install", 1)
}

func TestInstaller_remove(t *testing.T) {
	mockInstall := new(mockHcTerraformInstaller)
	mockInstall.On("Remove", mock.Anything).Return(nil)

	install := &tfInstaller{mockInstall}
	err := install.removeTerraform(context.Background())
	assert.NoError(t, err)
	mockInstall.AssertNumberOfCalls(t, "Remove", 1)
}
