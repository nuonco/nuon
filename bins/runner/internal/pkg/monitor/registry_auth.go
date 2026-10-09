package monitor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/runner/settings"
)

// DockerConfigDirectory is owned by the monitor so a stale credential can never shadow an anonymous pull.
const DockerConfigDirectory = "/opt/nuon/runner/docker"

// every control plane replica mints its own token, so keep the written one until it nears expiry
const registryAuthKeepWindow = 20 * time.Minute

var (
	writtenRegistryAuthMu sync.Mutex
	writtenRegistryAuth   = map[string]settings.RegistryAuth{}
)

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

	writtenRegistryAuthMu.Lock()
	defer writtenRegistryAuthMu.Unlock()

	path := filepath.Join(dir, "config.json")
	auth := s.ContainerImageRegistryAuth
	if written, ok := writtenRegistryAuth[dir]; ok && auth != nil && keepWrittenRegistryAuth(written, *auth) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}

	cfg := dockerConfig{Auths: map[string]dockerConfigAuth{}}
	if auth != nil && auth.Registry != "" {
		cfg.Auths[auth.Registry] = dockerConfigAuth{
			Auth: base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password)),
		}
	}
	contents, err := json.Marshal(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to marshal docker config")
	}

	if existing, err := os.ReadFile(path); err != nil || !bytes.Equal(existing, contents) {
		if err := writeFileAtomically(path, contents, 0o600); err != nil {
			return err
		}
	}

	if auth != nil && auth.Registry != "" {
		writtenRegistryAuth[dir] = *auth
	} else {
		delete(writtenRegistryAuth, dir)
	}
	return nil
}

func keepWrittenRegistryAuth(written, offered settings.RegistryAuth) bool {
	if written.Registry != offered.Registry || written.Username != offered.Username {
		return false
	}
	if written.ExpiresAt.IsZero() || offered.ExpiresAt.IsZero() {
		return written.Password == offered.Password
	}
	if time.Until(written.ExpiresAt) > registryAuthKeepWindow {
		return true
	}
	return !offered.ExpiresAt.After(written.ExpiresAt)
}

func registryAuthenticator(s *settings.Settings, ref name.Reference) (authn.Authenticator, bool) {
	auth := s.ContainerImageRegistryAuth
	if auth == nil || auth.Registry != ref.Context().RegistryStr() {
		return nil, false
	}
	return &authn.Basic{Username: auth.Username, Password: auth.Password}, true
}
