package main

import (
	"regexp"
	"strings"
)

var jsDirective = regexp.MustCompile(`^(//|/\*\*?)\s*(eslint|oxlint|@ts-|prettier-ignore|biome-ignore|istanbul|c8 |stylelint|@vite-ignore|webpack|#__PURE__|@__PURE__|@jsx|# sourceMappingURL|@license|@preserve)|^///\s*<reference|^/\*!`)

const regexPreceders = "(,=:[!&|?{};+-*%~^<>"

var regexKeywords = map[string]bool{
	"return": true, "typeof": true, "case": true, "in": true, "of": true, "new": true,
	"delete": true, "void": true, "instanceof": true, "yield": true, "await": true,
}

type jsComment struct {
	start, end int
}

func skipQuoted(src []byte, i int, q byte) int {
	i++
	for i < len(src) && src[i] != q && src[i] != '\n' {
		if src[i] == '\\' {
			i++
		}
		i++
	}
	return i + 1
}

func skipTemplate(src []byte, i int) int {
	i++
	for i < len(src) {
		switch {
		case src[i] == '\\':
			i += 2
			continue
		case src[i] == '`':
			return i + 1
		case src[i] == '$' && i+1 < len(src) && src[i+1] == '{':
			depth := 1
			i += 2
			for i < len(src) && depth > 0 {
				switch c := src[i]; c {
				case '{':
					depth++
				case '}':
					depth--
				case '\'', '"':
					i = skipQuoted(src, i, c) - 1
				case '`':
					i = skipTemplate(src, i) - 1
				}
				i++
			}
			continue
		}
		i++
	}
	return i
}

func skipRegex(src []byte, i int) int {
	i++
	inClass := false
	for i < len(src) {
		switch src[i] {
		case '\\':
			i += 2
			continue
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '\n':
			return i
		case '/':
			if !inClass {
				return i + 1
			}
		}
		i++
	}
	return i
}

func findJSComments(src []byte, css bool) []jsComment {
	var out []jsComment
	prevChar, prevWord := byte(0), ""
	for i := 0; i < len(src); {
		c := src[i]
		var next byte
		if i+1 < len(src) {
			next = src[i+1]
		}
		switch {
		case !css && c == '/' && next == '/' && (i == 0 || src[i-1] == ' ' || src[i-1] == '\t' || src[i-1] == '\n'):
			end := i
			for end < len(src) && src[end] != '\n' {
				end++
			}
			stop := end
			for stop > i+2 && (src[stop-1] == ' ' || src[stop-1] == '\t' || src[stop-1] == '\r') {
				stop--
			}
			out = append(out, jsComment{i, stop})
			i = end
		case c == '/' && next == '*':
			end := strings.Index(string(src[i+2:]), "*/")
			stop := len(src)
			if end >= 0 {
				stop = i + 2 + end + 2
			}
			out = append(out, jsComment{i, stop})
			i = stop
		case c == '\'' || c == '"':
			i = skipQuoted(src, i, c)
			prevChar, prevWord = 'x', ""
		case !css && c == '`':
			i = skipTemplate(src, i)
			prevChar, prevWord = 'x', ""
		case !css && c == '/' && (prevChar == 0 || strings.IndexByte(regexPreceders, prevChar) >= 0 || regexKeywords[prevWord]):
			i = skipRegex(src, i)
			prevChar, prevWord = 'x', ""
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '_' || c == '$' || (c|0x20 >= 'a' && c|0x20 <= 'z'):
			j := i
			for j < len(src) && (src[j] == '_' || src[j] == '$' || (src[j]|0x20 >= 'a' && src[j]|0x20 <= 'z') || (src[j] >= '0' && src[j] <= '9')) {
				j++
			}
			prevWord, prevChar = string(src[i:j]), 'x'
			i = j
		default:
			prevChar, prevWord = c, ""
			i++
		}
	}
	return out
}

func jsChecker(css bool) checker {
	return func(path string, src []byte, o options) ([]violation, []byte, error) {
		comments := findJSComments(src, css)
		var vs []violation
		for gi := 0; gi < len(comments); {
			gj := gi + 1
			for gj < len(comments) && strings.HasPrefix(string(src[comments[gj].start:]), "//") &&
				strings.TrimSpace(string(src[comments[gj-1].end:comments[gj].start])) == "" &&
				strings.Count(string(src[comments[gj-1].end:comments[gj].start]), "\n") == 1 {
				gj++
			}
			first := string(src[comments[gi].start:comments[gi].end])
			keepAll := keepFirstLine(strings.TrimLeft(first, "/*! \t\r\n"))
			for k := gi; k < gj; k++ {
				if license(string(src[comments[k].start:comments[k].end])) {
					keepAll = true
				}
			}
			for k := gi; k < gj; k++ {
				text := string(src[comments[k].start:comments[k].end])
				if keepAll || jsDirective.MatchString(text) {
					continue
				}
				vs = append(vs, violation{
					line:  lineOf(src, comments[k].start),
					text:  text,
					start: comments[k].start,
					end:   comments[k].end,
				})
			}
			gi = gj
		}
		if len(vs) == 0 {
			return nil, src, nil
		}
		return vs, removeComments(src, vs, !css), nil
	}
}
