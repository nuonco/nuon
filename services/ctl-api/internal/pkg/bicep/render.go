package bicep

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"text/template"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cloudformation"
)

func Render(inputs *cloudformation.TemplateInput) ([]byte, string, error) {
	t, err := template.New("bicep-stack").Parse(tmpl)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to parse bicep template")
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, inputs)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to execute bicep template")
	}
	res := buf.Bytes()

	hash := sha256.Sum256(res)
	checksum := hex.EncodeToString(hash[:])

	return res, checksum, nil
}
