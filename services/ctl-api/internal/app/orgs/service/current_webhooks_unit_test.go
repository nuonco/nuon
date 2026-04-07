package service

import "testing"

func TestNormalizeWebhookURL(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedURL string
		expectErr   bool
	}{
		{
			name:        "normalizes valid https URL",
			input:       " https://example.com/hooks/step-lifecycle ",
			expectedURL: "https://example.com/hooks/step-lifecycle",
			expectErr:   false,
		},
		{
			name:      "rejects empty URL",
			input:     "",
			expectErr: true,
		},
		{
			name:      "rejects non-http scheme",
			input:     "ftp://example.com/hooks",
			expectErr: true,
		},
		{
			name:      "rejects relative URL",
			input:     "/hooks/step-lifecycle",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normalizedURL, err := normalizeWebhookURL(tc.input)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error for input %q", tc.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error for input %q, got: %v", tc.input, err)
			}

			if normalizedURL != tc.expectedURL {
				t.Fatalf("expected normalized URL %q, got %q", tc.expectedURL, normalizedURL)
			}
		})
	}
}
