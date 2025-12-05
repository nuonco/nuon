package sync

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/hasher"
)

type componentChecksum struct {
	LegacyChecksum string
	Checksum       string
}

func (s *componentChecksum) Equals(checksum string) bool {
	return s.Checksum == checksum || s.LegacyChecksum == checksum
}

func (s *sync) generateComponentChecksun(ctx context.Context, comp *config.Component) (componentChecksum, error) {
	legacyChecksum, err := hasher.HashStruct(comp, hasher.StructHasherOptions{
		EnableOmitEmpty: false,
	})
	if err != nil {
		return componentChecksum{}, err
	}

	checksum, err := hasher.HashStruct(comp, hasher.StructHasherOptions{
		EnableOmitEmpty: true,
	})
	if err != nil {
		return componentChecksum{}, err
	}
	return componentChecksum{
		LegacyChecksum: legacyChecksum,
		Checksum:       checksum,
	}, nil
}
