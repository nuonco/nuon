package parse

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse/dir"
	"github.com/powertoolsdev/mono/pkg/config/parse/get"
)

const (
	defaultFieldGetTimeout time.Duration = time.Second * 5
)

func parseDirName(dirname string) (string, error) {
	if _, err := os.Stat(dirname); err != nil {
		if os.IsNotExist(err) {
			return "", config.ErrConfig{
				Description: dirname + " does not exist",
				Err:         err,
			}
		}

		return "", err
	}

	absPath, err := filepath.Abs(dirname)
	if err != nil {
		return "", errors.Wrap(err, "unable to get absolute path")
	}

	return absPath, nil
}

func ParseDir(ctx context.Context, parseCfg ParseConfig) (*config.AppConfig, error) {
	fp, err := parseDirName(parseCfg.Dirname)
	if err != nil {
		return nil, err
	}

	fs := afero.NewOsFs()
	cfgFS := afero.NewBasePathFs(fs, fp)

	// parse the directory
	var obj ConfigDir
	if err := dir.Parse(ctx, cfgFS, &obj, &dir.ParseOptions{
		Root:     fp,
		Ext:      ".toml",
		ParserFn: func(rc io.ReadCloser, s string, a any) error { return parseTomlFile(rc, s, a, parseCfg.FileProcessor) },
	}); err != nil {
		return nil, errors.Wrap(err, "unable to parse directory")
	}

	// NOTE(jm): this will go away once we deprecate the legacy config, and we can just have a pipeline of
	// `config.AppConfig` parsers.
	appCfg, err := obj.toAppConfig()
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert to app config")
	}

	// parse all get functions
	if err := get.Parse(ctx, appCfg, &get.Options{
		FieldTimeout: defaultFieldGetTimeout,
		RootDir:      fp,
	}); err != nil {
		return nil, ParseErr{
			Description: "unable to get fields",
			Err:         err,
		}
	}

	err = appCfg.Parse()
	if err != nil {
		return nil, ParseErr{
			Description: "error parsing config",
			Err:         err,
		}
	}

	return appCfg, nil
}
