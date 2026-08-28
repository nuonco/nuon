package statestore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Disk struct {
	root string
	mu   sync.Mutex
}

func NewDisk(root string) (*Disk, error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create state root: %w", err)
	}
	return &Disk{root: root}, nil
}

func (d *Disk) WriteFile(relPath string, data []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	path, err := d.safePath(relPath)
	if err != nil {
		return err
	}
	return d.atomicWrite(path, data)
}

func (d *Disk) ReadFile(relPath string) ([]byte, bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	path, err := d.safePath(relPath)
	if err != nil {
		return nil, false, err
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	return b, err == nil, err
}

func (d *Disk) safePath(relPath string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(relPath))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe relative state path %q", relPath)
	}
	return filepath.Join(d.root, clean), nil
}

func (d *Disk) ReadStatus() (*Status, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	b, err := os.ReadFile(filepath.Join(d.root, "status.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var status Status
	if err := json.Unmarshal(b, &status); err != nil {
		return nil, err
	}
	if status.RunID != "" {
		latest, eventErr := d.readLatestEvent(status.RunID)
		if eventErr != nil {
			return nil, eventErr
		}
		if latest != nil {
			return &latest.Status, nil
		}
	}
	return &status, nil
}

func (d *Disk) WriteStatus(status *Status) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if status.RunID != "" {
		sequence, err := d.nextEventSequence(status.RunID)
		if err != nil {
			return err
		}
		event := StatusEvent{SchemaVersion: 1, Sequence: sequence, CreatedAt: time.Now().UTC(), Status: *status}
		if err := d.writeJSON(filepath.Join(d.root, filepath.FromSlash(InstallRunEventKey(status.RunID, sequence))), event); err != nil {
			return err
		}
		if err := d.writeJSON(filepath.Join(d.root, filepath.FromSlash(InstallRunStatusKey(status.RunID))), status); err != nil {
			return err
		}
	}
	return d.writeJSON(filepath.Join(d.root, "status.json"), status)
}

func (d *Disk) nextEventSequence(runID string) (uint64, error) {
	latest, err := d.readLatestEvent(runID)
	if err != nil || latest == nil {
		return 1, err
	}
	return latest.Sequence + 1, nil
}

func (d *Disk) readLatestEvent(runID string) (*StatusEvent, error) {
	dir := filepath.Join(d.root, filepath.FromSlash(InstallRunEventsPrefix(runID)))
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			names = append(names, entry.Name())
		}
	}
	if len(names) == 0 {
		return nil, nil
	}
	sort.Strings(names)
	b, err := os.ReadFile(filepath.Join(dir, names[len(names)-1]))
	if err != nil {
		return nil, err
	}
	var event StatusEvent
	if err := json.Unmarshal(b, &event); err != nil {
		return nil, err
	}
	if event.Sequence == 0 {
		value := strings.TrimSuffix(names[len(names)-1], ".json")
		event.Sequence, _ = strconv.ParseUint(value, 10, 64)
	}
	return &event, nil
}

func (d *Disk) WriteResult(id string, value any) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.writeJSON(d.stepPath(id, "result.json"), value)
}

func (d *Disk) WriteOutputs(id string, value any) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.writeJSON(d.stepPath(id, "outputs.json"), value)
}

func (d *Disk) AppendExecution(id string, value any) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	path := d.stepPath(id, "executions.json")
	var values []json.RawMessage
	if b, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(b, &values); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	values = append(values, b)
	return d.writeJSON(path, values)
}

func (d *Disk) GetTFState(id string) ([]byte, bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	b, err := os.ReadFile(d.tfPath(id, ""))
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	return b, err == nil, err
}

func (d *Disk) PutTFState(id string, state []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.atomicWrite(d.tfPath(id, ""), state)
}

func (d *Disk) PutTFStateShow(id string, doc []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.atomicWrite(d.tfPath(id, ".show"), doc)
}

func (d *Disk) ReadResult(id string) (json.RawMessage, bool, error) {
	return d.readStepFile(id, "result.json")
}

func (d *Disk) ReadOutputs(id string) (json.RawMessage, bool, error) {
	return d.readStepFile(id, "outputs.json")
}

func (d *Disk) ReadExecutions(id string) (json.RawMessage, bool, error) {
	return d.readStepFile(id, "executions.json")
}

func (d *Disk) readStepFile(id, name string) (json.RawMessage, bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	b, err := os.ReadFile(d.stepPath(id, name))
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	return b, err == nil, err
}

func (d *Disk) WriteReport(value any) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.writeJSON(filepath.Join(d.root, "report.json"), value)
}

func (d *Disk) ReadReport() (json.RawMessage, bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	b, err := os.ReadFile(filepath.Join(d.root, "report.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	return b, err == nil, err
}

func (d *Disk) WriteHealth(value any) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.writeJSON(filepath.Join(d.root, "health", "latest.json"), value)
}

func (d *Disk) ReadHealth() (json.RawMessage, bool, error) {
	return d.readFile(filepath.Join(d.root, "health", "latest.json"))
}

func (d *Disk) AppendHealthTransitions(values []any) error {
	if len(values) == 0 {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	path := filepath.Join(d.root, "health", "transitions.json")
	var existing []json.RawMessage
	if b, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(b, &existing); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, value := range values {
		b, err := json.Marshal(value)
		if err != nil {
			return err
		}
		existing = append(existing, b)
	}
	return d.writeJSON(path, existing)
}

func (d *Disk) ReadHealthTransitions() (json.RawMessage, bool, error) {
	return d.readFile(filepath.Join(d.root, "health", "transitions.json"))
}

func (d *Disk) WriteHealthContext(value any) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.writeJSON(filepath.Join(d.root, "health", "context.json"), value)
}

func (d *Disk) ReadHealthContext() (json.RawMessage, bool, error) {
	return d.readFile(filepath.Join(d.root, "health", "context.json"))
}

func (d *Disk) readFile(path string) (json.RawMessage, bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	return b, err == nil, err
}

func (d *Disk) LockTF(id string, lock []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	path := d.tfPath(id, ".lock")
	if existing, err := os.ReadFile(path); err == nil {
		return &LockConflictError{Existing: existing}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return d.atomicWrite(path, lock)
}

func (d *Disk) UnlockTF(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	err := os.Remove(d.tfPath(id, ".lock"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (d *Disk) stepPath(id, name string) string {
	return filepath.Join(d.root, "steps", safe(id), name)
}

func (d *Disk) tfPath(id, suffix string) string {
	return filepath.Join(d.root, "tfstate", safe(id)+".json"+suffix)
}

func safe(value string) string {
	return strings.ReplaceAll(filepath.Base(value), "..", "_")
}

func (d *Disk) writeJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return d.atomicWrite(path, append(b, '\n'))
}

func (d *Disk) atomicWrite(path string, value []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(value); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
