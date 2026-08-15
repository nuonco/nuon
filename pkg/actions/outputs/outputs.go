// Package outputs implements the action-output file grammar shared by the
// runner's action handler and the actions-supervisor. Both ends must agree on
// how a step's outputs file is written and parsed, so the grammar lives here
// rather than in either binary.
package outputs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
)

const (
	kvDelimiter    string = "="
	jsonObjStart   string = "{"
	jsonArrayStart string = "["

	// FilenameFormat is the per-step outputs filename, indexed by step index.
	FilenameFormat string = "%d.nuon-outputs.json"

	// MaxFileSize bounds how much of an action's outputs file the runner will
	// read, so a step can't exhaust runner memory by emitting an enormous file.
	MaxFileSize int64 = 1048576
)

// Filename returns the outputs filename for a given step index.
func Filename(idx int64) string {
	return fmt.Sprintf(FilenameFormat, idx)
}

// ParseLine parses a single outputs line. A line beginning with "{" is a JSON
// object; a top-level JSON array is unsupported; otherwise the line is a
// "key=value" pair.
func ParseLine(str string) (map[string]interface{}, error) {
	if strings.HasPrefix(str, jsonObjStart) {
		var out map[string]interface{}
		if err := json.Unmarshal([]byte(str), &out); err != nil {
			return nil, errors.Wrap(err, "unable to parse as json")
		}
		return out, nil
	}

	if strings.HasPrefix(str, jsonArrayStart) {
		return nil, errors.New("outputs with top level json arrays are not supported yet")
	}

	pieces := strings.SplitN(str, kvDelimiter, 2)
	if len(pieces) == 2 {
		return map[string]interface{}{
			pieces[0]: pieces[1],
		}, nil
	}

	return nil, errors.New("unsupported outputs format, must be a json object or k=v string")
}

// ParseFile reads an outputs file and merges every line into a single map. A
// missing file is treated as empty output. An image-backed action shares this
// workspace, so the path is attacker-controlled: O_NOFOLLOW stops a symlink
// redirecting the read at host state, O_NONBLOCK stops a FIFO hanging the
// runner on open, and the fstat rejects anything but a regular file.
func ParseFile(path string) (map[string]interface{}, error) {
	out := make(map[string]interface{}, 0)

	fh, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, errors.Wrap(err, "unable to open outputs file")
	}
	defer fh.Close()

	info, err := fh.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "unable to stat outputs file")
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("outputs file is not a regular file")
	}
	if info.Size() > MaxFileSize {
		return nil, errors.Errorf("outputs file is %d bytes, which exceeds the %d byte limit", info.Size(), MaxFileSize)
	}

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		lineOutputs, err := ParseLine(line)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing outputs")
		}
		out = generics.MergeMap(out, lineOutputs)
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "unable to scan outputs file")
	}

	return out, nil
}
