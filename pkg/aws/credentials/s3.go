package credentials

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Options returns s3.Options function overrides based on Config fields.
// When S3Endpoint or S3ForcePathStyle are set, the corresponding options
// are returned. When neither is set, returns nil (no overrides).
func (c *Config) S3Options() []func(*s3.Options) {
	if c == nil {
		return nil
	}

	var opts []func(*s3.Options)

	if c.S3Endpoint != "" {
		endpoint := c.S3Endpoint
		opts = append(opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	if c.S3ForcePathStyle {
		opts = append(opts, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	return opts
}
