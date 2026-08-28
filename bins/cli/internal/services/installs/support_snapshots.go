package installs

import (
	"context"
	"os"
	"strconv"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) UploadSupportSnapshot(ctx context.Context, installID, path string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	archive, err := os.Open(path)
	if err != nil {
		return ui.PrintError(err)
	}
	defer archive.Close()

	snapshot, err := s.api.CreateInstallSupportSnapshot(ctx, installID, archive)
	if err != nil {
		return ui.PrintError(err)
	}
	return printSupportSnapshot(snapshot, asJSON)
}

func (s *Service) ListSupportSnapshots(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	snapshots, err := s.api.ListInstallSupportSnapshots(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(snapshots)
		return nil
	}

	rows := [][]string{{"ID", "CAPTURED AT", "SIZE", "INTEGRITY", "ASSOCIATION"}}
	for _, snapshot := range snapshots {
		rows = append(rows, []string{
			snapshot.ID,
			snapshot.CapturedAt,
			strconv.FormatInt(snapshot.ArchiveSize, 10),
			snapshot.IntegrityStatus,
			snapshot.AssociationStatus,
		})
	}
	ui.NewListView().Render(rows)
	return nil
}

func (s *Service) GetSupportSnapshot(ctx context.Context, installID, snapshotID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	snapshot, err := s.api.GetInstallSupportSnapshot(ctx, installID, snapshotID)
	if err != nil {
		return ui.PrintError(err)
	}
	return printSupportSnapshot(snapshot, asJSON)
}

func printSupportSnapshot(snapshot *models.ServiceSupportSnapshotResponse, asJSON bool) error {
	if asJSON {
		ui.PrintJSON(snapshot)
		return nil
	}
	ui.NewGetView().Render([][]string{
		{"id", snapshot.ID},
		{"install id", snapshot.InstallID},
		{"captured at", snapshot.CapturedAt},
		{"created at", snapshot.CreatedAt},
		{"size (bytes)", strconv.FormatInt(snapshot.ArchiveSize, 10)},
		{"sha256", snapshot.ArchiveSha256},
		{"integrity", snapshot.IntegrityStatus},
		{"association", snapshot.AssociationStatus},
	})
	return nil
}
