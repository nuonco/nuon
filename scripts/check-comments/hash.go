package main

import (
	"regexp"
	"strings"
)

type lang int

const (
	langShell lang = iota
	langYAML
	langPython
	langHCL
	langDocker
)

var hashDirective = regexp.MustCompile(`^#\s*(shellcheck|syntax=|escape=|check=|yaml-language-server:|nolint|noqa|type:|pylint:|pyright:|mypy:|ruff:|fmt:|isort:|pragma|tflint-ignore|checkov|tfsec:|trivy:|kics-scan|hadolint|prettier-ignore|-\*-|renovate:|@|--|nosec|shfmt:)`)

var (
	heredocStart = regexp.MustCompile(`<<-?\s*['"]?([A-Za-z_][A-Za-z0-9_]*)['"]?`)
	blockScalar  = regexp.MustCompile(`(^|[:\-]\s*)[|>][-+0-9]*\s*(#.*)?$`)
)

func hashChecker(l lang) checker {
	return func(path string, src []byte, o options) ([]violation, []byte, error) {
		lines := strings.SplitAfter(string(src), "\n")
		comment := make([]bool, len(lines))

		heredoc, heredocDash := "", false
		yamlBlock := -1
		inTriple := ""

		for i, raw := range lines {
			line := strings.TrimRight(raw, "\r\n")
			trimmed := strings.TrimSpace(line)
			indent := len(line) - len(strings.TrimLeft(line, " \t"))

			if heredoc != "" {
				check := line
				if heredocDash {
					check = strings.TrimLeft(line, "\t")
				}
				if check == heredoc {
					heredoc = ""
				}
				continue
			}
			if yamlBlock >= 0 {
				if trimmed == "" || indent > yamlBlock {
					continue
				}
				yamlBlock = -1
			}
			if inTriple != "" {
				if strings.Count(line, inTriple)%2 == 1 {
					inTriple = ""
				}
				continue
			}

			isComment := strings.HasPrefix(trimmed, "#") || (l == langHCL && strings.HasPrefix(trimmed, "//"))
			if isComment {
				comment[i] = true
				continue
			}

			switch l {
			case langShell, langDocker, langHCL:
				if m := heredocStart.FindStringSubmatch(line); m != nil && !strings.Contains(line, "<<<") {
					heredoc = m[1]
					heredocDash = strings.Contains(line, "<<-")
				}
			case langYAML:
				if blockScalar.MatchString(line) {
					yamlBlock = indent
				}
			case langPython:
				for _, q := range []string{`"""`, `'''`} {
					if strings.Count(line, q)%2 == 1 {
						inTriple = q
						break
					}
				}
			}
		}

		var vs []violation
		offset := 0
		for i := 0; i < len(lines); i++ {
			if !comment[i] {
				offset += len(lines[i])
				continue
			}
			j := i
			groupOffset := offset
			for j < len(lines) && comment[j] {
				j++
			}
			first := hashText(lines[i])
			keepAll := keepFirstLine(first)
			for k := i; k < j; k++ {
				if license(hashText(lines[k])) {
					keepAll = true
				}
			}
			for k := i; k < j; k++ {
				body := strings.TrimRight(lines[k], "\r\n")
				t := strings.TrimSpace(body)
				if keepAll || hashDirective.MatchString(t) || (k == 0 && strings.HasPrefix(t, "#!")) {
					groupOffset += len(lines[k])
					continue
				}
				start := groupOffset + strings.Index(body, t)
				vs = append(vs, violation{line: k + 1, text: t, start: start, end: start + len(t)})
				groupOffset += len(lines[k])
			}
			offset = groupOffset
			i = j - 1
		}
		if len(vs) == 0 {
			return nil, src, nil
		}
		return vs, removeComments(src, vs, false), nil
	}
}

func hashText(line string) string {
	t := strings.TrimSpace(line)
	t = strings.TrimPrefix(t, "#")
	t = strings.TrimPrefix(t, "//")
	return strings.TrimSpace(t)
}
