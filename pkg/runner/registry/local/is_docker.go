package local

import (
	"bufio"
	"os"
	"strings"
)

func IsDocker() bool {
	switch {
	case fileExists("/.dockerenv"):
		return true
	case hasDockerInCgroup("/proc/1/cgroup"):
		return true
	case hasDockerInCgroup("/proc/self/cgroup"):
		return true
	default:
		return false
	}
}

func fileExists(path string) bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func hasDockerInCgroup(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		l := strings.ToLower(s.Text())
		if strings.Contains(l, "/docker/") {
			return true
		}
	}
	return false
}
