package main

import (
	"sort"
	"strings"
)

type edit struct {
	start, end int
	repl       string
}

func removeComments(src []byte, vs []violation, jsx bool) []byte {
	lines := strings.SplitAfter(string(src), "\n")
	starts := make([]int, len(lines))
	for i, off := 0, 0; i < len(lines); i++ {
		starts[i] = off
		off += len(lines[i])
	}
	lineAt := func(off int) int {
		return sort.Search(len(starts), func(i int) bool { return starts[i] > off }) - 1
	}

	removed := make([]bool, len(lines))
	edits := map[int][]edit{}

	for _, v := range vs {
		s, e := v.start, v.end
		if jsx {
			i := s - 1
			for i >= 0 && (src[i] == ' ' || src[i] == '\t') {
				i--
			}
			j := e
			for j < len(src) && (src[j] == ' ' || src[j] == '\t') {
				j++
			}
			if i >= 0 && j < len(src) && src[i] == '{' && src[j] == '}' {
				s, e = i, j+1
			}
		}

		first, last := lineAt(s), lineAt(max(e-1, s))
		lineStart := starts[first]
		lineEnd := starts[last] + len(strings.TrimRight(lines[last], "\r\n"))
		before := strings.TrimSpace(string(src[lineStart:s]))
		after := strings.TrimSpace(string(src[e:lineEnd]))

		switch {
		case before == "" && after == "":
			for l := first; l <= last; l++ {
				removed[l] = true
			}
		case first != last:
			continue
		case after == "":
			for s > lineStart && (src[s-1] == ' ' || src[s-1] == '\t') {
				s--
			}
			edits[first] = append(edits[first], edit{s - lineStart, e - lineStart, ""})
		default:
			edits[first] = append(edits[first], edit{s - lineStart, e - lineStart, " "})
		}
	}

	for l, es := range edits {
		if removed[l] {
			continue
		}
		sort.Slice(es, func(a, b int) bool { return es[a].start > es[b].start })
		line := lines[l]
		body := strings.TrimRight(line, "\r\n")
		nl := line[len(body):]
		for _, ed := range es {
			body = body[:ed.start] + ed.repl + body[ed.end:]
		}
		lines[l] = strings.TrimRight(body, " \t") + nl
	}

	var out []string
	for i := 0; i < len(lines); i++ {
		if !removed[i] {
			out = append(out, lines[i])
			continue
		}
		for i < len(lines) && removed[i] {
			i++
		}
		if i >= len(lines) {
			break
		}
		next := lines[i]
		prevBlank := len(out) > 0 && blank(out[len(out)-1])
		switch {
		case prevBlank && blank(next):
		case prevBlank && closer(next):
			out[len(out)-1] = next
		case (len(out) == 0 || opener(out[len(out)-1])) && blank(next):
		default:
			out = append(out, next)
		}
	}

	return []byte(strings.Join(out, ""))
}

func blank(line string) bool {
	return strings.TrimSpace(line) == ""
}

func opener(line string) bool {
	t := strings.TrimSpace(line)
	return strings.HasSuffix(t, "{") || strings.HasSuffix(t, "(") || strings.HasSuffix(t, "[")
}

func closer(line string) bool {
	t := strings.TrimSpace(line)
	for _, p := range []string{"}", ")", "]", "</", "fi", "done", "esac"} {
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	return false
}
