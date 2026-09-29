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

	FilenameFormat string = "%d.nuon-outputs.json"

	MaxFileSize int64 = 1048576
)

func Filename(idx int64) string {
	return fmt.Sprintf(FilenameFormat, idx)
}

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
