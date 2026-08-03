package cloudformation

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (t *Templates) Template(inputs *stacks.TemplateInput) ([]byte, string, error) {
	tmpl, err := t.getAWSTemplate(inputs)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to create aws template")
	}

	jsonBytes, err := tmpl.JSON()
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to marshal template to JSON")
	}

	jsonBytes, err = injectTargetAccountRule(jsonBytes, inputs)
	if err != nil {
		return nil, "", err
	}

	hash := sha256.Sum256(jsonBytes)
	checksum := hex.EncodeToString(hash[:])

	return jsonBytes, checksum, nil
}
