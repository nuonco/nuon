package installs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

var errInstallComponentOutputsPreviewDisabled = errors.New("[NUON_PREVIEW=false] installs components outputs is a preview feature, set NUON_PREVIEW=true to enable")

func (s *Service) ComponentOutputs(ctx context.Context, installID, componentID string, asJSON bool) error {
	view := ui.NewGetView()
	if !s.cfg.Preview {
		return view.Error(errInstallComponentOutputsPreviewDisabled)
	}

	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return view.Error(err)
	}

	outputs, err := s.api.GetInstallComponentOutputs(ctx, installID, componentID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(outputs)
		return nil
	}

	b, err := json.MarshalIndent(outputs, "", "  ")
	if err != nil {
		return view.Error(err)
	}

	fmt.Println(string(b))
	return nil
}
