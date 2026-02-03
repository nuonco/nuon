package principal

import (
	"strings"
	"testing"
)

func TestParsePrincipal(t *testing.T) {
	tests := []struct {
		name          string
		principal     string
		expectedType  string
		expectedName  string
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid component with name",
			principal:    "nuon::component:database",
			expectedType: "component",
			expectedName: "database",
			expectError:  false,
		},
		{
			name:         "valid component with wildcard",
			principal:    "nuon::component:*",
			expectedType: "component",
			expectedName: "*",
			expectError:  false,
		},
		{
			name:         "valid sandbox without name",
			principal:    "nuon::sandbox",
			expectedType: "sandbox",
			expectedName: "",
			expectError:  false,
		},
		{
			name:         "valid action with name",
			principal:    "nuon::action:migrate_db",
			expectedType: "action",
			expectedName: "migrate_db",
			expectError:  false,
		},
		{
			name:         "valid action with wildcard",
			principal:    "nuon::action:*",
			expectedType: "action",
			expectedName: "*",
			expectError:  false,
		},
		{
			name:         "component with hyphenated name",
			principal:    "nuon::component:api-server",
			expectedType: "component",
			expectedName: "api-server",
			expectError:  false,
		},
		{
			name:         "component with underscored name",
			principal:    "nuon::component:db_migration",
			expectedType: "component",
			expectedName: "db_migration",
			expectError:  false,
		},
		{
			name:         "action with complex name",
			principal:    "nuon::action:post-deploy-healthcheck",
			expectedType: "action",
			expectedName: "post-deploy-healthcheck",
			expectError:  false,
		},
		{
			name:          "missing nuon prefix",
			principal:     "component:database",
			expectedType:  "",
			expectedName:  "",
			expectError:   true,
			errorContains: "must start with 'nuon::'",
		},
		{
			name:          "empty string",
			principal:     "",
			expectedType:  "",
			expectedName:  "",
			expectError:   true,
			errorContains: "must start with 'nuon::'",
		},
		{
			name:         "only nuon prefix",
			principal:    "nuon::",
			expectedType: "",
			expectedName: "",
			expectError:  true,
		},
		{
			name:          "invalid prefix",
			principal:     "invalid::component:database",
			expectedType:  "",
			expectedName:  "",
			expectError:   true,
			errorContains: "must start with 'nuon::'",
		},
		{
			name:         "double colon in name",
			principal:    "nuon::component:db:primary",
			expectedType: "component",
			expectedName: "db:primary",
			expectError:  false,
		},
		{
			name:         "trailing colon",
			principal:    "nuon::component:",
			expectedType: "component",
			expectedName: "",
			expectError:  false,
		},
		{
			name:         "multiple colons in name",
			principal:    "nuon::action:step:1:init",
			expectedType: "action",
			expectedName: "step:1:init",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParsePrincipal(tt.principal)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
				return
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
			}

			if string(p.Type) != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, p.Type)
			}

			if p.Name != tt.expectedName {
				t.Errorf("expected name %q, got %q", tt.expectedName, p.Name)
			}
		})
	}
}
