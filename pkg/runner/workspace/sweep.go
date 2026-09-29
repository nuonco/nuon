package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const workspaceDirPrefix = "workspace-"

func SweepStale(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var removed []string
	var errs []error
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), workspaceDirPrefix) {
			continue
		}
		path := filepath.Join(root, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			errs = append(errs, fmt.Errorf("failed to remove stale workspace %s: %w", path, err))
			continue
		}
		removed = append(removed, entry.Name())
	}

	return removed, errors.Join(errs...)
}

func FilesystemType(path string) string {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return ""
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	abs = filepath.Clean(abs)

	var bestPoint, bestType string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		point := strings.ReplaceAll(fields[1], `\040`, " ")
		if !isSubPath(point, abs) {
			continue
		}
		if len(point) >= len(bestPoint) {
			bestPoint, bestType = point, fields[2]
		}
	}

	return bestType
}

func isSubPath(mountPoint, path string) bool {
	if mountPoint == "/" || mountPoint == path {
		return true
	}
	return strings.HasPrefix(path, strings.TrimSuffix(mountPoint, "/")+"/")
}
