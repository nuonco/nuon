package arm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

type Templates struct {
	cfg *internal.Config
}

type Params struct {
	fx.In

	Cfg *internal.Config
}

func NewTemplates(params Params) *Templates {
	return &Templates{
		cfg: params.Cfg,
	}
}

func (t *Templates) runnerAPIURL(inp *stacks.TemplateInput) string {
	if inp.Settings != nil && inp.Settings.RunnerAPIURL != "" {
		return inp.Settings.RunnerAPIURL
	}
	return t.cfg.RunnerAPIURL
}

func (t *Templates) Template(inp *stacks.TemplateInput) ([]byte, string, error) {
	tmpl, err := t.getAzureTemplate(inp)
	if err != nil {
		return nil, "", err
	}

	return marshalTemplate(tmpl)
}

func (t *Templates) CustomStacksTemplate(inp *stacks.TemplateInput) ([]byte, string, map[string]map[string]string, map[string]map[string]string, error) {
	tmpl, outputMap, inputParametersMap, err := t.getAzureCustomStacksOnlyTemplate(inp)
	if err != nil {
		return nil, "", nil, nil, err
	}

	tmplBytes, checksum, err := marshalTemplate(tmpl)
	if err != nil {
		return nil, "", nil, nil, err
	}

	return tmplBytes, checksum, outputMap, inputParametersMap, nil
}

func marshalTemplate(tmpl *ARMTemplate) ([]byte, string, error) {
	tmplBytes, err := json.MarshalIndent(tmpl, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("unable to marshal ARM template: %w", err)
	}

	hash := sha256.Sum256(tmplBytes)
	checksum := hex.EncodeToString(hash[:])

	return tmplBytes, checksum, nil
}
