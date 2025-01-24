package git

type Source struct {
	URL string `json:"url" validate:"required"`
	Ref string `json:"ref" validate:"required"`
}
