//go:build unix

package rootfs

import (
	"github.com/go-git/go-billy/v5"
	"github.com/spf13/afero"
	"github.com/spf13/afero/mem"
	"golang.org/x/sys/unix"
)

func lock(f afero.File, acquire bool) error {
	if file, ok := f.(interface{ Fd() uintptr }); ok {
		operation := unix.LOCK_UN
		if acquire {
			operation = unix.LOCK_EX
		}
		return unix.Flock(int(file.Fd()), operation)
	}
	if base, ok := f.(*afero.BasePathFile); ok {
		return lock(base.File, acquire)
	}
	if _, ok := f.(*mem.File); ok {
		// Injected memory workspaces have no interprocess writers.
		return nil
	}
	return billy.ErrNotSupported
}
