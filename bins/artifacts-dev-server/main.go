// Local-dev artifact server. For each registered artifact, builds the
// binary on demand for the requested OS/arch and serves its install.sh
// with BASE_URL rewritten to the externally-visible host so curl-pipe
// commands work without any S3/CDN setup.
//
// Run from the nuon repo root:
//
//	go run ./bins/artifacts-dev-server
//
// Also runs as a first-class nuonctl service ("artifacts-dev-server");
// see mono/bins/artifacts-dev-server/service.yml.
package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type artifact struct {
	name       string // URL prefix segment, e.g. "stack-cli"
	pkg        string // Go package path, e.g. "./bins/stack-cli"
	bin        string // output binary basename, e.g. "stack-cli"
	scriptPath string // relative install.sh path inside repo root
}

var artifacts = []artifact{
	{
		name:       "stack-cli",
		pkg:        "./bins/stack-cli",
		bin:        "stack-cli",
		scriptPath: "bins/stack-cli/install.sh",
	},
}

var allowed = map[string]bool{
	"darwin/amd64": true,
	"darwin/arm64": true,
	"linux/amd64":  true,
	"linux/arm64":  true,
}

func main() {
	addr := flag.String("addr", ":8189", "listen address")
	repoRoot := flag.String("repo-root", "", "nuon repo root (defaults to the nearest ancestor with bins/artifacts-dev-server/main.go)")
	flag.Parse()

	if *repoRoot == "" {
		root, err := findRepoRoot()
		if err != nil {
			log.Fatalf("locate nuon repo root: %v", err)
		}
		*repoRoot = root
	}

	cacheDir, err := os.MkdirTemp("", "artifacts-dev-server-")
	if err != nil {
		log.Fatalf("mkdir cache: %v", err)
	}
	log.Printf("artifacts-dev-server: repo=%s cache=%s", *repoRoot, cacheDir)

	mux := http.NewServeMux()
	for _, a := range artifacts {
		registerArtifact(mux, a, *repoRoot, cacheDir)
		log.Printf("registered artifact %q (pkg=%s)", a.name, a.pkg)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, "artifacts-dev-server\n\nregistered artifacts:\n")
		for _, a := range artifacts {
			fmt.Fprintf(w, "  - %s\n", a.name)
		}
	})

	log.Printf("artifacts-dev-server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}

func registerArtifact(mux *http.ServeMux, a artifact, repoRoot, cacheDir string) {
	scriptFullPath := filepath.Join(repoRoot, a.scriptPath)
	// Match the prefix the request actually arrives with — `tailscale serve`
	// preserves the public path, so we register both `/<name>/` (direct) and
	// `/artifacts/<name>/` (behind serve) to be tolerant.
	prefixes := []string{"/" + a.name, "/artifacts/" + a.name}

	devPathRE := regexp.MustCompile(`/dev/` + regexp.QuoteMeta(a.bin) + `_([a-z0-9]+)_([a-z0-9]+)(\.gz)?$`)

	for _, prefix := range prefixes {
		prefix := prefix

		mux.HandleFunc(prefix+"/install.sh", func(w http.ResponseWriter, r *http.Request) {
			raw, err := os.ReadFile(scriptFullPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			base := externalBase(r, a.name)
			out := rewriteBase(string(raw), base)
			w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
			_, _ = io.WriteString(w, out)
		})

		mux.HandleFunc(prefix+"/latest.txt", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = io.WriteString(w, "dev\n")
		})

		mux.HandleFunc(prefix+"/dev/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			for osarch := range allowed {
				parts := strings.SplitN(osarch, "/", 2)
				goos, goarch := parts[0], parts[1]
				bin := filepath.Join(cacheDir, fmt.Sprintf("%s_%s_%s", a.bin, goos, goarch))
				cmd := exec.Command("go", "build", "-o", bin, a.pkg)
				cmd.Dir = repoRoot
				cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED=0")
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					http.Error(w, "build failed: "+err.Error(), http.StatusInternalServerError)
					return
				}
				sum, err := sha256File(bin)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				fmt.Fprintf(w, "%s  %s_%s_%s\n", sum, a.bin, goos, goarch)
			}
		})

		mux.HandleFunc(prefix+"/dev/", func(w http.ResponseWriter, r *http.Request) {
			m := devPathRE.FindStringSubmatch(r.URL.Path)
			if m == nil {
				http.NotFound(w, r)
				return
			}
			goos, goarch, suffix := m[1], m[2], m[3]
			if !allowed[goos+"/"+goarch] {
				http.Error(w, "unsupported OS/arch", http.StatusBadRequest)
				return
			}

			bin := filepath.Join(cacheDir, fmt.Sprintf("%s_%s_%s", a.bin, goos, goarch))
			log.Printf("build %s %s/%s", a.name, goos, goarch)
			cmd := exec.Command("go", "build", "-o", bin, a.pkg)
			cmd.Dir = repoRoot
			cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED=0")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stderr
			if err := cmd.Run(); err != nil {
				http.Error(w, "build failed: "+err.Error(), http.StatusInternalServerError)
				return
			}

			f, err := os.Open(bin)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer f.Close()

			w.Header().Set("Content-Type", "application/octet-stream")
			if suffix == ".gz" {
				w.Header().Set("Content-Encoding", "identity")
				gzw := gzip.NewWriter(w)
				defer gzw.Close()
				_, _ = io.Copy(gzw, f)
				return
			}
			_, _ = io.Copy(w, f)
		})
	}
}

// externalBase derives the public base URL for an artifact, accounting for
// tailscale serve which terminates TLS and strips the /artifacts mount prefix
// before forwarding.
func externalBase(r *http.Request, name string) string {
	scheme := "http"
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	forwardedHost := r.Header.Get("X-Forwarded-Host")
	if forwardedHost != "" {
		host = forwardedHost
	}
	// Behind tailscale serve, the /artifacts mount prefix is stripped before
	// forwarding — but the public URL needs it back. Either (a) the request
	// path still has the prefix (no stripping), or (b) X-Forwarded-Host is set
	// (we're behind a proxy that stripped). In both cases, advertise /artifacts.
	if strings.HasPrefix(r.URL.Path, "/artifacts/"+name+"/") || forwardedHost != "" {
		return fmt.Sprintf("%s://%s/artifacts/%s", scheme, host, name)
	}
	return fmt.Sprintf("%s://%s/%s", scheme, host, name)
}

// findRepoRoot walks up from the current working directory looking for
// bins/artifacts-dev-server/main.go — the marker that we're in the nuon repo.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "bins", "artifacts-dev-server", "main.go")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find nuon repo root from %s", dir)
		}
		dir = parent
	}
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func rewriteBase(script, base string) string {
	lines := strings.Split(script, "\n")
	for i, line := range lines {
		if strings.Contains(line, "# DEFAULT_BASE_URL") {
			for j := i + 1; j < len(lines); j++ {
				if strings.HasPrefix(strings.TrimSpace(lines[j]), "BASE_URL=") {
					lines[j] = fmt.Sprintf(`BASE_URL="${STACK_CLI_BASE_URL:-%s}"`, base)
					return strings.Join(lines, "\n")
				}
			}
		}
	}
	return script
}
