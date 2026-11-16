package app

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestPreloadedModelSizes(t *testing.T) {
	tests := map[string]struct {
		size     uintptr
		maxBytes uintptr
	}{
		"install deploy": {
			size:     unsafe.Sizeof(InstallDeploy{}),
			maxBytes: 16 << 10,
		},
		"install component": {
			size:     unsafe.Sizeof(InstallComponent{}),
			maxBytes: 20 << 10,
		},
		"component": {
			size:     unsafe.Sizeof(Component{}),
			maxBytes: 8 << 10,
		},
		"install sandbox run": {
			size:     unsafe.Sizeof(InstallSandboxRun{}),
			maxBytes: 16 << 10,
		},
		"install app branch connection": {
			size:     unsafe.Sizeof(InstallAppBranchConnection{}),
			maxBytes: 64 << 10,
		},
		"install roles": {
			size:     unsafe.Sizeof(InstallRoles{}),
			maxBytes: 16 << 10,
		},
		"install inputs": {
			size:     unsafe.Sizeof(InstallInputs{}),
			maxBytes: 16 << 10,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require.LessOrEqual(t, tc.size, tc.maxBytes, "embed large relations as pointers")
		})
	}
}
