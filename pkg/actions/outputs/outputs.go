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

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
)

const (
	kvDelimiter    string = "="
	jsonObjStart   string = "{"
	jsonArrayStart string = "["

	// FilenameFormat is the per-step outputs filename, indexed by step index.
	FilenameFormat string = "%d.nuon-outputs.json"
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

// ParseFile reads an outputs file and merges every line into a single map.
// A missing file is treated as empty output.
func ParseFile(path string) (map[string]interface{}, error) {
	out := make(map[string]interface{}, 0)

	fh, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, errors.Wrap(err, "unable to open outputs file")
	}
	defer fh.Close()

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
