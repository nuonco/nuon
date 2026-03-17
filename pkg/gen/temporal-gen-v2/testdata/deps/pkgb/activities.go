package pkgb

import (
	"context"

	// Import pkgc to establish a dependency edge in the import graph.
	// This ensures pkgc is processed before pkgb.
	_ "github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/testdata/deps/pkgc"
)

// PkgbActivity is an activity that depends on pkgc being processed first.
// @temporal-gen-v2 activity
func PkgbActivity(ctx context.Context, input string) (string, error) {
	return input, nil
}
