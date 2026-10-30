//go:build !linux

package procisolation

func setNonDumpable() error {
	return nil
}
