package ui

import (
	"github.com/cockroachdb/errors/withstack"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/pkg/errs"
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

func (v *ListView) RenderPagingWithContext(data [][]string, offset, limit int, hasMore bool, contextLabel, contextValue string) {
	v.tableView.RenderPagingWithContext(data, offset, limit, hasMore, contextLabel, contextValue)
}

func (v *ListView) RenderTotal(data [][]string, total int) {
	v.tableView.RenderTotal(data, total)
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
