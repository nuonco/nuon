package handlers

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/parser/toml"
)

func TestBuildTableFoldingRanges(t *testing.T) {
	// Nuon-style config with components and nested tables
	config := `version = "v1"
description = "test app"

[installer]
name = "My App"
description = "A demo app"
documentation_url = "https://docs.nuon.co/"

[sandbox]
terraform_version = "1.7.5"

[sandbox.public_repo]
directory = "aws-ecs"
repo = "nuonco/sandboxes"
branch = "main"

[[components]]
name = "ecs_service"
type = "terraform_module"

[components.public_repo]
repo = "nuonco/guides"
directory = "aws-ecs-tutorial"

[[components]]
name = "docker_image"
type = "docker_build"
dockerfile = "Dockerfile"
`
	lines := strings.Split(config, "\n")
	doc := toml.ParseToml(config)

	ranges := buildTableFoldingRanges(doc.Tables, lines)

	if len(ranges) == 0 {
		t.Error("expected folding ranges for tables, got none")
	}

	// Verify we have ranges for: installer, sandbox, sandbox.public_repo,
	// components (x2), components.public_repo
	tableNames := []string{}
	for _, table := range doc.Tables {
		tableNames = append(tableNames, table.Name)
	}
	t.Logf("Found tables: %v", tableNames)
	t.Logf("Found %d folding ranges", len(ranges))

	for i, r := range ranges {
		t.Logf("Range %d: lines %d-%d", i, r.StartLine, r.EndLine)
	}
}

func TestFindMultiLineStringRanges(t *testing.T) {
	// Config with multi-line strings (common in Nuon for markdown)
	config := `[installer]
post_install_markdown = """
# My App

This is a multi-line
markdown description.
"""

copyright_markdown = """
© 2024 Nuon.
"""

name = "single line"
`
	lines := strings.Split(config, "\n")
	ranges := findMultiLineStringRanges(lines)

	if len(ranges) != 2 {
		t.Errorf("expected 2 multi-line string ranges, got %d", len(ranges))
	}

	for i, r := range ranges {
		t.Logf("Multi-line string range %d: lines %d-%d", i, r.StartLine, r.EndLine)
	}
}

func TestFindCommentBlockRanges(t *testing.T) {
	config := `# This is a Nuon configuration file
# It defines an application with components
# See https://docs.nuon.co for more info

version = "v1"

# Component section
[[components]]
name = "web"
`
	lines := strings.Split(config, "\n")
	ranges := findCommentBlockRanges(lines)

	if len(ranges) != 1 {
		t.Errorf("expected 1 comment block range, got %d", len(ranges))
	}

	if len(ranges) > 0 {
		r := ranges[0]
		if r.StartLine != 0 || r.EndLine != 2 {
			t.Errorf("expected comment block at lines 0-2, got %d-%d", r.StartLine, r.EndLine)
		}
		t.Logf("Comment block range: lines %d-%d", r.StartLine, r.EndLine)
	}
}

func TestNuonFullConfig(t *testing.T) {
	// Real-world Nuon config structure
	config := `#:schema https://api.nuon.co/v1/general/config-schema
# This file contains template values for common Nuon application configuration options.
# To use it for your app, edit as needed, then rename this file and run:
#   nuon apps sync -c app.toml

version = "v1"
description = "template with sources"
display_name = "template-app"

[installer]
name = "installer"
description = "one click installer"
documentation_url = "docs-url"

[runner]
runner_type = "aws-ecs"

[sandbox]
terraform_version = "1.7.5"

[sandbox.public_repo]
directory = "aws-ecs-byovpc"
repo = "nuonco/sandboxes"
branch = "main"

[[components]]
name = "toml_terraform"
type = "terraform_module"
terraform_version = "1.7.5"

[components.connected_repo]
directory = "infra"
repo = "nuonco/nuon"
branch = "main"

[components.vars]
AWS_REGION = "{{.nuon.install.sandbox.account.region}}"

[[components]]
name = "toml_helm"
type = "helm_chart"
chart_name = "e2e-helm"

[components.connected_repo]
directory = "deployment"
repo = "nuonco/nuon"

[[values_file]]
contents = """
image.tag = {{.nuon.components.toml_docker_build.image.name}}
"""

[[input]]
name = "vpc_id"
description = "vpc_id to install application into"
default = ""
sensitive = false

[[input]]
name = "api_key"
description = "API key"
default = ""
sensitive = true
`
	lines := strings.Split(config, "\n")
	doc := toml.ParseToml(config)

	tableRanges := buildTableFoldingRanges(doc.Tables, lines)
	stringRanges := findMultiLineStringRanges(lines)
	commentRanges := findCommentBlockRanges(lines)

	totalRanges := len(tableRanges) + len(stringRanges) + len(commentRanges)

	t.Logf("Full Nuon config analysis:")
	t.Logf("  Tables found: %d", len(doc.Tables))
	t.Logf("  Table folding ranges: %d", len(tableRanges))
	t.Logf("  Multi-line string ranges: %d", len(stringRanges))
	t.Logf("  Comment block ranges: %d", len(commentRanges))
	t.Logf("  Total folding ranges: %d", totalRanges)

	if totalRanges == 0 {
		t.Error("expected some folding ranges for full Nuon config")
	}

	// Should have ranges for: installer, runner, sandbox, sandbox.public_repo,
	// components (x2), components.connected_repo (x2), components.vars,
	// values_file, input (x2)
	if len(doc.Tables) < 10 {
		t.Errorf("expected at least 10 tables in full config, got %d", len(doc.Tables))
	}
}
