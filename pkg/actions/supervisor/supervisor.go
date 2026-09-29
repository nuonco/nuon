package supervisor

import (
	_ "embed"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
)

//go:embed supervisor.sh
var Script []byte

const (
	OutputFilepathEnvVar = "NUON_ACTIONS_OUTPUT_FILEPATH"
	RootEnvVar           = "NUON_ACTIONS_ROOT"

	Filename = ".nuon-actions-supervisor.sh"
)

func Write(dir string) (string, error) {
	path := filepath.Join(dir, Filename)

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return "", errors.Wrap(err, "unable to clear existing supervisor path")
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0o755)
	if err != nil {
		return "", errors.Wrap(err, "unable to create actions supervisor script")
	}
	defer f.Close()

	if _, err := f.Write(Script); err != nil {
		return "", errors.Wrap(err, "unable to write actions supervisor script")
	}
	return path, nil
}
