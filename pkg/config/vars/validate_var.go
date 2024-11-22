package vars

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

func (v *varsValidator) validateVar(inputVar string, tmplData map[string]interface{}) error {
	re := regexp.MustCompile(`\{\{(.*?)\}\}`)
	matches := re.FindAllStringSubmatch(inputVar, -1)

	for _, matchP := range matches {
		// get the current match, minus go strings, for the current depth
		match := matchP[0]
		match = strings.ReplaceAll(match, "{{", "")
		match = strings.ReplaceAll(match, "}}", "")
		match = strings.Replace(match, ".", "", 1)

		// this could only happen if a user added `{{}}`
		if match == "" {
			return nil
		}

		// split the input value into strings and validate one layer of depth at a time
		newPieces := make([]string, 0)
		matchPieces := strings.Split(match, ".")
		if len(matchPieces) < 1 {
			return nil
		}

		var isSandbox bool
		for _, matchPiece := range matchPieces {
			if matchPiece == "sandbox" {
				isSandbox = true
			}

			if !isSandbox && matchPiece == "outputs" {
				break
			}

			newPieces = append(newPieces, matchPiece)
		}

		newTmpl := fmt.Sprintf("{{.%s}}", strings.Join(newPieces, "."))

		temp, err := template.New("input").Option("missingkey=zero").Parse(newTmpl)
		if err != nil {
			return errors.Wrap(err, "unable to create template")
		}

		buf := new(bytes.Buffer)
		if err := temp.Execute(buf, tmplData); err != nil {
			return errors.Wrap(err, "unabel to execute template")
		}

		outputVal := buf.String()
		if outputVal == "" {
			return fmt.Errorf("checked value was empty: %s (original) %s", newTmpl, inputVar)
		}

		if outputVal == "" || outputVal == "<no value>" {
			return fmt.Errorf("invalid reference. Not found in intermediate data.")
		}
	}

	return nil
}
