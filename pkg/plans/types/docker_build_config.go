package plantypes

type DockerBuildPlan struct {
	BuildArgs  map[string]*string `json:"build_args"`
	Target     string             `json:"target"`
	Context    string             `json:"context"`
	Dockerfile string             `json:"dockerfile"`
}
