package render

import (
	htemplate "html/template"
	ttemplate "text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
)

// disabledFuncs are sprig's non-hermetic functions (minus the date helpers configs rely on) plus functions whose cost
// is controlled by the template author. Templates are rendered inside the control plane, so env/expandenv would leak
// its secrets, getHostByName would resolve attacker-chosen names from inside our network, and the rest could exhaust
// memory or CPU shared across tenants.
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

	// unbounded allocation or CPU inside the shared control plane
	"repeat",
	"until",
	"untilStep",
	"seq",
	"bcrypt",
	"htpasswd",
	"derivePassword",
	"genPrivateKey",
	"buildCustomCert",
	"genCA",
	"genCAWithKey",
	"genSelfSignedCert",
	"genSelfSignedCertWithKey",
	"genSignedCert",
	"genSignedCertWithKey",
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
