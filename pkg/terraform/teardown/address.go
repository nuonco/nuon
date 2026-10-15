package teardown

import (
	"fmt"
	"strings"
)

type step struct {
	name     string
	index    string
	hasIndex bool
	module   bool
	data     bool
}

type address []step

func (a address) isModulePath() bool {
	return len(a) > 0 && a[len(a)-1].module
}

// matches reports whether a, used as a retain entry, selects target. Steps
// without an index select every instance of that step, and a module path
// selects everything beneath it.
func (a address) matches(target address) bool {
	if len(a) > len(target) {
		return false
	}
	if !a.isModulePath() && len(a) != len(target) {
		return false
	}

	for idx := range a {
		if a[idx].module != target[idx].module || a[idx].name != target[idx].name {
			return false
		}
		if a[idx].hasIndex && (!target[idx].hasIndex || a[idx].index != target[idx].index) {
			return false
		}
	}

	return true
}

// ValidateAddress checks that addr is a terraform module or managed resource
// address that can be passed to `terraform state rm`.
func ValidateAddress(addr string) error {
	parsed, err := parseAddress(addr)
	if err != nil {
		return err
	}

	for _, s := range parsed {
		if s.data {
			return fmt.Errorf("%q is a data source, only managed resources and modules can be retained", addr)
		}
	}

	return nil
}

func parseAddress(addr string) (address, error) {
	tokens, err := splitTokens(addr)
	if err != nil {
		return nil, err
	}

	var parsed address
	for idx := 0; idx < len(tokens); {
		tok := tokens[idx]

		switch {
		case tok.ident == "module":
			if tok.hasIndex || idx+1 >= len(tokens) {
				return nil, fmt.Errorf("%q: module must be followed by a module name", addr)
			}
			name := tokens[idx+1]
			parsed = append(parsed, step{
				name:     name.ident,
				index:    name.index,
				hasIndex: name.hasIndex,
				module:   true,
			})
			idx += 2
		case tok.ident == "data":
			if tok.hasIndex || idx+2 >= len(tokens) || idx+3 != len(tokens) || tokens[idx+1].hasIndex {
				return nil, fmt.Errorf("%q: invalid data source address", addr)
			}
			name := tokens[idx+2]
			parsed = append(parsed, step{
				name:     tokens[idx+1].ident + "." + name.ident,
				index:    name.index,
				hasIndex: name.hasIndex,
				data:     true,
			})
			idx += 3
		default:
			if tok.hasIndex || idx+2 != len(tokens) {
				return nil, fmt.Errorf("%q: expected a resource address of the form <type>.<name> or a module path", addr)
			}
			name := tokens[idx+1]
			parsed = append(parsed, step{
				name:     tok.ident + "." + name.ident,
				index:    name.index,
				hasIndex: name.hasIndex,
			})
			idx += 2
		}
	}

	return parsed, nil
}

type token struct {
	ident    string
	index    string
	hasIndex bool
}

func splitTokens(addr string) ([]token, error) {
	if strings.TrimSpace(addr) != addr || addr == "" {
		return nil, fmt.Errorf("%q: address must be non-empty and have no surrounding whitespace", addr)
	}

	var tokens []token
	pos := 0
	for {
		start := pos
		for pos < len(addr) && isIdentChar(addr[pos]) {
			pos++
		}
		if start == pos || !isIdentStart(addr[start]) {
			return nil, fmt.Errorf("%q: invalid identifier at position %d", addr, start)
		}
		tok := token{ident: addr[start:pos]}

		if pos < len(addr) && addr[pos] == '[' {
			end, err := indexEnd(addr, pos)
			if err != nil {
				return nil, err
			}
			tok.index = addr[pos+1 : end]
			tok.hasIndex = true
			pos = end + 1
		}
		tokens = append(tokens, tok)

		if pos == len(addr) {
			return tokens, nil
		}
		if addr[pos] != '.' {
			return nil, fmt.Errorf("%q: unexpected %q at position %d", addr, addr[pos], pos)
		}
		pos++
	}
}

// indexEnd returns the position of the "]" closing the index that opens at
// pos. Indexes are either integers or double-quoted strings.
func indexEnd(addr string, pos int) (int, error) {
	inner := pos + 1
	if inner < len(addr) && addr[inner] == '"' {
		for i := inner + 1; i < len(addr); i++ {
			switch addr[i] {
			case '\\':
				i++
			case '"':
				if i+1 < len(addr) && addr[i+1] == ']' {
					return i + 1, nil
				}
				return 0, fmt.Errorf("%q: expected ] after string index at position %d", addr, i+1)
			}
		}
		return 0, fmt.Errorf("%q: unterminated string index", addr)
	}

	i := inner
	for i < len(addr) && addr[i] >= '0' && addr[i] <= '9' {
		i++
	}
	if i == inner || i >= len(addr) || addr[i] != ']' {
		return 0, fmt.Errorf("%q: index at position %d must be an integer or a double-quoted string", addr, pos)
	}
	return i, nil
}

func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentChar(c byte) bool {
	return isIdentStart(c) || c == '-' || (c >= '0' && c <= '9')
}
