package get

import (
	"context"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	getter "github.com/hashicorp/go-getter"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config/features"
)

func (g *get) GetAll(ctx context.Context) error {
	return g.walkFields(ctx, g.dst, "")
}

var errIsDirectory = errors.New("git source path is a directory")

type sourceFileGetter interface {
	GetSourceFile() string
}

func nextSubdir(current, fieldName string) string {
	switch fieldName {
	case "Components":
		return "components"
	case "Actions":
		return "actions"
	case "Permissions":
		return "permissions"
	case "BreakGlass":
		return "permissions"
	case "Installs":
		return "installs"
	default:
		return current
	}
}

func (g *get) walkFields(ctx context.Context, v interface{}, subdir string) error {
	var err error
	subdir, err = g.nextSourceFileSubdir(v, subdir)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(v)

	// If it's not a pointer, we need to get a pointer to make it settable
	if val.Kind() != reflect.Ptr {
		// Create a new pointer to the value
		ptr := reflect.New(val.Type())
		ptr.Elem().Set(val)
		val = ptr
	}

	// Now we can safely get the underlying value
	val = val.Elem()

	// We only process struct types
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		fieldSubdir := nextSubdir(subdir, fieldType.Name)

		// if the record is nested, recurse
		switch field.Kind() {
		case reflect.Ptr:
			// If it's a nil pointer, skip it
			if field.IsNil() {
				continue
			}

			// Recurse with the dereferenced pointer
			if err := g.walkFields(ctx, field.Interface(), fieldSubdir); err != nil {
				return err
			}
		case reflect.Struct:
			if err := g.walkFields(ctx, field.Interface(), fieldSubdir); err != nil {
				return err
			}
		case reflect.Slice:
			// Handle slices of structs
			for i := 0; i < field.Len(); i++ {
				elem := field.Index(i)
				switch elem.Kind() {
				case reflect.Struct:
					// Create a pointer to make it settable
					ptr := reflect.New(elem.Type())
					ptr.Elem().Set(elem)
					if err := g.walkFields(ctx, ptr.Interface(), fieldSubdir); err != nil {
						return err
					}
					// Update the slice element with potentially modified value
					if elem.CanSet() {
						elem.Set(ptr.Elem())
					}
				case reflect.Ptr:
					if elem.IsNil() {
						continue
					}

					if elem.Elem().Kind() == reflect.Struct {
						if err := g.walkFields(ctx, elem.Interface(), fieldSubdir); err != nil {
							return err
						}
					}
				}
			}
		}

		// check if get-enabled exists
		getEnabled, err := features.HasGetFeature(fieldType)
		if err != nil {
			return errors.Wrap(err, "unable to parse field "+fieldType.Name)
		}

		if !getEnabled {
			continue
		}

		// now, check if it is a string field
		if field.Kind() == reflect.String {
			strValue := field.String()

			val, err := g.processField(ctx, strValue, fieldSubdir)
			if err != nil {
				return errors.Wrap(err, "unable to fetch field value")
			}

			if !field.CanSet() {
				return errors.New("field is not settable: " + fieldType.Name)
			}

			if field.Kind() == reflect.Ptr {
				if field.IsNil() {
					// Create a new pointer if it's nil
					field.Set(reflect.New(field.Type().Elem()))
				}
				// Set the value on the element that the pointer points to
				field.Elem().SetString(val)
			} else {
				field.SetString(val)
			}

		} else {
			return errors.New("get feature enabled on a non-string field " + fieldType.Name)
		}
	}

	return nil
}

func (g *get) nextSourceFileSubdir(v interface{}, current string) (string, error) {
	getter, ok := v.(sourceFileGetter)
	if !ok {
		return current, nil
	}

	sourceFile := getter.GetSourceFile()
	if sourceFile == "" {
		return current, nil
	}

	sourceDir := filepath.Dir(sourceFile)
	if sourceDir == "." {
		return current, nil
	}

	sourceDirPath := filepath.Join(g.opts.RootDir, sourceDir)
	isDir, err := g.isDir(sourceDirPath)
	if err != nil {
		return "", err
	}
	if isDir {
		return sourceDir, nil
	}

	return current, nil
}

func (g *get) isDir(name string) (bool, error) {
	if g.opts.Source == nil {
		info, err := os.Stat(name)
		return err == nil && info.IsDir(), nil
	}
	resolved, err := g.opts.Source.Resolve(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	info, err := g.opts.Source.Lstat(resolved)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (g *get) processField(ctx context.Context, inputVal string, subdir string) (string, error) {
	prefixes := []string{
		"http",
		"./",
		"../",
		"git",
		"file",
	}
	isGettable := g.opts.Source != nil && filepath.IsAbs(inputVal)
	for _, prefix := range prefixes {
		if strings.HasPrefix(inputVal, prefix) {
			isGettable = true
		}
	}

	if strings.HasPrefix(inputVal, "./nuon") {
		return inputVal, nil
	}

	if !isGettable {
		return inputVal, nil
	}

	pwd := g.opts.RootDir
	if subdir != "" {
		subdirPath := filepath.Join(g.opts.RootDir, subdir)
		isDir, err := g.isDir(subdirPath)
		if err != nil {
			return "", err
		}
		if isDir {
			pwd = subdirPath
		}
	}

	if g.opts.Source != nil {
		localPath, local, err := confinedLocalPath(inputVal)
		if err != nil {
			return "", err
		}
		if local {
			if !filepath.IsAbs(localPath) {
				localPath = filepath.Join(pwd, localPath)
			}
			contents, err := g.opts.Source.ReadFile(localPath)
			if err != nil {
				return "", errors.Wrap(err, "unable to read local source file")
			}
			value := string(contents)
			if g.opts.OnLocalFile != nil {
				relative, err := filepath.Rel(g.opts.RootDir, localPath)
				if err == nil && filepath.IsLocal(relative) {
					if err := g.opts.OnLocalFile(filepath.ToSlash(relative), contents); err != nil {
						return "", err
					}
				}
			}
			return value, nil
		}
	}

	detected, err := getter.Detect(inputVal, pwd, GetDetectors())
	if err != nil {
		return inputVal, nil
	}
	if g.opts.Source != nil {
		for remote := detected; ; {
			raw := remote
			if parts := forcedSourcePattern.FindStringSubmatch(remote); parts != nil {
				if parts[1] != "http" && parts[1] != "https" {
					return "", errors.New("remote field imports must use public HTTP(S), not Git or other getter protocols")
				}
				raw = parts[2]
			}
			u, err := url.Parse(raw)
			if err != nil {
				return "", err
			}
			if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.Opaque != "" || u.User != nil {
				return "", errors.New("remote field imports must use HTTP(S) URLs without credentials")
			}
			checksum := u.Query().Get("checksum")
			if !strings.HasPrefix(checksum, "file:") {
				break
			}
			remote = strings.TrimPrefix(checksum, "file:")
			if _, local, err := confinedLocalPath(remote); err != nil || local {
				return "", errors.New("remote source checksum must not reference a local file")
			}
			remote, err = getter.Detect(remote, pwd, GetDetectors())
			if err != nil {
				return "", err
			}
		}
	}
	if err := g.recordLocalFile(inputVal, pwd); err != nil {
		return "", err
	}

	tmpParent := pwd
	if g.opts.Source != nil {
		// A borrowed source's display path is not authority for host writes.
		tmpParent = ""
	}
	tmpDir, err := os.MkdirTemp(tmpParent, ".nuon-get-*")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp directory")
	}
	defer os.RemoveAll(tmpDir)

	ctx, cancel := context.WithTimeout(ctx, g.opts.FieldTimeout)
	defer cancel()

	// go-getter's ClientModeFile path for git is broken for `//path/to/file`
	// references: GitGetter.GetFile() always strips the last path segment of
	// the URL as the "filename" and treats the rest as the repo URL, which
	// produces clone errors against non-existent repos. For git sources we
	// instead clone the containing directory with ClientModeDir and read the
	// requested file ourselves.
	if strings.HasPrefix(detected, "git::") {
		content, err := g.fetchGitFile(ctx, detected, tmpDir, pwd)
		if errors.Is(err, errIsDirectory) {
			return inputVal, nil
		}
		return content, err
	}

	tmpFP := filepath.Join(tmpDir, "field")

	// Configure the client
	client := &getter.Client{
		Ctx:     ctx,
		Src:     inputVal,
		Dir:     true,
		Dst:     tmpFP,
		Pwd:     pwd,
		Mode:    getter.ClientModeFile,
		Getters: g.remoteGetters(),
	}

	if err := client.Get(); err != nil {
		return "", errors.Wrap(err, "failed to fetch file "+inputVal)
	}

	content, err := os.ReadFile(tmpFP)
	if err != nil {
		return "", errors.Wrap(err, "failed to read file")
	}

	return string(content), nil
}

var forcedSourcePattern = regexp.MustCompile(`^([A-Za-z0-9]+)::(.+)$`)

func confinedLocalPath(input string) (string, bool, error) {
	force, raw := "", input
	if parts := forcedSourcePattern.FindStringSubmatch(input); parts != nil {
		force, raw = parts[1], parts[2]
	}
	if force != "file" && !strings.HasPrefix(raw, "file:") {
		// FileDetector probes the host Pwd. Only remote shorthand detection is
		// safe before handing local names to the borrowed source.
		var detectors []getter.Detector
		for _, detector := range GetDetectors() {
			if _, local := detector.(*getter.FileDetector); !local {
				detectors = append(detectors, detector)
			}
		}
		if _, err := getter.Detect(input, "", detectors); err == nil {
			return "", false, nil
		}
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", false, errors.Wrap(err, "invalid source URL")
	}
	if force == "file" && u.Scheme != "" && u.Scheme != "file" {
		return "", false, errors.New("forced file source must use a local path or file URL")
	}
	if u.Scheme != "" && u.Scheme != "file" {
		return "", false, nil
	}
	if force != "" && force != "file" {
		return "", false, errors.New("local sources cannot use a forced remote getter")
	}
	if u.Host != "" && u.Host != "localhost" {
		return "", false, errors.New("local file URL must not specify a remote host")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return "", false, errors.New("local source query and fragment options are not supported")
	}
	name := u.Path
	if u.Opaque != "" {
		name = u.Opaque
	}
	return filepath.FromSlash(name), true, nil
}

func (g *get) remoteGetters() map[string]getter.Getter {
	if g.opts.Source == nil {
		return nil
	}
	httpGetter := &getter.HttpGetter{
		Client:                g.httpClient,
		XTerraformGetDisabled: true,
		DoNotCheckHeadFirst:   true,
	}
	return map[string]getter.Getter{"http": httpGetter, "https": httpGetter}
}

func (g *get) recordLocalFile(inputVal, pwd string) error {
	if g.opts.OnLocalFile == nil || (!strings.HasPrefix(inputVal, "./") && !strings.HasPrefix(inputVal, "../")) {
		return nil
	}

	root, err := filepath.Abs(g.opts.RootDir)
	if err != nil {
		return errors.Wrap(err, "unable to resolve config root")
	}
	resolved, err := filepath.Abs(filepath.Join(pwd, inputVal))
	if err != nil {
		return errors.Wrap(err, "unable to resolve local source file")
	}
	relative, err := filepath.Rel(root, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return nil
	}
	info, err := os.Lstat(resolved)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	contents, err := os.ReadFile(resolved)
	if err != nil {
		return errors.Wrap(err, "unable to read local source file")
	}
	return g.opts.OnLocalFile(filepath.ToSlash(relative), contents)
}

// fetchGitFile clones a git repo and reads a single file from it. detected
// must be a `git::`-prefixed URL produced by getter.Detect with a
// `//path/to/file` subdir reference identifying the file inside the repo.
//
// We deliberately do not forward the `//subdir` portion to go-getter. After
// every git clone go-getter calls fetchSubmodules, which mutates the shared
// Client to set DisableSymlinks=true. Its subsequent copyDir then fails on
// macOS because EvalSymlinks rewrites /var/folders/... -> /private/var/...,
// which copyDir flags as a symlink escape (ErrSymlinkCopy: "copying of
// symlinks has been disabled"). Cloning the whole repo into our own temp dir
// avoids the copyDir path entirely.
func (g *get) fetchGitFile(ctx context.Context, detected, tmpDir, pwd string) (string, error) {
	repoURL, fileSubdir := getter.SourceDirSubdir(detected)
	if fileSubdir == "" {
		return "", errors.New("git source must include a `//path/to/file` reference")
	}

	cleanSubdir := path.Clean(fileSubdir)
	if path.IsAbs(cleanSubdir) || cleanSubdir == ".." || strings.HasPrefix(cleanSubdir, "../") {
		return "", errors.New("git source path must be relative and within the repo")
	}

	// go-getter's git getter only knows how to clone into a directory that
	// does not yet exist; if we hand it tmpDir directly it falls into its
	// "update" path and runs `git fetch origin -- "<empty ref>"` which fails
	// with "empty string is not a valid pathspec". Use a fresh subdirectory.
	cloneDir := filepath.Join(tmpDir, "repo")

	client := &getter.Client{
		Ctx:       ctx,
		Src:       repoURL,
		Dst:       cloneDir,
		Pwd:       pwd,
		Mode:      getter.ClientModeDir,
		Detectors: GetDetectors(),
		Getters:   g.remoteGetters(),
	}

	if err := client.Get(); err != nil {
		return "", errors.Wrap(err, "failed to fetch git source")
	}

	target := filepath.Join(cloneDir, filepath.FromSlash(cleanSubdir))

	// A directory is a module address (e.g. a curated gcp custom stack), not
	// inlinable file content — pass the original value through untouched.
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		return "", errIsDirectory
	}

	content, err := os.ReadFile(target)
	if err != nil {
		return "", errors.Wrap(err, "failed to read file from git source")
	}

	return string(content), nil
}
