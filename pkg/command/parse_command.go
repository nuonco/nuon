package command

import (
	"fmt"
	"regexp"
	"strings"
)

// leadingEnv is a KEY=value prefix with no spaces in the value. Quoted values
// are left for the argument splitter.
var leadingEnv = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)=([^ ]+)\s*`)

func ParseCommand(str string) (string, []string, map[string]string, error) {
	cmdEnv := make(map[string]string)
	origStr := str
	str = strings.TrimLeft(str, " \t")

	for {
		matches := leadingEnv.FindStringSubmatch(str)
		if len(matches) == 0 {
			break
		}

		cmdEnv[matches[1]] = matches[2]
		str = strings.TrimLeft(str[len(matches[0]):], " \t")
	}

	if str == "" {
		return "", nil, nil, fmt.Errorf("final command was empty (%s)", origStr)
	}

	args, err := splitCommandArgs(str)
	if err != nil {
		return "", nil, nil, err
	}
	if len(args) == 0 {
		return "", nil, nil, fmt.Errorf("final command was empty (%s)", origStr)
	}

	return args[0], args[1:], cmdEnv, nil
}

// splitCommandArgs splits on whitespace, keeping single- or double-quoted
// sections together. Quotes are removed. A quoted section may sit inside a
// token, so -ldflags="-s -w" is one argument.
func splitCommandArgs(s string) ([]string, error) {
	var args []string
	var buf strings.Builder
	var quoting byte
	started := false

	flush := func() {
		if !started {
			return
		}
		args = append(args, buf.String())
		buf.Reset()
		started = false
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if quoting != 0 {
			if c == quoting {
				quoting = 0
				started = true
				continue
			}
			if quoting == '"' && c == '\\' && i+1 < len(s) {
				i++
				buf.WriteByte(s[i])
				started = true
				continue
			}
			buf.WriteByte(c)
			started = true
			continue
		}

		switch c {
		case ' ', '\t':
			flush()
		case '"', '\'':
			quoting = c
			started = true
		default:
			buf.WriteByte(c)
			started = true
		}
	}
	if quoting != 0 {
		return nil, fmt.Errorf("unterminated quote in command: %s", s)
	}
	flush()
	return args, nil
}
