// Command signaldocs walks all signal packages under internal/app/*/signals/
// and generates an HTML page with Mermaid flowcharts and GitHub code links for
// each signal's Validate and Execute methods.
//
// Usage:
//
//	go run ./services/ctl-api/cmd/signaldocs -out signals.html
//	go run ./services/ctl-api/cmd/signaldocs -out signals.html -repo https://github.com/nuonco/nuon -ref main
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ----- data model -----------------------------------------------------------

type ActivityCall struct {
	Name string // e.g. "AwaitGetByRunnerID"
	Line int    // source line
	File string // relative path from repo root
}

type Branch struct {
	Condition string
	Steps     []ActivityCall
}

type SignalInfo struct {
	Domain      string         // e.g. "runners", "installs", "apps"
	Name        string         // e.g. "processinit", "provision"
	SignalType  string         // e.g. "process_init", "runner-provision"
	StructName  string         // always "Signal"
	Fields      []string       // e.g. ["RunnerID string", "ProcessID string"]
	Interfaces  []string       // e.g. ["SleepAfter"]
	ValidateAct []ActivityCall // activity calls in Validate
	ExecuteAct  []ActivityCall // activity calls in Execute (flat, in order)
	Branches    []Branch       // switch/if branches in Execute
	PkgDir      string         // relative path to signal package
	GoFiles     []string       // relative paths of .go files in the package
}

// ----- static analysis ------------------------------------------------------

var reAwait = regexp.MustCompile(`(?:activities|statusactivities|sharedactivities|kuberunner|workflowactivities|job|plan|client)\.\s*(Await\w+)`)

func extractActivityCalls(body *ast.BlockStmt, fset *token.FileSet, relFile string) []ActivityCall {
	if body == nil {
		return nil
	}
	var calls []ActivityCall
	ast.Inspect(body, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		name := sel.Sel.Name
		if strings.HasPrefix(name, "Await") {
			pos := fset.Position(ce.Pos())
			calls = append(calls, ActivityCall{
				Name: name,
				Line: pos.Line,
				File: relFile,
			})
		}
		return true
	})
	return calls
}

func parseSignalPackage(pkgDir, repoRoot string) (*SignalInfo, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgDir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", pkgDir, err)
	}

	relDir, _ := filepath.Rel(repoRoot, pkgDir)

	info := &SignalInfo{
		PkgDir: relDir,
	}

	// Collect Go files
	for _, pkg := range pkgs {
		for fname := range pkg.Files {
			rel, _ := filepath.Rel(repoRoot, fname)
			info.GoFiles = append(info.GoFiles, rel)
		}
	}
	sort.Strings(info.GoFiles)

	// Extract name from directory path: .../signals/v2/{name}
	parts := strings.Split(filepath.ToSlash(relDir), "/")
	info.Name = parts[len(parts)-1]

	// Find domain: .../internal/app/{domain}/signals/...
	for i, p := range parts {
		if p == "app" && i+1 < len(parts) {
			info.Domain = parts[i+1]
			break
		}
	}

	for _, pkg := range pkgs {
		for fname, file := range pkg.Files {
			relFile, _ := filepath.Rel(repoRoot, fname)

			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.GenDecl:
					for _, spec := range d.Specs {
						// Extract const SignalType
						vs, ok := spec.(*ast.ValueSpec)
						if !ok {
							continue
						}
						for i, name := range vs.Names {
							if name.Name == "SignalType" && i < len(vs.Values) {
								if lit, ok := vs.Values[i].(*ast.BasicLit); ok {
									info.SignalType = strings.Trim(lit.Value, `"`)
								}
							}
						}

						// Extract Signal struct fields
						ts, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						if ts.Name.Name == "Signal" {
							info.StructName = "Signal"
							if st, ok := ts.Type.(*ast.StructType); ok {
								for _, f := range st.Fields.List {
									if len(f.Names) > 0 && f.Names[0].IsExported() {
										info.Fields = append(info.Fields, fmt.Sprintf("%s %s", f.Names[0].Name, typeString(f.Type)))
									}
								}
							}
						}
					}

				case *ast.FuncDecl:
					if d.Recv == nil || len(d.Recv.List) == 0 {
						continue
					}

					switch d.Name.Name {
					case "Validate":
						info.ValidateAct = extractActivityCalls(d.Body, fset, relFile)
					case "Execute":
						info.ExecuteAct = extractActivityCalls(d.Body, fset, relFile)
					case "SleepAfter":
						info.Interfaces = append(info.Interfaces, "SleepAfter")
					}
				}
			}
		}
	}

	// Handle TypeSpec inside GenDecl (the above switch doesn't reach TypeSpec properly)
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				gd, ok := decl.(*ast.GenDecl)
				if !ok || gd.Tok != token.TYPE {
					continue
				}
				for _, spec := range gd.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if ts.Name.Name == "Signal" {
						info.StructName = "Signal"
						if st, ok := ts.Type.(*ast.StructType); ok {
							info.Fields = nil // reset in case already populated
							for _, f := range st.Fields.List {
								if len(f.Names) > 0 && f.Names[0].IsExported() {
									info.Fields = append(info.Fields, fmt.Sprintf("%s %s", f.Names[0].Name, typeString(f.Type)))
								}
							}
						}
					}
				}
			}
		}
	}

	return info, nil
}

func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return typeString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	case *ast.ArrayType:
		return "[]" + typeString(t.Elt)
	case *ast.MapType:
		return "map[" + typeString(t.Key) + "]" + typeString(t.Value)
	default:
		return "any"
	}
}

// ----- mermaid generation ---------------------------------------------------

func (s *SignalInfo) MermaidDiagram(repoURL, ref string) string {
	var b strings.Builder
	b.WriteString("graph TD\n")

	nodeID := 0
	nextID := func() string {
		nodeID++
		return fmt.Sprintf("n%d", nodeID)
	}

	startID := nextID()
	b.WriteString(fmt.Sprintf("    %s([\"📨 %s\"])\n", startID, s.SignalType))

	// Validate phase
	valID := nextID()
	b.WriteString(fmt.Sprintf("    %s --> %s{{\"Validate\"}}\n", startID, valID))

	prevID := valID
	for _, act := range s.ValidateAct {
		id := nextID()
		link := makeLink(repoURL, ref, act.File, act.Line)
		b.WriteString(fmt.Sprintf("    %s --> %s[\"%s\"]\n", prevID, id, act.Name))
		b.WriteString(fmt.Sprintf("    click %s \"%s\" _blank\n", id, link))
		prevID = id
	}

	// Execute phase
	execID := nextID()
	b.WriteString(fmt.Sprintf("    %s --> %s{{\"Execute\"}}\n", prevID, execID))

	prevID = execID
	for _, act := range s.ExecuteAct {
		id := nextID()
		link := makeLink(repoURL, ref, act.File, act.Line)
		b.WriteString(fmt.Sprintf("    %s --> %s[\"%s\"]\n", prevID, id, act.Name))
		b.WriteString(fmt.Sprintf("    click %s \"%s\" _blank\n", id, link))
		prevID = id
	}

	endID := nextID()
	b.WriteString(fmt.Sprintf("    %s --> %s([\"✅ Done\"])\n", prevID, endID))

	return b.String()
}

func makeLink(repoURL, ref, file string, line int) string {
	if repoURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/blob/%s/%s#L%d", repoURL, ref, file, line)
}

// ----- discovery ------------------------------------------------------------

func discoverSignalPackages(root string) ([]string, error) {
	var dirs []string

	signalsBase := filepath.Join(root, "services", "ctl-api", "internal", "app")

	err := filepath.Walk(signalsBase, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if !fi.IsDir() {
			return nil
		}

		// Match pattern: .../signals/v2/{name}/ or .../signals/{name}/
		rel, _ := filepath.Rel(signalsBase, path)
		parts := strings.Split(filepath.ToSlash(rel), "/")

		// Look for signals/v2/{name} pattern (leaf directory with signal.go)
		for i, p := range parts {
			if p == "signals" {
				// Check if this is a leaf signal dir (has signal.go)
				if _, err := os.Stat(filepath.Join(path, "signal.go")); err == nil {
					// Ensure it's not an "activities" directory
					if parts[len(parts)-1] != "activities" {
						dirs = append(dirs, path)
						return filepath.SkipDir
					}
				}
				// Skip sub-walk into activities dirs
				if i+1 < len(parts) && parts[len(parts)-1] == "activities" {
					return filepath.SkipDir
				}
			}
		}

		return nil
	})

	sort.Strings(dirs)
	return dirs, err
}

// ----- HTML template --------------------------------------------------------

const htmlTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Signal Documentation</title>
<script src="https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"></script>
<style>
  :root { --bg: #0d1117; --fg: #c9d1d9; --card: #161b22; --border: #30363d; --accent: #58a6ff; }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif; background: var(--bg); color: var(--fg); padding: 2rem; }
  h1 { font-size: 1.8rem; margin-bottom: 0.5rem; }
  .subtitle { color: #8b949e; margin-bottom: 2rem; }
  .toc { margin-bottom: 3rem; }
  .toc h2 { font-size: 1.2rem; margin-bottom: 0.75rem; }
  .toc ul { list-style: none; columns: 3; }
  .toc li { margin-bottom: 0.3rem; }
  .toc a { color: var(--accent); text-decoration: none; font-size: 0.9rem; }
  .toc a:hover { text-decoration: underline; }
  .domain-group { margin-bottom: 2rem; }
  .domain-group h3 { font-size: 1rem; color: #8b949e; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 0.5rem; border-bottom: 1px solid var(--border); padding-bottom: 0.3rem; }
  .signal-card { background: var(--card); border: 1px solid var(--border); border-radius: 8px; padding: 1.5rem; margin-bottom: 1.5rem; }
  .signal-card h2 { font-size: 1.3rem; margin-bottom: 0.5rem; display: flex; align-items: center; gap: 0.5rem; }
  .signal-card h2 .badge { background: #1f6feb; color: #fff; font-size: 0.7rem; padding: 2px 8px; border-radius: 12px; font-weight: 500; }
  .meta { font-size: 0.85rem; color: #8b949e; margin-bottom: 1rem; }
  .meta a { color: var(--accent); text-decoration: none; }
  .meta a:hover { text-decoration: underline; }
  .fields { display: flex; flex-wrap: wrap; gap: 0.5rem; margin-bottom: 1rem; }
  .field { background: #21262d; padding: 3px 10px; border-radius: 4px; font-family: 'SF Mono', Monaco, monospace; font-size: 0.8rem; }
  .diagram-container { background: #fff; border-radius: 6px; padding: 1rem; overflow-x: auto; }
  .files { margin-top: 1rem; font-size: 0.8rem; }
  .files a { color: var(--accent); text-decoration: none; margin-right: 1rem; }
  .files a:hover { text-decoration: underline; }
  .filter-bar { margin-bottom: 1.5rem; }
  .filter-bar input { background: var(--card); border: 1px solid var(--border); color: var(--fg); padding: 8px 14px; border-radius: 6px; width: 100%; max-width: 400px; font-size: 0.95rem; }
  .filter-bar input::placeholder { color: #484f58; }
</style>
</head>
<body>
<h1>⚡ Signal & Handler Documentation</h1>
<p class="subtitle">Auto-generated from <code>services/ctl-api/internal/app/*/signals/</code> — {{len .Signals}} signals across {{.DomainCount}} domains</p>

<div class="filter-bar">
  <input type="text" id="filter" placeholder="Filter signals..." oninput="filterSignals(this.value)">
</div>

<div class="toc">
  <h2>Table of Contents</h2>
  {{range .Domains}}
  <div class="domain-group">
    <h3>{{.Name}}</h3>
    <ul>
      {{range .Signals}}
      <li><a href="#{{.Domain}}-{{.Name}}">{{.SignalType}}</a></li>
      {{end}}
    </ul>
  </div>
  {{end}}
</div>

{{range .Domains}}
<h2 style="font-size:1.4rem; margin: 2rem 0 1rem; border-bottom: 1px solid #30363d; padding-bottom: 0.5rem;">📦 {{.Name}}</h2>
{{range .Signals}}
<div class="signal-card" id="{{.Domain}}-{{.Name}}" data-signal-type="{{.SignalType}}" data-domain="{{.Domain}}" data-name="{{.Name}}">
  <h2>
    {{.SignalType}}
    <span class="badge">{{.Domain}}</span>
    {{range .Interfaces}}<span class="badge" style="background:#238636">{{.}}</span>{{end}}
  </h2>
  <div class="meta">
    <a href="{{$.RepoURL}}/tree/{{$.Ref}}/{{.PkgDir}}" target="_blank">{{.PkgDir}}</a>
  </div>
  {{if .Fields}}
  <div class="fields">
    {{range .Fields}}<span class="field">{{.}}</span>{{end}}
  </div>
  {{end}}
  <div class="diagram-container">
    <pre class="mermaid">
{{.MermaidDiagram $.RepoURL $.Ref}}
    </pre>
  </div>
  <div class="files">
    {{range .GoFiles}}<a href="{{$.RepoURL}}/blob/{{$.Ref}}/{{.}}" target="_blank">{{.}}</a>{{end}}
  </div>
</div>
{{end}}
{{end}}

<script>
  mermaid.initialize({ startOnLoad: true, theme: 'default', securityLevel: 'loose' });
  function filterSignals(q) {
    q = q.toLowerCase();
    document.querySelectorAll('.signal-card').forEach(card => {
      const t = (card.dataset.signalType + ' ' + card.dataset.domain + ' ' + card.dataset.name).toLowerCase();
      card.style.display = t.includes(q) ? '' : 'none';
    });
  }
</script>
</body>
</html>
`

// ----- template data --------------------------------------------------------

type DomainGroup struct {
	Name    string
	Signals []*SignalInfo
}

type TemplateData struct {
	Signals     []*SignalInfo
	Domains     []DomainGroup
	DomainCount int
	RepoURL     string
	Ref         string
}

// ----- main -----------------------------------------------------------------

func main() {
	outFile := flag.String("out", "signals.html", "output HTML file")
	repoURL := flag.String("repo", "https://github.com/nuonco/nuon", "GitHub repo URL")
	ref := flag.String("ref", "main", "git ref for code links")
	flag.Parse()

	// Find repo root (directory containing go.mod)
	repoRoot, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	dirs, err := discoverSignalPackages(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error discovering signals: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Found %d signal packages\n", len(dirs))

	var signals []*SignalInfo
	for _, dir := range dirs {
		info, err := parseSignalPackage(dir, repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  WARN: %v\n", err)
			continue
		}
		if info.SignalType == "" {
			fmt.Fprintf(os.Stderr, "  SKIP: %s (no SignalType const)\n", dir)
			continue
		}
		signals = append(signals, info)
		fmt.Fprintf(os.Stderr, "  ✓ %s/%s (%s) — %d validate, %d execute activities\n",
			info.Domain, info.Name, info.SignalType,
			len(info.ValidateAct), len(info.ExecuteAct))
	}

	// Group by domain
	domainMap := map[string][]*SignalInfo{}
	for _, s := range signals {
		domainMap[s.Domain] = append(domainMap[s.Domain], s)
	}
	var domains []DomainGroup
	for name, sigs := range domainMap {
		sort.Slice(sigs, func(i, j int) bool { return sigs[i].SignalType < sigs[j].SignalType })
		domains = append(domains, DomainGroup{Name: name, Signals: sigs})
	}
	sort.Slice(domains, func(i, j int) bool { return domains[i].Name < domains[j].Name })

	data := TemplateData{
		Signals:     signals,
		Domains:     domains,
		DomainCount: len(domains),
		RepoURL:     strings.TrimRight(*repoURL, "/"),
		Ref:         *ref,
	}

	tmpl, err := template.New("signals").Parse(htmlTmpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "template error: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Create(*outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create output: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		fmt.Fprintf(os.Stderr, "template execute: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "\n✅ Generated %s with %d signals across %d domains\n", *outFile, len(signals), len(domains))
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
