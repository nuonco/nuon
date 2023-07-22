package dir

import (
	"context"
	"fmt"
	"os"
)

func (d *dir) Init(ctx context.Context) error {
	fh, err := os.Stat(d.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist: %w", d.Path, err)
		}
		return fmt.Errorf("unable to check path %s: %w", d.Path, err)
	}

	if !fh.IsDir() {
		return fmt.Errorf("is not a directory")
	}

	return nil
}
