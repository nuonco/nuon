package app

type Install struct {
	Model
	CreatedByID string

	Name  string
	AppID string
	App   App
}
