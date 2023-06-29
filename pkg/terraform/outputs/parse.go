package outputs

import (
	"fmt"
	"sort"

	"github.com/tidwall/gjson"
)

// Output represents a single out value.
// A slice of Output structs with Paths can be used
// to represent deeply nested structures of objects and arrays
// as a flat list by way of modeling the tree structure
// as slices for path strings
type Output struct {
	// used to model nesting allowing us to flatten deep trees of values
	// for example []string{"server", "port"}
	// could become a jq path `server.port` or
	// an env var `SERVER_PORT`
	Path []string
	// string enum of JSON/HCL data types
	// "null", "boolean", "number", "string", "array", "object"
	Type string

	// For this type of dynamically typed data in golang,
	// there's naming conventions we could attempt to follow
	// from either gjson or go standard library reflect package perhaps.
	// For now I just took literally the go keyword and capitalized the
	// first letter

	// holds the actual value for strings
	String string
	// holds the actual value for booleans
	Bool bool
	// holds the actual value for booleans
	Float64 float64
}

type valueType int

const (
	valueTypeUknown  valueType = iota
	valueTypeNull    valueType = iota
	valueTypeBoolean valueType = iota
	valueTypeNumber  valueType = iota
	valueTypeString  valueType = iota
	valueTypeArray   valueType = iota
	valueTypeObject  valueType = iota
)

func getValueType(gValue gjson.Result) valueType {
	if gValue.IsObject() {
		return valueTypeObject
	}
	if gValue.IsArray() {
		return valueTypeArray
	}
	switch gValue.Type {
	case gjson.Null:
		return valueTypeNull
	case gjson.False:
		fallthrough
	case gjson.True:
		return valueTypeBoolean
	case gjson.Number:
		return valueTypeNumber
	case gjson.String:
		return valueTypeString
	}
	return valueTypeUknown
}

func scalarValueToOutput(gValue gjson.Result) Output {
	switch gValue.Type {
	case gjson.Null:
		return Output{Type: "null"}
	case gjson.False:
		fallthrough
	case gjson.True:
		return Output{Type: "boolean", Bool: gValue.Bool()}
	case gjson.Number:
		return Output{Type: "number", Float64: gValue.Num}
	case gjson.String:
		return Output{Type: "string", String: gValue.Str}
	}
	return Output{Type: "unknown"}
}

// We sort keys for both developer convenience in writing
// expected values for unit tests as well as just
// deterministic behavior is easier for end users.
// Of course in go map key iteration order is forcably
// made non-deterministic to prevent developers from
// assuming and relying on deterministc order
// but our []Output slice is not a go library API
// so we sort for reduction of global chaos
func sortKeys(input map[string]gjson.Result) []string {
	keys := make([]string, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// This is the primary recursive parsing function
func parseValue(path []string, gValue gjson.Result) ([]Output, error) {
	outputs := []Output{}
	switch getValueType(gValue) {
	case valueTypeBoolean:
		fallthrough
	case valueTypeNumber:
		fallthrough
	case valueTypeString:
		output := scalarValueToOutput(gValue)
		// Make a copy of the path slice function argument
		// so it's immutable from change
		// in other parts of the program
		output.Path = path[:]
		outputs = append(outputs, output)
	case valueTypeObject:
		outputs = append(outputs, Output{Path: path, Type: "object"})
		gMap := gValue.Map()
		for _, key := range sortKeys(gMap) {
			// This recursive call achieves the nesting and flattening
			nestedOutputs, err := parseValue(append(path, key), gMap[key])
			if err != nil {
				return nil, fmt.Errorf("parse error at path: %+v, key %s: %w", path, key, err)
			}
			outputs = append(outputs, nestedOutputs...)
		}
	case valueTypeArray:
		outputs = append(outputs, Output{Path: path, Type: "array"})
		for index, nestedValue := range gValue.Array() {
			position := fmt.Sprintf("[%d]", index)
			// This recursive call achieves the nesting and flattening
			nestedOutputs, err := parseValue(append(path, position), nestedValue)
			if err != nil {
				return nil, fmt.Errorf("parse error at path: %+v, key %s: %w", path, position, err)
			}
			outputs = append(outputs, nestedOutputs...)
		}

	}
	return outputs, nil
}

// ParseJSONL parses terraform output [jsonl](https://jsonlines.org/) format
// as well as any JSON including pretty-printed
// and represents a tree of terraform output values
// as a flat slice of Output structs with slice path keys
// denoting the location in the tree structure.
func ParseJSONL(jsonl []byte) ([]Output, error) {
	outputs := []Output{}
	oops := []error{}
	gjson.ForEachLine(string(jsonl), func(line gjson.Result) bool {
		// filter to find just the outputs and extract just the metadata wrapper object
		wrapperMaps := gjson.GetMany(line.String(), `..#(type="outputs").outputs`)
		for outputNumber, wrapperMap := range wrapperMaps {
			// outputNumber is not necessarily line number in the .jsonl file
			// because we skip past irrelevant values such as UI messages.
			// Parsing error messages are not really going to be able to pinpoint
			// errors for you, but we'll at least try to help.

			// wrapperMap is the outer map of key name to object metadata
			// all terraform outputs are named so everything gets represented
			// in this top level object with the output name as the key
			// and value is an object of metadata about the value plus the actual value
			for _, key := range sortKeys(wrapperMap.Map()) {
				valueObj := wrapperMap.Map()[key]
				lineOutputs, err := parseValue([]string{key}, valueObj.Get("value"))
				if err != nil {
					oops = append(oops, fmt.Errorf("parsing error at output %d: %w", outputNumber, err))
					return false
				}
				outputs = append(outputs, lineOutputs...)
			}
		}
		return true
	})
	if len(oops) > 0 {
		// TODO build compound error
		return outputs, oops[0]
	}
	return outputs, nil
}
