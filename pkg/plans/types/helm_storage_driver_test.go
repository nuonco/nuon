package plantypes

import "testing"

func TestNormalizeHelmStorageDriver(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "secret", want: "secrets"},
		{in: "secrets", want: "secrets"},
		{in: "configmap", want: "configmap"},
		{in: "configmaps", want: "configmaps"},
		{in: "nuon", want: "nuon"},
		{in: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := NormalizeHelmStorageDriver(tt.in); got != tt.want {
				t.Errorf("NormalizeHelmStorageDriver(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
