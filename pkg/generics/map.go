package generics

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
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
