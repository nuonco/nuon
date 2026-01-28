package aws

// PresignedRequest contains the data needed to make a presigned AWS request
type PresignedRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}
