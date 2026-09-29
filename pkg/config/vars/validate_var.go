package vars

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/render"
)

func (v *varsValidator) validateVar(inputVar string, tmplData map[string]interface{}) error {
	re := regexp.MustCompile(`\{\{(.*?)\}\}`)
	matches := re.FindAllStringSubmatch(inputVar, -1)

	for _, matchP := range matches {
		match := matchP[0]
		match = strings.ReplaceAll(match, "{{", "")
		match = strings.ReplaceAll(match, "}}", "")
		match = strings.Replace(match, ".", "", 1)

		if match == "" {
			return nil
		}

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

func (v *varsValidator) validateVarV2(inputVar string, tmplData map[string]interface{}) error {
	inputVar = strings.ReplaceAll(inputVar, ".install.", ".")
	_, err := render.RenderV2(inputVar, tmplData)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to render %s", inputVar))
	}

	return nil
}
