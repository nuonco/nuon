package get

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	getter "github.com/hashicorp/go-getter"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/publicnet"
)

const (
	maxRestrictedFieldBytes = 10 << 20
	maxRestrictedRedirects  = 5
)

var forcedGetterRe = regexp.MustCompile(`^([A-Za-z0-9]+)::(.+)$`)

// processRestrictedField resolves a get-enabled field for configs from an
// untrusted source. Local paths must stay inside RootDir, remote fetches must
// reach the public internet, and every other go-getter source is refused.
func (g *get) processRestrictedField(ctx context.Context, inputVal, pwd string) (string, error) {
	if strings.HasPrefix(inputVal, "./") || strings.HasPrefix(inputVal, "../") {
		return g.readRootFile(inputVal, pwd)
	}

	detected, err := getter.Detect(inputVal, pwd, GetDetectors())
	if err != nil {
		return inputVal, nil
	}

	ctx, cancel := context.WithTimeout(ctx, g.opts.FieldTimeout)
	defer cancel()

	if m := forcedGetterRe.FindStringSubmatch(detected); m != nil {
		if m[1] != "git" {
			return "", fmt.Errorf("unsupported source %q: %s:: sources are not allowed", inputVal, m[1])
		}

		tmpDir, err := os.MkdirTemp("", ".nuon-get-*")
		if err != nil {
			return "", errors.Wrap(err, "failed to create temp directory")
		}
		defer os.RemoveAll(tmpDir)

		content, err := fetchRestrictedGitFile(ctx, m[2], tmpDir)
		if errors.Is(err, errIsDirectory) {
			return inputVal, nil
		}
		return content, err
	}

	u, err := url.Parse(detected)
	if err != nil {
		return "", fmt.Errorf("invalid source %q", inputVal)
	}

	switch u.Scheme {
	case "http", "https":
		if err := requireHostname(u); err != nil {
			return "", err
		}
		return fetchRestrictedHTTP(ctx, u.String())
	case "file":
		// The file detector turns scheme-less values such as "files/a.yaml"
		// into absolute file URLs; those are still repo-relative paths.
		if orig, err := url.Parse(inputVal); err == nil && orig.Scheme == "" {
			return g.readRootFile(inputVal, pwd)
		}
	}

	return "", fmt.Errorf("unsupported source %q: only repository paths and http(s) or git https URLs are allowed", inputVal)
}

func (g *get) readRootFile(inputVal, pwd string) (string, error) {
	root, err := filepath.Abs(g.opts.RootDir)
	if err != nil {
		return "", errors.Wrap(err, "unable to resolve config root")
	}
	confine := root
	if g.opts.ConfineDir != "" {
		confine, err = filepath.Abs(g.opts.ConfineDir)
		if err != nil {
			return "", errors.Wrap(err, "unable to resolve confine directory")
		}
	}
	resolved, err := filepath.Abs(filepath.Join(pwd, inputVal))
	if err != nil {
		return "", errors.Wrap(err, "unable to resolve local source file")
	}
	confined, ok := relativeWithin(confine, resolved)
	if !ok {
		return "", fmt.Errorf("source %q is outside the repository", inputVal)
	}

	contents, err := readFileInRoot(confine, confined)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read %q", inputVal)
	}

	if relative, ok := relativeWithin(root, resolved); ok && g.opts.OnLocalFile != nil {
		if err := g.opts.OnLocalFile(filepath.ToSlash(relative), contents); err != nil {
			return "", err
		}
	}

	return string(contents), nil
}

func relativeWithin(dir, target string) (string, bool) {
	relative, err := filepath.Rel(dir, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", false
	}
	return relative, true
}

// inGitDir reports whether name, after following symlinks, is inside a .git
// directory under dir. Clone remotes there can hold VCS credentials.
func inGitDir(dir, name string) (bool, error) {
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return false, err
	}
	realPath, err := filepath.EvalSymlinks(filepath.Join(dir, name))
	if err != nil {
		return false, err
	}
	relative, ok := relativeWithin(realDir, realPath)
	if !ok {
		return false, errors.New("path escapes the repository")
	}
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		if part == ".git" {
			return true, nil
		}
	}
	return false, nil
}

// readFileInRoot reads name through an os.Root so that neither ".." nor a
// symlink can resolve to a file outside dir.
func readFileInRoot(dir, name string) ([]byte, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	f, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if gitDir, err := inGitDir(dir, name); err != nil {
		return nil, err
	} else if gitDir {
		return nil, errors.New("files under .git cannot be read")
	}

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errIsDirectory
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("not a regular file")
	}

	return readLimited(f)
}

// requireHostname refuses URLs that address a host by IP. Legitimate sources
// use hostnames; the dial-time check still catches hostnames that resolve to
// non-public addresses.
func requireHostname(u *url.URL) error {
	if _, err := netip.ParseAddr(u.Hostname()); err == nil {
		return fmt.Errorf("source %q must use a hostname, not an IP address", u.Redacted())
	}
	return nil
}

func readLimited(r io.Reader) ([]byte, error) {
	contents, err := io.ReadAll(io.LimitReader(r, maxRestrictedFieldBytes+1))
	if err != nil {
		return nil, err
	}
	if len(contents) > maxRestrictedFieldBytes {
		return nil, fmt.Errorf("source is larger than %d bytes", maxRestrictedFieldBytes)
	}
	return contents, nil
}

func newRestrictedHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = publicnet.DialContext(dialer)

	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRestrictedRedirects {
				return errors.New("too many redirects")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return fmt.Errorf("redirect to %s is not allowed", req.URL.Scheme)
			}
			if err := requireHostname(req.URL); err != nil {
				return err
			}
			return nil
		},
	}
}

func fetchRestrictedHTTP(ctx context.Context, src string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return "", errors.Wrap(err, "invalid source url")
	}

	resp, err := newRestrictedHTTPClient().Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to fetch "+src)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to fetch %s: %s", src, resp.Status)
	}

	contents, err := readLimited(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read "+src)
	}
	return string(contents), nil
}

// fetchRestrictedGitFile clones an https git repository on a public host and
// reads one file from it. It shells out to git directly rather than through
// go-getter so that submodules, LFS, non-https transports, and redirects are
// all disabled.
func fetchRestrictedGitFile(ctx context.Context, src, tmpDir string) (string, error) {
	repoURL, fileSubdir := getter.SourceDirSubdir(src)
	if fileSubdir == "" {
		return "", errors.New("git source must include a `//path/to/file` reference")
	}

	cleanSubdir := path.Clean(fileSubdir)
	if path.IsAbs(cleanSubdir) || cleanSubdir == ".." || strings.HasPrefix(cleanSubdir, "../") {
		return "", errors.New("git source path must be relative and within the repo")
	}

	u, err := url.Parse(repoURL)
	if err != nil {
		return "", errors.Wrap(err, "invalid git source")
	}
	if u.Scheme != "https" {
		return "", fmt.Errorf("git sources must use https, got %q", u.Scheme)
	}
	if err := requireHostname(u); err != nil {
		return "", err
	}
	if err := publicnet.RequirePublicHost(ctx, u.Hostname()); err != nil {
		return "", errors.Wrap(err, "git source must be on a public host")
	}

	ref := u.Query().Get("ref")
	if ref == "" {
		ref = "HEAD"
	}
	u.RawQuery = ""

	cloneDir := filepath.Join(tmpDir, "repo")
	if err := runRestrictedGit(ctx, "", "init", "--quiet", cloneDir); err != nil {
		return "", err
	}
	if err := runRestrictedGit(ctx, cloneDir, "fetch", "--quiet", "--depth", "1", "--no-tags", "--no-recurse-submodules", "--", u.String(), ref); err != nil {
		return "", err
	}
	if err := runRestrictedGit(ctx, cloneDir, "checkout", "--quiet", "FETCH_HEAD"); err != nil {
		return "", err
	}

	contents, err := readFileInRoot(cloneDir, filepath.FromSlash(cleanSubdir))
	if err != nil {
		if errors.Is(err, errIsDirectory) {
			return "", err
		}
		return "", errors.Wrap(err, "failed to read file from git source")
	}
	return string(contents), nil
}

func runRestrictedGit(ctx context.Context, dir string, args ...string) error {
	gitArgs := []string{
		"-c", "protocol.allow=never",
		"-c", "protocol.https.allow=always",
		"-c", "http.followRedirects=false",
		"-c", "submodule.recurse=false",
		"-c", "core.hooksPath=/dev/null",
	}
	if dir != "" {
		gitArgs = append(gitArgs, "-C", dir)
	}
	gitArgs = append(gitArgs, args...)

	cmd := exec.CommandContext(ctx, "git", gitArgs...)
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_LFS_SKIP_SMUDGE=1",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_ALLOW_PROTOCOL=https",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %s failed: %w: %s", args[0], err, strings.TrimSpace(string(out)))
	}
	return nil
}
