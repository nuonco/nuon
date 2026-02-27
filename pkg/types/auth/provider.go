package auth

import "fmt"

// CloudProvider represents a cloud provider type
type CloudProvider string

const (
	CloudProviderAWS   CloudProvider = "aws"
	CloudProviderGCP   CloudProvider = "gcp"
	CloudProviderAzure CloudProvider = "azure"
)

// String returns the string representation of the cloud provider
func (c CloudProvider) String() string {
	return string(c)
}

// Valid checks if the cloud provider is valid
func (c CloudProvider) Valid() bool {
	switch c {
	case CloudProviderAWS, CloudProviderGCP, CloudProviderAzure:
		return true
	default:
		return false
	}
}

// Validate returns an error if the cloud provider is invalid
func (c CloudProvider) Validate() error {
	if !c.Valid() {
		return fmt.Errorf("invalid cloud provider: %s", c)
	}
	return nil
}
