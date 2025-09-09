package ui

import (
	"github.com/cockroachdb/errors/withstack"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/bubbles"
	"github.com/powertoolsdev/mono/pkg/errs"
)

type GetView struct {
	tableView *bubbles.TableView
}

func NewGetView() *GetView {
	return &GetView{
		tableView: bubbles.NewTableView(),
	}
}

func (v *GetView) Print(msg string) {
	v.tableView.Print(msg)
}

func (v *GetView) Render(data [][]string) {
	v.tableView.Render(data)
}

func (v *GetView) Error(err error) error {
	if !errs.HasNuonStackTrace(err) {
		err = withstack.WithStackDepth(err, 1)
	}
	PrintError(err)
	return err
}
