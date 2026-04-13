//go:build ignore

package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/tools/imports"
)

func collapseImportGroups(src []byte) []byte {
	lines := strings.Split(string(src), "\n")
	var out []string
	inImport := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inImport && (trimmed == "import (" || strings.HasPrefix(trimmed, "import (")) {
			inImport = true
			out = append(out, line)
			continue
		}
		if inImport {
			if trimmed == ")" {
				inImport = false
				out = append(out, line)
				continue
			}
			if trimmed == "" {
				continue
			}
			out = append(out, line)
			continue
		}
		out = append(out, line)
	}
	return []byte(strings.Join(out, "\n"))
}

func main() {
	path := os.Args[1]
	src, _ := os.ReadFile(path)
	collapsed := collapseImportGroups(src)
	result, err := imports.Process(path, collapsed, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(string(result))
}
