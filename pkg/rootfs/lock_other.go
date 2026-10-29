//go:build !unix

package rootfs

import (
	"github.com/go-git/go-billy/v5"
	"github.com/spf13/afero"
)

func lock(f afero.File, acquire bool) error { return billy.ErrNotSupported }
