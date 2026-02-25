package credentials

type Config struct {
	ProjectID string `json:"project_id" temporaljson:"project_id"`
	Region    string `json:"region" temporaljson:"region"`
}
