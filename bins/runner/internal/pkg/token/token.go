package token

import (
	_ "embed"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
)

const (
	Filename = "/opt/nuon/runner/token"
)

//go:embed templates/token.env
var tokenTemplate string

// WriteFile writes the runner token to the token file using the token template.
func WriteFile(tok string) error {
	dir := filepath.Dir(Filename)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.Wrap(err, "unable to create token directory")
	}

	tmpl := template.Must(template.New("").Parse(tokenTemplate))
	f, err := os.Create(Filename)
	if err != nil {
		return errors.Wrap(err, "unable to create token file")
	}
	defer f.Close()

	if err := os.Chmod(Filename, 0600); err != nil {
		return errors.Wrap(err, "unable to set token file permissions")
	}

	data := struct {
		RunnerAPIToken string
	}{
		RunnerAPIToken: tok,
	}
	if err := tmpl.Execute(f, data); err != nil {
		return errors.Wrap(err, "unable to execute template for token file")
	}

	return nil
}
