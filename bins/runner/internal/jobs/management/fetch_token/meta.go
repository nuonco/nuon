package fetchtoken

import "github.com/nuonco/nuon/sdks/nuon-runner-go/models"

func (h *handler) Name() string {
	return "fetch_token"
}

func (h *handler) JobType() models.AppRunnerJobType {
	// TODO: Add AppRunnerJobTypeMngDashFetchDashToken to the API spec
	return "mng-fetch-token"
}

func (h *handler) JobStatus() models.AppRunnerJobStatus {
	return models.AppRunnerJobStatusAvailable
}
