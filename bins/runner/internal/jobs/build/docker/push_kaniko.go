package docker

import (
	"context"

	"go.uber.org/zap"
)

func (b *handler) pushWithKaniko(
	ctx context.Context,
	log *zap.Logger,
	localRef string,
) error {
	_, err := b.kanikoPath()
	if err != nil {
		log.Info("pushing to local registry using local podman/docker")
		return b.pushLocal(
			ctx,
			log,
			localRef,
		)
	}

	log.Info("kaniko build --destination already pushed to local registry")
	return nil
}
