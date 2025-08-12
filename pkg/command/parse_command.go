package command

import (
	"fmt"
	"regexp"
	"strings"
)

func ParseCommand(str string) (string, []string, map[string]string, error) {
	cmdEnv := make(map[string]string, 0)

	origStr := str
	str = strings.TrimLeft(str, " \t")

	re := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)=([^ ]+)\s*`)
	for {
		matches := re.FindStringSubmatch(str)
		if len(matches) == 0 {
			break
		}

		cmdEnv[matches[1]] = matches[2]
		str = strings.TrimLeft(str[len(matches[0]):], " \t")
	}

	if str == "" {
		return "", nil, nil, fmt.Errorf("final command was empty (%s)", origStr)
	}

	pieces := strings.Split(str, " ")
	return pieces[0], pieces[1:], cmdEnv, nil
}
