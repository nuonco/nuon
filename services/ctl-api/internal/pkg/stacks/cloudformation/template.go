package cloudformation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (t *Templates) Template(inputs *stacks.TemplateInput) (*cloudformation.Template, string, error) {
	tmpl, err := t.getAWSTemplate(inputs)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to create aws-eks template")
	}

	// Marshal the template to JSON
	jsonBytes, err := json.Marshal(tmpl)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to marshal template to JSON")
	}

	hash := sha256.Sum256(jsonBytes)
	checksum := hex.EncodeToString(hash[:])

	return tmpl, checksum, nil
}
