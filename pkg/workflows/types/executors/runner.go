package executors

const (
	ProvisionRunnerWorkflowName   = "ProvisionRunner"
	DeprovisionRunnerWorkflowName = "DeprovisionRunner"
)

type ProvisionRunnerRequestImage struct {
	URL string `validate:"required"`
	Tag string `validate:"tag"`
}

type ProvisionRunnerRequest struct {
	RunnerID string                      `validate:"required"`
	APIURL   string                      `validate:"required"`
	APIToken string                      `validate:"required"`
	Image    ProvisionRunnerRequestImage `validate:"required"`
}

type ProvisionRunnerResponse struct{}

type DeprovisionRunnerRequest struct {
	RunnerID string `validate:"required"`
}

type DeprovisionRunnerResponse struct{}
