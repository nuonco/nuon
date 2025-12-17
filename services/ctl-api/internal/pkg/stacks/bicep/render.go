package bicep

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"text/template"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	"github.com/pkg/errors"
)

func Render(inputs *stacks.TemplateInput) ([]byte, string, error) {
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
