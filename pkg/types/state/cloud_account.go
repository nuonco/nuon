package state

func NewCloudAccount() *CloudAccount {
	return &CloudAccount{}
}

type CloudAccount struct {
	AWS   *AWSCloudAccount   `json:"aws"`
	Azure *AzureCloudAccount `json:"azure"`
}

type AWSCloudAccount struct {
	Region string `json:"region"`
}

type AzureCloudAccount struct {
	Location string `json:"location"`
}
