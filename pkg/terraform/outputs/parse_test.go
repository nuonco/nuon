package outputs

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func load(relativePath string) []byte {
	bytes, err := os.ReadFile(relativePath)
	if err != nil {
		panic(fmt.Sprintf("parse_test.go error loading test JSON file %s. %s", relativePath, err))
	}
	return bytes
}

func TestParseJSONL(t *testing.T) {
	tests := map[string]struct {
		jsonl    []byte
		expected []Output
	}{
		"JSONL with flat scalar outputs": {
			jsonl: load("testdata/outputs_flat.jsonl"),
			expected: []Output{
				{Path: []string{"apiKey1"}, Type: "string", String: "aaa"},
				{Path: []string{"apiKey2"}, Type: "string", String: "bbb"},
				{Path: []string{"hostname"}, Type: "string", String: "example.com"},
				{Path: []string{"tls"}, Type: "boolean", Bool: true},
				{Path: []string{"CACHE_SIZE"}, Type: "number", Float64: 1024},
				{Path: []string{"nada"}, Type: "null"},
			},
		},
		"JSON with array": {
			jsonl: load("testdata/outputs_array.json"),
			expected: []Output{
				{Path: []string{"sample_array_mixed"}, Type: "array"},
				{Path: []string{"sample_array_mixed", "[0]"}, Type: "boolean", Bool: true},
				{Path: []string{"sample_array_mixed", "[1]"}, Type: "number", Float64: 42.42},
				{Path: []string{"sample_array_mixed", "[2]"}, Type: "string", String: "cheese"},
				{Path: []string{"sample_array_mixed", "[3]"}, Type: "null"},
				{Path: []string{"sample_array_mixed", "[4]"}, Type: "boolean", Bool: false},
			},
		},
		"JSON with object": {
			jsonl: load("testdata/outputs_object.json"),
			expected: []Output{
				{Path: []string{"server"}, Type: "object"},
				{Path: []string{"server", "barkey"}, Type: "boolean", Bool: true},
				{Path: []string{"server", "fookey"}, Type: "string", String: "FOOVALUE"},
			},
		},
		"JSON with nested array": {
			jsonl: load("testdata/outputs_nested_array.json"),
			expected: []Output{
				{Path: []string{"sample_nested_array"}, Type: "array"},
				{Path: []string{"sample_nested_array", "[0]"}, Type: "boolean", Bool: true},
				{Path: []string{"sample_nested_array", "[1]"}, Type: "number", Float64: 42.42},
				{Path: []string{"sample_nested_array", "[2]"}, Type: "array"},
				{Path: []string{"sample_nested_array", "[2]", "[0]"}, Type: "string", String: "nested"},
				{Path: []string{"sample_nested_array", "[2]", "[1]"}, Type: "number", Float64: 17},
			},
		},
		"JSON with nested objects and arrays": {
			jsonl: load("testdata/outputs_nested_object.json"),
			expected: []Output{
				{Path: []string{"client"}, Type: "object"},
				{Path: []string{"client", "theme"}, Type: "string", String: "dark"},
				{Path: []string{"server"}, Type: "object"},
				{Path: []string{"server", "allow_hosts"}, Type: "array"},
				{Path: []string{"server", "allow_hosts", "[0]"}, Type: "string", String: "example.com"},
				{Path: []string{"server", "allow_hosts", "[1]"}, Type: "string", String: "example.org"},
				{Path: []string{"server", "bind"}, Type: "string", String: "127.0.0.1"},
				{Path: []string{"server", "tls"}, Type: "boolean", Bool: true},
			},
		},
		"invalid JSONL": {
			jsonl:    []byte("this is not json\nnor is this"),
			expected: []Output{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := ParseJSONL(test.jsonl)
			assert.EqualValues(t, test.expected, actual)
		})
	}
}
