package state

func NewCloudAccount() *CloudAccount {
	return &CloudAccount{}
}

type CloudAccount struct {
	AWS   *AWSCloudAccount   `json:"aws"`
	Azure *AzureCloudAccount `json:"azure"`
	GCP   *GCPCloudAccount   `json:"gcp"`
}

type AWSCloudAccount struct {
	Region string `json:"region"`
}

type AzureCloudAccount struct {
	Location string `json:"location"`
}

type GCPCloudAccount struct {
	Region    string `json:"region"`
	ProjectID string `json:"project_id"`
}
