package executors

type LogConfiguration struct {
	RunnerID       string `json:"runner_id"`
	RunnerAPIToken string `json:"runner_api_token"`
	RunnerAPIURL   string `json:"runner_api_url"`
	RunnerJobID    string `json:"runner_job_id"`

	Attrs map[string]string `json:"attrs"`
}
