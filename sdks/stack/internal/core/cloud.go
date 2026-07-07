package core

import "fmt"

// Cloud selects which cloud provider an install stack targets. Every cloud is
// provisioned with the Terraform method (internal/terraform).
type Cloud string

const (
	// CloudAWS provisions into AWS (EKS/EC2/IAM/Secrets Manager).
	CloudAWS Cloud = "aws"
	// CloudGCP provisions into GCP (GKE/GCE/service accounts/Secret Manager).
	CloudGCP Cloud = "gcp"
	// CloudAzure is reserved for future support and is not yet wired.
	CloudAzure Cloud = "azure"
)

// DefaultCloud is used when neither Config nor Options specifies one.
const DefaultCloud = CloudAWS

// DefaultMethodForCloud returns the provisioning method to use when none is
// explicitly set. Terraform is currently the only method for every cloud.
func DefaultMethodForCloud(_ Cloud) Method {
	return MethodTerraform
}

// ValidateCloud reports whether the cloud is supported by the Terraform method.
func ValidateCloud(cloud Cloud) error {
	switch cloud {
	case CloudAWS, CloudGCP:
		return nil
	}
	return fmt.Errorf("unsupported cloud %q", cloud)
}
