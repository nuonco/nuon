package get

import (
	"strings"

	getter "github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
)

func GetDetectors() []getter.Detector {
	return append([]getter.Detector{
		detector{},
		new(getter.GitHubDetector),
		new(getter.GitDetector),
		new(getter.S3Detector),
		new(getter.GCSDetector),
		new(getter.FileDetector),
	})
}

type detector struct{}

func (detector) Detect(val string, pwd string) (string, bool, error) {
	invalidPrefixes := []string{
		"{",
		"}",
	}

	for _, ivp := range invalidPrefixes {
		if strings.HasPrefix(val, ivp) {
			return val, false, errors.New("invalid prefix in getter method")
		}
	}

	return val, false, nil
}
