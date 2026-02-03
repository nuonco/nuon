package policies

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) ExportReport(ctx context.Context, reportID, format, output string, asJSON bool) error {
	view := ui.NewGetView()

	data, contentType, err := s.api.ExportPolicyReport(ctx, reportID, format)
	if err != nil {
		return view.Error(errors.Wrap(err, "failed to export policy report"))
	}

	if output != "" {
		dir := filepath.Dir(output)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return view.Error(errors.Wrap(err, "failed to create output directory"))
			}
		}

		if err := os.WriteFile(output, data, 0644); err != nil {
			return view.Error(errors.Wrap(err, "failed to write output file"))
		}

		fmt.Printf("Report exported to: %s\n", output)
		return nil
	}

	// If no output file specified, print to stdout
	if contentType == "application/pdf" {
		return view.Error(errors.New("PDF output requires --output flag to specify a file path"))
	}

	fmt.Println(string(data))
	return nil
}
