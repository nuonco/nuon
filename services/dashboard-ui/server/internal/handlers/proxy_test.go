package handlers

import (
	"encoding/json"
	"testing"
)

func TestRewriteSwaggerSpecV2(t *testing.T) {
	tests := []struct {
		name         string
		apiPrefix    string
		wantBasePath string
	}{
		{name: "admin routes through /admin", apiPrefix: "/admin", wantBasePath: "/admin"},
		{name: "public routes same-origin", apiPrefix: "", wantBasePath: "/"},
	}

	spec := []byte(`{
		"swagger": "2.0",
		"host": "ctl.nuon.us-west-2.prod.internal.nuon.co",
		"schemes": ["http"],
		"basePath": "/",
		"paths": {"/v1/installs": {}}
	}`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := rewriteSwaggerSpec(spec, tt.apiPrefix)
			if err != nil {
				t.Fatalf("rewriteSwaggerSpec: %v", err)
			}

			var doc map[string]any
			if err := json.Unmarshal(out, &doc); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if _, ok := doc["host"]; ok {
				t.Errorf("host should be dropped, got %v", doc["host"])
			}
			if _, ok := doc["schemes"]; ok {
				t.Errorf("schemes should be dropped, got %v", doc["schemes"])
			}
			if got := doc["basePath"]; got != tt.wantBasePath {
				t.Errorf("basePath = %v, want %v", got, tt.wantBasePath)
			}
		})
	}
}

func TestRewriteSwaggerSpecV3(t *testing.T) {
	spec := []byte(`{
		"openapi": "3.0.0",
		"servers": [{"url": "http://ctl.nuon.us-west-2.prod.internal.nuon.co"}],
		"paths": {"/v1/installs": {}}
	}`)

	out, err := rewriteSwaggerSpec(spec, "/admin")
	if err != nil {
		t.Fatalf("rewriteSwaggerSpec: %v", err)
	}

	var doc struct {
		Servers []struct {
			URL string `json:"url"`
		} `json:"servers"`
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(doc.Servers) != 1 || doc.Servers[0].URL != "/admin" {
		t.Errorf("servers = %+v, want single url /admin", doc.Servers)
	}
}
