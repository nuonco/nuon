package vars

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/render"
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

		newTmpl := fmt.Sprintf("{{.%s}}", strings.TrimSpace(strings.Join(newPieces, ".")))

		rendered, err := render.RenderV2(newTmpl, tmplData)
		if err != nil {
			return errors.Wrap(err, "unable to render "+newTmpl)
		}

		if rendered == "" {
			return config.ErrConfig{
				Warning:     true,
				Description: fmt.Sprintf("rendered variable %s was empty", newTmpl),
				Err:         fmt.Errorf("rendered variable %s was empty", newTmpl),
			}
		}
	}

	return nil
}
