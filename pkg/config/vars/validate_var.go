package vars

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/render"
)

func (v *varsValidator) validateVarV2(inputVar string, tmplData map[string]interface{}) error {
	// vars are namespaced with the install so need to remove it since we template against a install state.
	// we may want to change this later if we plan to expose first level vars
	inputVar = strings.ReplaceAll(inputVar, ".install.", ".")
	_, err := render.RenderV2(inputVar, tmplData)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to render %s", inputVar))
	}

	return nil
}
