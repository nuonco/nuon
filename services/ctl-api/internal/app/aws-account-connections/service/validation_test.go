package service

import "testing"

func TestValidation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		ok   bool
	}{
		{name: "account", err: validateAccountID("123456789012"), ok: true},
		{name: "short account", err: validateAccountID("123"), ok: false},
		{name: "region west", err: validateRegion("us-west-2"), ok: true},
		{name: "region east", err: validateRegion("us-east-1"), ok: true},
		{name: "arbitrary region", err: validateRegion("somewhere-1"), ok: false},
		{name: "role", err: validateRoleARN("arn:aws:iam::123456789012:role/customer/nuon", "123456789012"), ok: true},
		{name: "user", err: validateRoleARN("arn:aws:iam::123456789012:user/nuon", "123456789012"), ok: false},
		{name: "root", err: validateRoleARN("arn:aws:iam::123456789012:root", "123456789012"), ok: false},
		{name: "sts", err: validateRoleARN("arn:aws:sts::123456789012:assumed-role/nuon/session", "123456789012"), ok: false},
		{name: "mismatch", err: validateRoleARN("arn:aws:iam::210987654321:role/nuon", "123456789012"), ok: false},
		{name: "wildcard", err: validateRoleARN("arn:aws:iam::123456789012:role/*", "123456789012"), ok: false},
		{name: "management role", err: validateManagementRoleARN("arn:aws:iam::123456789012:role/nuon-management"), ok: true},
		{name: "management wildcard", err: validateManagementRoleARN("arn:aws:iam::123456789012:role/*"), ok: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if (test.err == nil) != test.ok {
				t.Fatalf("unexpected error: %v", test.err)
			}
		})
	}
}
