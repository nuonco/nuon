package agentclient

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Record is the per-agent file stored under ~/.nuon.agents.
type Record struct {
	Agent      string    `yaml:"agent"`
	CLIVersion string    `yaml:"cli_version"`
	AppID      string    `yaml:"app_id"`
	FirstSeen  time.Time `yaml:"first_seen"`
	LastSeen   time.Time `yaml:"last_seen"`
}

// Dir is ~/.nuon.agents for the given home directory.
func Dir(home string) string {
	return filepath.Join(home, ".nuon.agents")
}

// Touch creates or updates the YAML file for rec.Agent. created is true only
// the first time that file is written. now is stored as UTC.
func Touch(dir string, rec Record, now time.Time) (created bool, err error) {
	name, err := fileName(rec.Agent)
	if err != nil {
		return false, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return false, err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return false, err
	}

	path := filepath.Join(dir, name)
	existing, err := readRecord(path)
	if err != nil {
		return false, err
	}
	if existing == nil {
		created = true
		rec.FirstSeen = now.UTC()
	} else {
		rec.FirstSeen = existing.FirstSeen.UTC()
	}
	rec.LastSeen = now.UTC()

	body, err := yaml.Marshal(rec)
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		return false, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return false, err
	}
	return created, nil
}

func fileName(agent string) (string, error) {
	if agent == "" || agent != filepath.Base(agent) || strings.ContainsAny(agent, `/\`) {
		return "", fmt.Errorf("invalid agent name %q", agent)
	}
	return agent + ".yaml", nil
}

func readRecord(path string) (*Record, error) {
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rec Record
	if err := yaml.Unmarshal(body, &rec); err != nil {
		return nil, nil
	}
	if rec.FirstSeen.IsZero() {
		return nil, nil
	}
	return &rec, nil
}
