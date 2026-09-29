package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGo(t *testing.T) {
	tests := map[string]struct {
		src   string
		count int
		fixed string
	}{
		"swagger block": {
			src: "package a\n\n// @ID GetThing\n// @Summary get a thing\n// @Router /v1/things [get]\nfunc F() {}\n",
		},
		"temporal annotations": {
			src: "package a\n\n// @temporal-gen-v2 activity\n// @start-to-close-timeout 5m\nfunc F() {}\n",
		},
		"directives": {
			src: "//go:build linux\n\npackage a\n\n//go:generate echo\nfunc F() {\n\t_ = 1 //nolint:gosec // matches upstream\n}\n",
		},
		"why and todo": {
			src: "package a\n\nfunc F() {\n\t// why: continue-as-new can run this twice\n\t_ = 1\n\t// TODO(jm): remove\n\t_ = 2\n}\n",
		},
		"field docs allowed without specs": {
			src: "package a\n\n// Thing is a thing.\ntype Thing struct {\n\t// Name of the thing.\n\tName string\n}\n",
		},
		"body narration": {
			src:   "package a\n\nfunc F() {\n\t// Fetch the installs\n\t_ = 1\n\n\t// Check if steps already exist\n\t_ = 2 // trailing\n}\n",
			count: 3,
			fixed: "package a\n\nfunc F() {\n\t_ = 1\n\n\t_ = 2\n}\n",
		},
		"func doc narration": {
			src:   "package a\n\n// F does a thing.\nfunc F() {}\n",
			count: 1,
			fixed: "package a\n\nfunc F() {}\n",
		},
		"deprecated kept": {
			src:   "package a\n\n// F does a thing.\n//\n// Deprecated: use G.\nfunc F() {}\n",
			count: 2,
			fixed: "package a\n\n// Deprecated: use G.\nfunc F() {}\n",
		},
		"example output": {
			src: "package a\n\nfunc ExampleF() {\n\tprintln(1)\n\t// Output: 1\n}\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			vs, fixed, err := checkGo("a.go", []byte(tc.src), options{})
			require.NoError(t, err)
			require.Len(t, vs, tc.count)
			if tc.fixed != "" {
				require.Equal(t, tc.fixed, string(fixed))
			}
		})
	}
}

func TestGoSwaggerTypes(t *testing.T) {
	src := "package app\n\n// Thing doc.\ntype Thing struct {\n\t// Name doc.\n\tName string\n}\n\n// Other doc.\ntype Other struct {\n\t// X doc.\n\tX int\n}\n"
	vs, _, err := checkGo("a.go", []byte(src), options{swaggerTypes: map[string]bool{"app.Thing": true}})
	require.NoError(t, err)
	require.Len(t, vs, 2)
	require.Equal(t, "// Other doc.", vs[0].text)
}

func TestJS(t *testing.T) {
	tests := map[string]struct {
		src   string
		count int
		fixed string
	}{
		"strings and urls": {src: "const a = 'https://x.co'\nconst b = \"http://x.co\"\nconst t = `a//b ${x} c`\n"},
		"regex":            {src: "const r = /https?:\\/\\//g\nconst d = a / b / c\n"},
		"jsx url text":     {src: "const a = <p>see https://docs.example.com</p>\n"},
		"jsx slashes":      {src: "const a = <span className=\"font-mono\">//nuon.co/ctl-api</span>\n"},
		"multibyte ending": {
			src:   "const a = 1 // wait…\n// ok — done\nconst b = 2\n",
			count: 2,
			fixed: "const a = 1\nconst b = 2\n",
		},
		"jsx apostrophe": {src: "const a = <p>Don't do it</p>\nconst b = 'https://x'\n"},
		"directives":     {src: "// eslint-disable-next-line no-console\n// @ts-expect-error stand-in\n/// <reference types=\"x\" />\n"},
		"why":            {src: "// why: posthog touches the DOM at import time\n// and throws under the test DOM\nimport 'x'\n"},
		"narration": {
			src:   "// Fetch the installs\nconst a = 1 // trailing\n",
			count: 2,
			fixed: "const a = 1\n",
		},
		"jsx block comment": {
			src:   "return (\n  <div>\n    {/* Header */}\n    <h1 />\n  </div>\n)\n",
			count: 1,
			fixed: "return (\n  <div>\n    <h1 />\n  </div>\n)\n",
		},
		"jsdoc": {
			src:   "/**\n * Counts lines.\n */\nexport const f = 1\n",
			count: 1,
			fixed: "export const f = 1\n",
		},
		"blank line collapse": {
			src:   "const a = 1\n\n// section\n\nconst b = 2\n",
			count: 1,
			fixed: "const a = 1\n\nconst b = 2\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			vs, fixed, err := jsChecker(false)("a.tsx", []byte(tc.src), options{})
			require.NoError(t, err)
			require.Len(t, vs, tc.count)
			if tc.fixed != "" {
				require.Equal(t, tc.fixed, string(fixed))
			}
		})
	}
}

func TestCSS(t *testing.T) {
	src := "a { background: url(//cdn.example.com/x.png); }\n/* Card tables */\nb { color: red; }\n/* why: Lightning CSS minifies inherit to normal */\nc { color: blue; }\n"
	vs, fixed, err := jsChecker(true)("a.css", []byte(src), options{})
	require.NoError(t, err)
	require.Len(t, vs, 1)
	require.NotContains(t, string(fixed), "Card tables")
	require.Contains(t, string(fixed), "why: Lightning")
}

func TestHash(t *testing.T) {
	tests := map[string]struct {
		lang  lang
		src   string
		count int
	}{
		"shebang and shellcheck": {lang: langShell, src: "#!/bin/bash\n# shellcheck disable=SC2046\necho hi\n"},
		"shell narration":        {lang: langShell, src: "#!/bin/bash\n# ── parse args ──\necho hi\n", count: 1},
		"shell heredoc":          {lang: langShell, src: "cat <<EOF\n# not a comment\nEOF\n"},
		"yaml block scalar":      {lang: langYAML, src: "run: |\n  # shell comment in yaml string\n  echo hi\n# set elsewhere\nkey: v\n", count: 1},
		"python triple quote":    {lang: langPython, src: "CSS = \"\"\"\n#context { color: red }\n\"\"\"\n"},
		"dockerfile syntax":      {lang: langDocker, src: "# syntax=docker/dockerfile:1\n# install deps\nFROM x\n", count: 1},
		"why group":              {lang: langHCL, src: "# why: provider v5 drops this tag\n# on import\nresource \"x\" \"y\" {}\n"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			vs, fixed, err := hashChecker(tc.lang)("f", []byte(tc.src), options{})
			require.NoError(t, err)
			require.Len(t, vs, tc.count)
			if tc.count > 0 {
				require.Less(t, len(fixed), len(tc.src))
				require.False(t, strings.Contains(string(fixed), vs[0].text))
			}
		})
	}
}
