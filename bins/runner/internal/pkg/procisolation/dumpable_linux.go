//go:build linux

package procisolation

import "golang.org/x/sys/unix"

func setNonDumpable() error {
	return unix.Prctl(unix.PR_SET_DUMPABLE, 0, 0, 0, 0)
}
