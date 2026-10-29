// Command formatgen gofmts swagger-generated Go and drops imports the
// templates emitted but the file does not reference. It also adds a short
// list of standard-library imports the templates reference without declaring.
//
// go-swagger does this with golang.org/x/tools/imports, one process-wide
// lookup per file. These generated files only need a syntactic pass.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// stdlibByName is the package name to import path for packages the swagger
// templates reference without always importing. Names match the standard
// library package name.
var stdlibByName = map[string]string{
	"bytes":    "bytes",
	"context":  "context",
	"json":     "encoding/json",
	"fmt":      "fmt",
	"gzip":     "compress/gzip",
	"http":     "net/http",
	"httptest": "net/http/httptest",
	"io":       "io",
	"strconv":  "strconv",
	"strings":  "strings",
	"testing":  "testing",
	"time":     "time",
	"url":      "net/url",
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: formatgen dir...")
		os.Exit(2)
	}

	var files []string
	for _, root := range flag.Args() {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || filepath.Ext(path) != ".go" {
				return nil
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
		sem  = make(chan struct{}, 8)
	)
	for _, path := range files {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if err := formatFile(path); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

type imp struct {
	name string // local name, empty when it matches the package name
	path string
}

func formatFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	specs := importsOf(file)
	used := referencedSelectors(file)
	specs = addSiblingPackages(path, specs, used)
	specs = fixImports(specs, used)
	module := ""
	if mod, _, err := findModule(filepath.Dir(path)); err == nil {
		module = mod + "/"
	}
	rendered := renderImports(specs, module)

	next, err := replaceImports(src, rendered)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	formatted, err := format.Source(next)
	if err != nil {
		return fmt.Errorf("format %s: %w\n%s", path, err, next)
	}
	if bytes.Equal(src, formatted) {
		return nil
	}
	return os.WriteFile(path, formatted, 0o644)
}

func importsOf(file *ast.File) []imp {
	var out []imp
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			continue
		}
		for _, spec := range gen.Specs {
			im := spec.(*ast.ImportSpec)
			item := imp{path: strings.Trim(im.Path.Value, `"`)}
			if im.Name != nil {
				item.name = im.Name.Name
			}
			out = append(out, item)
		}
	}
	return out
}

func referencedSelectors(file *ast.File) map[string]struct{} {
	used := map[string]struct{}{}
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok || id.Obj != nil {
			return true
		}
		used[id.Name] = struct{}{}
		return true
	})
	return used
}

func localName(item imp) string {
	if item.name != "" && item.name != "_" && item.name != "." {
		return item.name
	}
	if item.name == "_" || item.name == "." {
		return item.name
	}
	parts := strings.Split(item.path, "/")
	return parts[len(parts)-1]
}

// addSiblingPackages imports a child directory whose name is used as a
// package selector but was not declared. The client facade references the
// operations package this way; goimports discovers it by loading the module.
func addSiblingPackages(path string, specs []imp, used map[string]struct{}) []imp {
	known := map[string]struct{}{}
	for _, item := range specs {
		known[localName(item)] = struct{}{}
	}
	dir := filepath.Dir(path)
	mod, root, err := findModule(dir)
	if err != nil {
		return specs
	}
	for name := range used {
		if _, ok := known[name]; ok {
			continue
		}
		if _, ok := stdlibByName[name]; ok {
			continue
		}
		sibling := filepath.Join(dir, name)
		info, err := os.Stat(sibling)
		if err != nil || !info.IsDir() {
			continue
		}
		rel, err := filepath.Rel(root, sibling)
		if err != nil {
			continue
		}
		specs = append(specs, imp{path: mod + "/" + filepath.ToSlash(rel)})
		known[name] = struct{}{}
	}
	return specs
}

func findModule(dir string) (mod, root string, err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", "", err
	}
	for {
		gomod := filepath.Join(dir, "go.mod")
		data, readErr := os.ReadFile(gomod)
		if readErr == nil {
			for _, line := range strings.Split(string(data), "\n") {
				fields := strings.Fields(line)
				if len(fields) >= 2 && fields[0] == "module" {
					return fields[1], dir, nil
				}
			}
			return "", "", fmt.Errorf("no module line in %s", gomod)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", fmt.Errorf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

func fixImports(specs []imp, used map[string]struct{}) []imp {
	known := map[string]struct{}{}
	var kept []imp
	seen := map[string]struct{}{}
	for _, item := range specs {
		name := localName(item)
		known[name] = struct{}{}
		if name == "_" || name == "." {
			kept = append(kept, item)
			continue
		}
		if _, ok := used[name]; !ok {
			continue
		}
		key := name + " " + item.path
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		kept = append(kept, item)
	}

	for name := range used {
		if _, ok := known[name]; ok {
			continue
		}
		path, ok := stdlibByName[name]
		if !ok {
			continue
		}
		kept = append(kept, imp{path: path})
	}
	return kept
}

func renderImports(specs []imp, module string) string {
	if len(specs) == 0 {
		return ""
	}
	if len(specs) == 1 {
		item := specs[0]
		if item.name == "" {
			return "import " + strconvQuote(item.path) + "\n"
		}
		return "import " + item.name + " " + strconvQuote(item.path) + "\n"
	}
	var std, third, local []imp
	for _, item := range specs {
		switch {
		case module != "" && strings.HasPrefix(item.path, module):
			local = append(local, item)
		case strings.Contains(strings.Split(item.path, "/")[0], "."):
			third = append(third, item)
		default:
			std = append(std, item)
		}
	}
	less := func(items []imp) {
		sort.Slice(items, func(i, j int) bool {
			if items[i].path != items[j].path {
				return items[i].path < items[j].path
			}
			return items[i].name < items[j].name
		})
	}
	less(std)
	less(third)
	less(local)

	var b strings.Builder
	b.WriteString("import (\n")
	groups := [][]imp{std, third, local}
	first := true
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}
		if !first {
			b.WriteByte('\n')
		}
		first = false
		for _, item := range group {
			b.WriteByte('\t')
			if item.name != "" {
				b.WriteString(item.name)
				b.WriteByte(' ')
			}
			b.WriteString(strconvQuote(item.path))
			b.WriteByte('\n')
		}
	}
	b.WriteString(")\n")
	return b.String()
}

func strconvQuote(s string) string {
	return `"` + s + `"`
}

// replaceImports swaps the first import declaration for rendered.
// rendered is empty when the file should have no imports.
func replaceImports(src []byte, rendered string) ([]byte, error) {
	text := string(src)
	start := strings.Index(text, "\nimport ")
	if start == -1 {
		if rendered == "" {
			return src, nil
		}
		pkg := strings.Index(text, "\npackage ")
		if pkg == -1 {
			return nil, fmt.Errorf("no package clause")
		}
		nl := strings.Index(text[pkg+1:], "\n")
		if nl == -1 {
			return nil, fmt.Errorf("no newline after package")
		}
		at := pkg + 1 + nl + 1
		return []byte(text[:at] + "\n" + rendered + text[at:]), nil
	}
	start++ // keep the newline, point at "import"
	rest := text[start:]
	var end int
	if strings.HasPrefix(rest, "import (") {
		rel := strings.Index(rest, "\n)")
		if rel == -1 {
			return nil, fmt.Errorf("unclosed import block")
		}
		end = start + rel + len("\n)")
	} else {
		rel := strings.Index(rest, "\n")
		if rel == -1 {
			end = len(text)
		} else {
			end = start + rel
		}
	}
	if rendered == "" {
		return []byte(text[:start] + text[end:]), nil
	}
	return []byte(text[:start] + rendered + text[end:]), nil
}
