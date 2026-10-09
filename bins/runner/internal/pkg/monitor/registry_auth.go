package monitor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/runner/settings"
)

// DockerConfigDirectory is owned by the monitor so a stale credential can never shadow an anonymous pull.
const DockerConfigDirectory = "/opt/nuon/runner/docker"

type dockerConfig struct {
	Auths map[string]dockerConfigAuth `json:"auths"`
}

type dockerConfigAuth struct {
	Auth string `json:"auth"`
}

func writeRegistryAuthConfig(dir string, s *settings.Settings) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return errors.Wrap(err, "unable to create docker config directory")
	}

	cfg := dockerConfig{Auths: map[string]dockerConfigAuth{}}
	if auth := s.ContainerImageRegistryAuth; auth != nil && auth.Registry != "" {
		cfg.Auths[auth.Registry] = dockerConfigAuth{
			Auth: base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password)),
		}
	}
	contents, err := json.Marshal(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to marshal docker config")
	}

	path := filepath.Join(dir, "config.json")
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, contents) {
		return nil
	}
	return writeFileAtomically(path, contents, 0o600)
}

func registryAuthenticator(s *settings.Settings, ref name.Reference) (authn.Authenticator, bool) {
	auth := s.ContainerImageRegistryAuth
	if auth == nil || auth.Registry != ref.Context().RegistryStr() {
		return nil, false
	}
	return &authn.Basic{Username: auth.Username, Password: auth.Password}, true
}
