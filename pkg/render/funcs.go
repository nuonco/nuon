package render

import (
	htemplate "html/template"
	ttemplate "text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
)

// disabledFuncs are sprig's non-hermetic functions, minus the date helpers configs rely on. Templates are rendered
// inside the control plane, so env/expandenv would leak its secrets and getHostByName would resolve attacker-chosen
// names from inside our network.
var disabledFuncs = []string{
	"date_in_zone",
	"date_modify",
	"htmlDateInZone",
	"dateInZone",
	"dateModify",

	"randAlphaNum",
	"randAlpha",
	"randAscii",
	"randNumeric",
	"randBytes",
	"uuidv4",

	"env",
	"expandenv",

	"getHostByName",
}

// TxtFuncMap returns the function set for rendering user-supplied templates with "text/template".
func TxtFuncMap() ttemplate.FuncMap {
	fm := sprig.TxtFuncMap()
	applyOverrides(fm)
	return fm
}

// HTMLFuncMap returns the function set for rendering user-supplied templates with "html/template".
func HTMLFuncMap() htemplate.FuncMap {
	fm := sprig.HtmlFuncMap()
	applyOverrides(fm)
	return fm
}

func applyOverrides(fm map[string]any) {
	fm["now"] = time.Now
	for _, name := range disabledFuncs {
		fm[name] = disabledFunc(name)
	}
}

func disabledFunc(name string) func(...any) (string, error) {
	return func(...any) (string, error) {
		return "", errors.Errorf("template function %q is not allowed", name)
	}
}
