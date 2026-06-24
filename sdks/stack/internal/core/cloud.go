package core

import "fmt"

// Cloud selects which cloud provider an install stack targets. It is
// orthogonal to Method: a provisioner is chosen by the (Cloud, Method) pair.
// AWS supports both the SDK and Terraform methods; GCP (and, later, Azure)
// support Terraform only.
type Cloud string

const (
	// CloudAWS provisions into AWS (EKS/EC2/IAM/Secrets Manager).
	CloudAWS Cloud = "aws"
	// CloudGCP provisions into GCP (GKE/GCE/service accounts/Secret Manager).
	CloudGCP Cloud = "gcp"
	// CloudAzure is reserved for future support and is not yet wired.
	CloudAzure Cloud = "azure"
)

// DefaultCloud is used when neither Config nor Options specifies one. AWS keeps
// the historical default so existing runs (whose Config omits cloud) behave
// exactly as before.
const DefaultCloud = CloudAWS

// DefaultMethodForCloud returns the provisioning method to use when none is
// explicitly set. AWS defaults to the SDK method (its historical default);
// every other cloud defaults to Terraform, the only method it supports.
func DefaultMethodForCloud(cloud Cloud) Method {
	if cloud == CloudAWS {
		return MethodSDK
	}
	return MethodTerraform
}

// ValidateCloudMethod reports whether the (cloud, method) pair is supported.
// It is the single source of truth for the selection matrix; selectProvisioner
// relies on it so unsupported combinations fail with a clear, uniform message
// rather than constructing a half-wired provisioner.
func ValidateCloudMethod(cloud Cloud, method Method) error {
	switch cloud {
	case CloudAWS:
		switch method {
		case MethodSDK, MethodTerraform:
			return nil
		}
	case CloudGCP:
		if method == MethodTerraform {
			return nil
		}
		return fmt.Errorf("cloud %q does not support method %q; use %q", cloud, method, MethodTerraform)
	}
	return fmt.Errorf("unsupported cloud/method combination: cloud %q, method %q", cloud, method)
}
