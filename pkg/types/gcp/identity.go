package gcp

// MetadataRequest contains the data needed to make a GCP Compute API request
// to independently read instance metadata. Parallel to aws.PresignedRequest.
type MetadataRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}
