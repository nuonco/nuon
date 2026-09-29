package get

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"reflect"
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
	subdir = g.nextSourceFileSubdir(v, subdir)

	val := reflect.ValueOf(v)

	if val.Kind() != reflect.Ptr {
		ptr := reflect.New(val.Type())
		ptr.Elem().Set(val)
		val = ptr
	}

	val = val.Elem()

	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		fieldSubdir := nextSubdir(subdir, fieldType.Name)

		switch field.Kind() {
		case reflect.Ptr:
			if field.IsNil() {
				continue
			}

			if err := g.walkFields(ctx, field.Interface(), fieldSubdir); err != nil {
				return err
			}
		case reflect.Struct:
			if err := g.walkFields(ctx, field.Interface(), fieldSubdir); err != nil {
				return err
			}
		case reflect.Slice:
			for i := 0; i < field.Len(); i++ {
				elem := field.Index(i)
				switch elem.Kind() {
				case reflect.Struct:
					ptr := reflect.New(elem.Type())
					ptr.Elem().Set(elem)
					if err := g.walkFields(ctx, ptr.Interface(), fieldSubdir); err != nil {
						return err
					}
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

		getEnabled, err := features.HasGetFeature(fieldType)
		if err != nil {
			return errors.Wrap(err, "unable to parse field "+fieldType.Name)
		}

		if !getEnabled {
			continue
		}

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
					field.Set(reflect.New(field.Type().Elem()))
				}
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

func (g *get) nextSourceFileSubdir(v interface{}, current string) string {
	getter, ok := v.(sourceFileGetter)
	if !ok {
		return current
	}

	sourceFile := getter.GetSourceFile()
	if sourceFile == "" {
		return current
	}

	sourceDir := filepath.Dir(sourceFile)
	if sourceDir == "." {
		return current
	}

	sourceDirPath := filepath.Join(g.opts.RootDir, sourceDir)
	if stat, err := os.Stat(sourceDirPath); err == nil && stat.IsDir() {
		return sourceDir
	}

	return current
}

func (g *get) processField(ctx context.Context, inputVal string, subdir string) (string, error) {
	prefixes := []string{
		"http",
		"./",
		"../",
		"git",
		"file",
	}
	isGettable := false
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
		if stat, err := os.Stat(subdirPath); err == nil && stat.IsDir() {
			pwd = subdirPath
		}
	}

	detected, err := getter.Detect(inputVal, pwd, GetDetectors())
	if err != nil {
		return inputVal, nil
	}
	if err := g.recordLocalFile(inputVal, pwd); err != nil {
		return "", err
	}

	tmpDir, err := os.MkdirTemp(pwd, ".nuon-get-*")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp directory")
	}
	defer os.RemoveAll(tmpDir)

	ctx, cancel := context.WithTimeout(ctx, g.opts.FieldTimeout)
	defer cancel()

	if strings.HasPrefix(detected, "git::") {
		content, err := g.fetchGitFile(ctx, detected, tmpDir, pwd)
		if errors.Is(err, errIsDirectory) {
			return inputVal, nil
		}
		return content, err
	}

	tmpFP := filepath.Join(tmpDir, "field")

	client := &getter.Client{
		Ctx:  ctx,
		Src:  inputVal,
		Dir:  true,
		Dst:  tmpFP,
		Pwd:  pwd,
		Mode: getter.ClientModeFile,
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

// why: fetchGitFile clones a git repo and reads a single file from it. detected
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

	cloneDir := filepath.Join(tmpDir, "repo")

	client := &getter.Client{
		Ctx:       ctx,
		Src:       repoURL,
		Dst:       cloneDir,
		Pwd:       pwd,
		Mode:      getter.ClientModeDir,
		Detectors: GetDetectors(),
	}

	if err := client.Get(); err != nil {
		return "", errors.Wrap(err, "failed to fetch git source")
	}

	target := filepath.Join(cloneDir, filepath.FromSlash(cleanSubdir))

	if stat, err := os.Stat(target); err == nil && stat.IsDir() {
		return "", errIsDirectory
	}

	content, err := os.ReadFile(target)
	if err != nil {
		return "", errors.Wrap(err, "failed to read file from git source")
	}

	return string(content), nil
}
