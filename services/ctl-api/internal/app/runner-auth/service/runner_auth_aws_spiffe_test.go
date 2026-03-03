package service

import (
	"testing"

	"github.com/spiffe/go-spiffe/v2/spiffeid"

	awstypes "github.com/nuonco/nuon/pkg/types/aws"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func TestValidateRunnerAuthAWSRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     RunnerAuthAWSRequest
		expectMode  string
		expectError bool
	}{
		{
			name: "presigned request",
			request: RunnerAuthAWSRequest{
				STSRequest:  &awstypes.PresignedRequest{Method: "GET", URL: "https://sts.us-west-2.amazonaws.com/"},
				TagsRequest: &awstypes.PresignedRequest{Method: "GET", URL: "https://ec2.us-west-2.amazonaws.com/"},
			},
			expectMode: runnerAuthModePresigned,
		},
		{
			name: "spiffe request",
			request: RunnerAuthAWSRequest{
				SPIFFEJWTSVID: "jwt",
			},
			expectMode: runnerAuthModeSPIFFE,
		},
		{
			name: "invalid mixed request",
			request: RunnerAuthAWSRequest{
				SPIFFEJWTSVID: "jwt",
				STSRequest:    &awstypes.PresignedRequest{},
			},
			expectError: true,
		},
		{
			name: "invalid partial presigned request",
			request: RunnerAuthAWSRequest{
				STSRequest: &awstypes.PresignedRequest{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, err := validateRunnerAuthAWSRequest(tt.request)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected validation error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}

			if mode != tt.expectMode {
				t.Fatalf("unexpected auth mode: got %q want %q", mode, tt.expectMode)
			}
		})
	}
}

func TestParseSPIFFEAWSIdentity(t *testing.T) {
	tests := []struct {
		name        string
		service     *service
		id          string
		expectError bool
		expect      *spiffeAWSIdentity
	}{
		{
			name:    "valid default prefix",
			service: &service{},
			id:      "spiffe://example.org/nuon/runner/aws/account/123456789012/instance/i-0abc1234def567890/runner/run123",
			expect: &spiffeAWSIdentity{
				RunnerID:   "run123",
				AccountID:  "123456789012",
				InstanceID: "i-0abc1234def567890",
				SPIFFEID:   "spiffe://example.org/nuon/runner/aws/account/123456789012/instance/i-0abc1234def567890/runner/run123",
			},
		},
		{
			name: "valid custom prefix",
			service: &service{cfg: &internal.Config{
				RunnerAuthAWSSPIFFEPathPrefix: "/custom/aws",
			}},
			id: "spiffe://example.org/custom/aws/account/123456789012/instance/i-0abc1234def567890/runner/run456",
			expect: &spiffeAWSIdentity{
				RunnerID:   "run456",
				AccountID:  "123456789012",
				InstanceID: "i-0abc1234def567890",
				SPIFFEID:   "spiffe://example.org/custom/aws/account/123456789012/instance/i-0abc1234def567890/runner/run456",
			},
		},
		{
			name:        "invalid account id",
			service:     &service{},
			id:          "spiffe://example.org/nuon/runner/aws/account/not-an-account/instance/i-0abc1234def567890/runner/run123",
			expectError: true,
		},
		{
			name:        "invalid prefix",
			service:     &service{},
			id:          "spiffe://example.org/other/prefix/account/123456789012/instance/i-0abc1234def567890/runner/run123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := spiffeid.RequireFromString(tt.id)
			identity, err := tt.service.parseSPIFFEAWSIdentity(id)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected parse error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			if identity.RunnerID != tt.expect.RunnerID ||
				identity.AccountID != tt.expect.AccountID ||
				identity.InstanceID != tt.expect.InstanceID ||
				identity.SPIFFEID != tt.expect.SPIFFEID {
				t.Fatalf("unexpected parsed identity: got %+v want %+v", identity, tt.expect)
			}
		})
	}
}
