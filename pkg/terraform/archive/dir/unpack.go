package dir

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/facebookgo/symwalk"
	"github.com/nuonco/nuon/pkg/terraform/archive"
)

const (
	dotTerraformPrefix string = ".terraform/"
	// terraformModulesPrefix is the subtree under .terraform/ that holds
	// vendored remote modules. We let it pass through the
	// IgnoreDotTerraformDir filter so build runners can ship modules
	// inside the OCI artifact (via `terraform get`) and install runners
	// pick them up at unpack time, avoiding a network fetch during
	// `terraform init`.
	terraformModulesPrefix string = ".terraform/modules/"
	terraformLockFile      string = ".terraform.lock.hcl"
	terraformStateFile     string = "terraform.tfstate"
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
		if d.IgnoreDotTerraformDir &&
			strings.HasPrefix(relPath, dotTerraformPrefix) &&
			!strings.HasPrefix(relPath, terraformModulesPrefix) {
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

	if d.AddBackendFile {
		str := fmt.Sprintf(`terraform { 
		                backend "%s" {
		                }
		        }
		`, d.AddBackendType)

		rc := io.NopCloser(strings.NewReader(str))
		if err := cb(ctx, "backend_config.tf", rc); err != nil {
			return fmt.Errorf("unable to add backend config file: %w", err)
		}
	}
	return nil
}
