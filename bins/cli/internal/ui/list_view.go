package ui

import (
	"github.com/cockroachdb/errors/withstack"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/bubbles"
	"github.com/powertoolsdev/mono/pkg/errs"
)

type ListView struct {
	tableView *bubbles.TableView
}

func NewListView() *ListView {
	return &ListView{
		tableView: bubbles.NewTableView(),
	}
}

func (v *ListView) Render(data [][]string) {
	v.tableView.Render(data)
}

func (v *ListView) RenderPaging(data [][]string, offset, limit int, hasMore bool) {
	v.tableView.RenderPaging(data, offset, limit, hasMore)
}

func (v *ListView) Error(err error) error {
	if !errs.HasNuonStackTrace(err) {
		err = withstack.WithStackDepth(err, 1)
	}
	return PrintError(err)
}

func (v *ListView) Print(msg string) {
	v.tableView.Print(msg)
}
