package allsignals

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// A signal package registers itself in an init(), which only runs if something imports
// it. This file is that something — so a new signal package that is not listed here is
// simply absent from the catalog at runtime, and the queue workflow fails to handle it
// with an error that says nothing about the real cause. Nothing else catches it: the
// code compiles, the tests pass, and the endpoint that enqueues the signal returns 200.
//
// So walk the source tree for signal packages and assert each one is imported here.
func TestEverySignalPackageIsImported(t *testing.T) {
	root := repoRelative(t, "services/ctl-api/internal/app")

	found := map[string]string{} // import path -> dir
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return err
		}
		if !strings.Contains(filepath.ToSlash(path), "/signals/") {
			return nil
		}
		if declaresSignalType(t, path) {
			rel, rerr := filepath.Rel(root, path)
			if rerr != nil {
				return rerr
			}
			imp := "github.com/nuonco/nuon/services/ctl-api/internal/app/" + filepath.ToSlash(rel)
			found[imp] = path
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walking signal packages: %v", err)
	}
	if len(found) == 0 {
		t.Fatal("found no signal packages — the walk is broken, not the imports")
	}

	imported := importsOfThisPackage(t)
	for imp := range found {
		if imported[imp] || registeredTransitively[imp] {
			continue
		}
		t.Errorf("signal package %s is not imported in allsignals.go, so its init() only runs in "+
			"binaries that happen to import it — add it here, or to registeredTransitively with "+
			"the importer that guarantees registration", imp)
	}
}

// Signal packages that are absent from allsignals.go but still reach the catalog because
// a service or fx module imports them directly. Listed rather than fixed because moving
// them is a behaviour change to binaries that currently link them for other reasons; the
// point of this list is that a *new* omission stands out instead of hiding among them.
var registeredTransitively = map[string]bool{
	// installs/service/sync_install_config.go, apps/service/trigger_install_config_sync.go
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/installconfigsync": true,
	// installs/service/admin_generate_state.go
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state/generatestate":        true,
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state/statepartialgenerate": true,
	// apps/signals/branches/*
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/syncinstalls": true,
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/setuppreview": true,
	// fxmodules/workers_shared.go, notebooks/service/notebooks.go
	"github.com/nuonco/nuon/services/ctl-api/internal/app/notebooks/signals/start": true,
}

// The catalog is populated by this package's imports, so a non-empty catalog also proves
// the blank imports are doing what they claim.
func TestCatalogIsPopulated(t *testing.T) {
	if len(catalog.SignalCatalog) == 0 {
		t.Fatal("signal catalog is empty")
	}

	for _, typ := range []string{"install-phone-home-backfill", "org-phone-home-backfill"} {
		if _, ok := catalog.SignalCatalog[signal.SignalType(typ)]; !ok {
			t.Errorf("signal type %q is not in the catalog", typ)
		}
	}
}

// declaresSignalType reports whether a directory contains `SignalType` — the marker of a
// real signal package, as opposed to a shared helper directory under signals/.
func declaresSignalType(t *testing.T, dir string) bool {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		byts, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		if strings.Contains(string(byts), "SignalType signal.SignalType = ") {
			return true
		}
	}

	return false
}

func importsOfThisPackage(t *testing.T) map[string]bool {
	t.Helper()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "allsignals.go", nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parsing allsignals.go: %v", err)
	}

	out := map[string]bool{}
	for _, imp := range f.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		out[path] = true
	}
	return out
}

func repoRelative(t *testing.T, rel string) string {
	t.Helper()

	// This test runs in its own package directory; walk up to the repo root.
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, rel)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repo root from %s", dir)
		}
		dir = parent
	}
}
