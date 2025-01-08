package plantypes

type GitSource struct {
	URL  string `json:"url" validate:"required"`
	Ref  string `json:"ref" validate:"required"`
	Path string `json:"path" validate:"required"`

	RecurseSubmodules bool `json:"recurse_submodules"`
}
