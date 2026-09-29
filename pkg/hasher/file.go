package hasher

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

func HashFiles(filePaths ...string) (string, error) {
	hash := sha256.New()

	for _, filePath := range filePaths {
		if err := func() error {
			f, err := os.Open(filePath)
			if err != nil {
				return errors.Wrapf(err, "unable to open file: %s", filePath)
			}
			defer f.Close()

			if _, err := io.Copy(hash, f); err != nil {
				return errors.Wrapf(err, "unable to hash file: %s", filePath)
			}
			return nil
		}(); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil))[:12], nil
}
