package shell

import (
	"os"
	"path/filepath"
)

func (s *shell) existsAndExecutable(name string) (bool, error) {
	fp := filepath.Join(s.rootDir, name)

	stat, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	perms := stat.Mode()
	if perms&0111 != 0 {
		return true, nil
	}

	return false, nil
}
