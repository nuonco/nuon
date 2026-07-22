package eventfilter

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/theory/jsonpath"
	"github.com/theory/jsonpath/spec"
)

type Source string
type Operator string

const (
	SourcePayload Source = "payload"
	SourceHeaders Source = "headers"

	OperatorEq        Operator = "eq"
	OperatorNEq       Operator = "neq"
	OperatorIn        Operator = "in"
	OperatorPrefix    Operator = "prefix"
	OperatorSuffix    Operator = "suffix"
	OperatorContains  Operator = "contains"
	OperatorGT        Operator = "gt"
	OperatorGTE       Operator = "gte"
	OperatorLT        Operator = "lt"
	OperatorLTE       Operator = "lte"
	OperatorRegex     Operator = "regex"
	OperatorExists    Operator = "exists"
	OperatorNotExists Operator = "not_exists"

	maxRegexLength = 1024
)

type Filter struct {
	From  Source
	Path  string
	Op    Operator
	Value any
}

type Compiled struct {
	filter Filter
	path   *jsonpath.Path
	regex  *regexp.Regexp
}

type Result struct {
	Matched  bool
	Selected []any
}

func Compile(filter Filter) (*Compiled, error) {
	if filter.From == "" {
		filter.From = SourcePayload
	}
	compiled := &Compiled{filter: filter}
	switch filter.From {
	case SourcePayload:
		path, err := ParsePath(filter.Path, true)
		if err != nil {
			return nil, err
		}
		compiled.path = path
	case SourceHeaders:
		if strings.TrimSpace(filter.Path) == "" {
			return nil, errors.New("header filter path is required")
		}
	default:
		return nil, fmt.Errorf("unsupported filter source %q", filter.From)
	}
	if err := validateValue(filter.Op, filter.Value); err != nil {
		return nil, err
	}
	if filter.Op == OperatorRegex {
		pattern := filter.Value.(string)
		if len(pattern) > maxRegexLength {
			return nil, fmt.Errorf("regex exceeds %d bytes", maxRegexLength)
		}
		compiled.regex = regexp.MustCompile(pattern)
	}
	return compiled, nil
}

func ParsePath(value string, allowWildcard bool) (*jsonpath.Path, error) {
	path, err := jsonpath.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("invalid JSONPath: %w", err)
	}
	for _, segment := range path.Query().Segments() {
		if segment.IsDescendant() {
			return nil, errors.New("JSONPath recursive descent is not supported")
		}
		selectors := segment.Selectors()
		if len(selectors) != 1 {
			return nil, errors.New("JSONPath unions are not supported")
		}
		switch selector := selectors[0].(type) {
		case spec.Name:
		case spec.Index:
			if selector < 0 {
				return nil, errors.New("JSONPath negative indexes are not supported")
			}
		case spec.WildcardSelector:
			if !allowWildcard {
				return nil, errors.New("JSONPath wildcards are not allowed in this context")
			}
		default:
			return nil, fmt.Errorf("unsupported JSONPath selector %T", selector)
		}
	}
	return path, nil
}

func IsPositive(op Operator) bool {
	return op != OperatorNEq && op != OperatorNotExists
}

func (c *Compiled) Evaluate(payload any, headers http.Header) Result {
	selected := c.selectNodes(payload, headers)
	result := Result{Selected: selected}
	switch c.filter.Op {
	case OperatorExists:
		result.Matched = len(selected) != 0
	case OperatorNotExists:
		result.Matched = len(selected) == 0
	case OperatorNEq:
		if len(selected) == 0 {
			return result
		}
		result.Matched = true
		for _, node := range selected {
			equal, comparable := valuesEqual(node, c.filter.Value)
			if !comparable || equal {
				result.Matched = false
				break
			}
		}
	default:
		for _, node := range selected {
			if c.matchesNode(node) {
				result.Matched = true
				break
			}
		}
	}
	return result
}

func (c *Compiled) selectNodes(payload any, headers http.Header) []any {
	if c.filter.From == SourceHeaders {
		values := headers.Values(c.filter.Path)
		selected := make([]any, len(values))
		for i := range values {
			selected[i] = values[i]
		}
		return selected
	}
	return c.path.Select(payload)
}

func (c *Compiled) matchesNode(node any) bool {
	switch c.filter.Op {
	case OperatorEq:
		equal, _ := valuesEqual(node, c.filter.Value)
		return equal
	case OperatorIn:
		values := reflect.ValueOf(c.filter.Value)
		for i := 0; i < values.Len(); i++ {
			equal, _ := valuesEqual(node, values.Index(i).Interface())
			if equal {
				return true
			}
		}
	case OperatorPrefix, OperatorSuffix, OperatorContains:
		actual, ok := node.(string)
		if !ok {
			return false
		}
		expected := c.filter.Value.(string)
		return c.filter.Op == OperatorPrefix && strings.HasPrefix(actual, expected) ||
			c.filter.Op == OperatorSuffix && strings.HasSuffix(actual, expected) ||
			c.filter.Op == OperatorContains && strings.Contains(actual, expected)
	case OperatorGT, OperatorGTE, OperatorLT, OperatorLTE:
		actual, ok := decimal(node)
		if !ok {
			return false
		}
		expected, _ := decimal(c.filter.Value)
		comparison := actual.Cmp(expected)
		return c.filter.Op == OperatorGT && comparison > 0 || c.filter.Op == OperatorGTE && comparison >= 0 ||
			c.filter.Op == OperatorLT && comparison < 0 || c.filter.Op == OperatorLTE && comparison <= 0
	case OperatorRegex:
		actual, ok := node.(string)
		return ok && c.regex.MatchString(actual)
	}
	return false
}

func validateValue(op Operator, value any) error {
	switch op {
	case OperatorExists, OperatorNotExists:
		return nil
	case OperatorEq, OperatorNEq:
		if isPrimitive(value) {
			return nil
		}
		return errors.New("filter value must be a JSON primitive")
	case OperatorIn:
		kind := reflect.Invalid
		if value != nil {
			kind = reflect.TypeOf(value).Kind()
		}
		if kind != reflect.Array && kind != reflect.Slice {
			return errors.New("in filter value must be an array")
		}
		values := reflect.ValueOf(value)
		if values.Len() == 0 {
			return errors.New("in filter value must not be empty")
		}
		for i := 0; i < values.Len(); i++ {
			if !isPrimitive(values.Index(i).Interface()) {
				return errors.New("in filter values must be JSON primitives")
			}
		}
		return nil
	case OperatorPrefix, OperatorSuffix, OperatorContains:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s filter value must be a string", op)
		}
		return nil
	case OperatorGT, OperatorGTE, OperatorLT, OperatorLTE:
		if _, ok := decimal(value); !ok {
			return fmt.Errorf("%s filter value must be a number", op)
		}
		return nil
	case OperatorRegex:
		pattern, ok := value.(string)
		if !ok {
			return errors.New("regex filter value must be a string")
		}
		if len(pattern) > maxRegexLength {
			return fmt.Errorf("regex exceeds %d bytes", maxRegexLength)
		}
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid regex: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported filter operator %q", op)
	}
}

func valuesEqual(left, right any) (bool, bool) {
	if leftNumber, ok := decimal(left); ok {
		rightNumber, rightOK := decimal(right)
		return rightOK && leftNumber.Cmp(rightNumber) == 0, rightOK
	}
	if left == nil || right == nil {
		return left == nil && right == nil, left == nil && right == nil
	}
	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return false, false
	}
	return reflect.DeepEqual(left, right), true
}

func decimal(value any) (*big.Rat, bool) {
	var encoded string
	switch value := value.(type) {
	case json.Number:
		encoded = value.String()
	case int, int8, int16, int32, int64:
		encoded = strconv.FormatInt(reflect.ValueOf(value).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		encoded = strconv.FormatUint(reflect.ValueOf(value).Uint(), 10)
	case float32:
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return nil, false
		}
		encoded = strconv.FormatFloat(float64(value), 'g', -1, 32)
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, false
		}
		encoded = strconv.FormatFloat(value, 'g', -1, 64)
	default:
		return nil, false
	}
	number, ok := new(big.Rat).SetString(encoded)
	return number, ok
}

func isPrimitive(value any) bool {
	if value == nil {
		return true
	}
	switch value.(type) {
	case bool, string:
		return true
	}
	_, ok := decimal(value)
	return ok
}
