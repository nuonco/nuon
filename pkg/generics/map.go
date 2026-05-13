package generics

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

func SliceToMapDefault[T comparable, V any](vals []T, deflt V) map[T]V {
	obj := make(map[T]V, 0)

	for _, v := range vals {
		obj[v] = deflt
	}

	return obj
}

func SliceToMap[T comparable, V any](vals []T) map[T]V {
	obj := make(map[T]V, 0)

	for _, v := range vals {
		var val V
		obj[v] = val
	}

	return obj
}

func MapValuesToSlice(inp map[string]string) []string {
	vals := make([]string, 0)
	for _, v := range inp {
		vals = append(vals, v)
	}

	return vals
}

func FindMap[K comparable, T comparable](k K, inps ...map[K]T) T {
	var empty T
	for _, inp := range inps {
		val, ok := inp[k]
		if ok && val != empty {
			return val
		}
	}

	return empty
}

func MergeMap[K comparable, T any](src map[K]T, vals ...map[K]T) map[K]T {
	for _, val := range vals {
		for k, v := range val {
			src[k] = v
		}
	}

	return src
}

func SubMap[K comparable, T any](newVals, oldVals map[K]T) map[K]T {
	addVals := make(map[K]T, 0)
	for k, v := range newVals {
		if _, ok := oldVals[k]; ok {
			continue
		}

		addVals[k] = v
	}

	return addVals
}

// DiffMaps returns two additions, the additions that need to be added, and the ones that need to be deleted
func DiffMaps[K comparable, T any](newVals, oldVals map[K]T) (map[K]T, map[K]T) {
	return SubMap(newVals, oldVals), SubMap(oldVals, newVals)
}

func ToMapstructureWithJSONTag(inp interface{}) (map[string]interface{}, error) {
	byts, err := json.Marshal(inp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert state to json")
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(byts, &obj); err != nil {
		return nil, errors.Wrap(err, "unable to convert to map[string]interface{}")
	}

	return obj, nil
}

func ToMapstructure(inp interface{}) (map[string]interface{}, error) {
	var obj map[string]interface{}
	if err := mapstructure.Decode(inp, &obj); err != nil {
		return nil, fmt.Errorf("unable to decode to mapstructure: %w", err)
	}

	return obj, nil
}

func ToIntMap[T any](inp map[string]T) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range inp {
		out[k] = v
	}

	return out
}

func ToStringMap(inp map[string]interface{}) map[string]string {
	out := make(map[string]string)
	for k, v := range inp {
		vStr, ok := v.(string)
		if ok {
			out[k] = vStr
			continue
		}

		out[k] = fmt.Sprintf("%v", v)
	}

	return out
}

func MapToKeys[T comparable, V any](in map[T]V) []T {
	out := make([]T, 0, len(in))

	for k := range in {
		out = append(out, k)
	}

	return out
}

// Merges source and destination map, preferring values from the source map
// Taken from github.com/helm/pkg/cli/values/options.go
func MergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

// StringToMapDecodeHook decodes hstore values back into maps.
//
// Two on-disk formats are accepted:
//
//  1. JSON object — the format written by EncodeNestedForHstore for new rows.
//     Whitespace, colons, and brackets in values round-trip losslessly.
//  2. Legacy fmt.Sprintf("%v", map) text shaped like "map[k1:v1 k2:v2]".
//     Pre-existing rows still use this format. The parser is lossy on any
//     value containing whitespace and is preserved only for backward
//     compatibility — new writes must go through EncodeNestedForHstore.
func StringToMapDecodeHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		// 1. Check if the target is an interface{} or map[string]string(any)
		if t.Kind() != reflect.Interface && t != reflect.TypeOf(map[string]string{}) {
			return data, nil // Pass through if types don't match
		}

		var raw string
		if str, ok := data.(string); ok {
			raw = str
		} else if strPtr, ok := data.(*string); ok {
			raw = FromPtrStr(strPtr)
		} else {
			return data, nil // Skip if not string or *string
		}

		if strings.HasPrefix(raw, "{") {
			if t == reflect.TypeOf(map[string]string{}) {
				out := map[string]string{}
				if err := json.Unmarshal([]byte(raw), &out); err == nil {
					return out, nil
				}
			} else {
				var out any
				if err := json.Unmarshal([]byte(raw), &out); err == nil {
					return out, nil
				}
			}
			// JSON parse failed — fall through to legacy parser. Defensive
			// only: EncodeNestedForHstore always emits valid JSON.
		}

		if !strings.HasPrefix(raw, "map[") || !strings.HasSuffix(raw, "]") {
			return data, nil // Skip if not in the expected format
		}

		// 1. Strip "map[" and "]"
		content := raw[4 : len(raw)-1]

		resultMap := make(map[string]string)

		// 2. Parse multiple key:value pairs
		// Split by space to get individual key:value pairs
		for pair := range strings.FieldsSeq(content) {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) == 2 {
				resultMap[parts[0]] = parts[1]
			}
		}

		return resultMap, nil
	}
}

// EncodeNestedForHstore returns a copy of data where any map-valued entry has
// been replaced with its JSON encoding. Scalars and slices pass through
// unchanged.
//
// Hstore stores map[string]string, so nested map values would otherwise be
// flattened by ToStringMap via fmt.Sprintf("%v", v) into "map[k:v]" text —
// a format StringToMapDecodeHook reads back by splitting on whitespace, which
// silently truncates any value containing a space, colon, or bracket. JSON
// encoding the nested level once at write time round-trips losslessly via the
// JSON branch of StringToMapDecodeHook.
//
// Only top-level map values are encoded. Slices are left alone because the
// composed decoder pairs this hook with mapstructure.StringToSliceHookFunc(","),
// which expects comma-separated strings, not JSON arrays.
func EncodeNestedForHstore(data map[string]any) map[string]any {
	out := make(map[string]any, len(data))
	for k, v := range data {
		rv := reflect.ValueOf(v)
		if rv.IsValid() && rv.Kind() == reflect.Map {
			if b, err := json.Marshal(v); err == nil {
				out[k] = string(b)
				continue
			}
		}
		out[k] = v
	}
	return out
}

// parseMap parses input map string into map[string]interface{}
// parses nested maps as well
// primarily crafted for roles which have ':' in them
// Beta: use with caution
func parseMap(input string) map[string]interface{} {
	// 1. Strip "map[" and "]"
	content := input[4 : len(input)-1]

	resultMap := make(map[string]interface{})

	// 2. Parse multiple key:value pairs
	// Split by space to get individual key:value pairs
	for pair := range strings.FieldsSeq(content) {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			resultMap[parts[0]] = parts[1]
			if !strings.HasPrefix(parts[1], "map[") || !strings.HasSuffix(parts[1], "]") {
				resultMap[parts[0]] = parseMap(parts[1])
			}
		}
	}

	return resultMap
}

// StringToNestedMapDecodeHook decodes input interface into nested map if needed
// Beta: use with caution, not used anywhere as of now, bugs expected
func StringToNestedMapDecodeHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		// 1. Check if the target is an interface{} or map[string]string(any)
		if t.Kind() != reflect.Interface && t != reflect.TypeOf(map[string]string{}) {
			return data, nil // Pass through if types don't match
		}

		var raw string
		if str, ok := data.(string); ok {
			raw = str
		} else if strPtr, ok := data.(*string); ok {
			raw = FromPtrStr(strPtr)
		} else {
			return data, nil // Skip if not string or *string
		}

		if !strings.HasPrefix(raw, "map[") || !strings.HasSuffix(raw, "]") {
			return data, nil // Skip if not in the expected format
		}

		return parseMap(raw), nil
	}
}
