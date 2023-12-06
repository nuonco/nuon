package dir

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/facebookgo/symwalk"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
)

const (
	dotTerraformPrefix string = ".terraform/"
	terraformLockFile  string = ".terraform.lock.hcl"
	terraformStateFile string = "terraform.tfstate"
)

func (d *dir) Unpack(ctx context.Context, cb archive.Callback) error {
	fn := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		rc, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("unable to open file: %w", err)
		}

		relPath := strings.TrimPrefix(path, d.Path+"/")
		if d.IgnoreDotTerraformDir && strings.HasPrefix(relPath, dotTerraformPrefix) {
			return nil
		}
		if d.IgnoreTerraformLockFile && relPath == terraformLockFile {
			return nil
		}
		if d.IgnoreTerraformStateFile && relPath == terraformStateFile {
			return nil
		}

		if err := cb(ctx, relPath, rc); err != nil {
			return fmt.Errorf("unable to execute callback: %w", err)
		}
		return nil
	}

	if err := symwalk.Walk(d.Path, fn); err != nil {
		return fmt.Errorf("unable to walk root directory: %w", err)
	}
	return nil
}
