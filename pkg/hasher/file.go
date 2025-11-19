package hasher

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

// HashFiles computes a SHA256 hash of multiple files combined
// Returns the first 12 characters of the hex hash
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

	// Return first 12 characters of hex hash
	return fmt.Sprintf("%x", hash.Sum(nil))[:12], nil
}
