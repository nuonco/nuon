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

func (t *Templates) Template(inp *stacks.TemplateInput) ([]byte, string, error) {
	tmpl, err := t.getAzureTemplate(inp)
	if err != nil {
		return nil, "", err
	}

	tmplBytes, err := json.MarshalIndent(tmpl, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("unable to marshal ARM template: %w", err)
	}

	hash := sha256.Sum256(tmplBytes)
	checksum := hex.EncodeToString(hash[:])

	return tmplBytes, checksum, nil
}
