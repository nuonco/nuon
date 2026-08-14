package jobloop

import (
	"os"
	"strings"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (j *jobLoop) isSandbox(job *models.AppRunnerJob) bool {
	if job.Type == models.AppRunnerJobTypeShutDashDown {
		return false
	}

	// dev-only: let image-backed actions run their real container locally so
	// the supervisor + launcher contract can be validated without cloud.
	if job.Group == models.AppRunnerJobGroupImageDashActions && devRealImageActions() {
		return false
	}

	return j.settings.SandboxMode
}

func devRealImageActions() bool {
	return os.Getenv("NUON_DEV_REAL_IMAGE_ACTIONS") == "true" &&
		strings.EqualFold(os.Getenv("ENV"), "development")
}
