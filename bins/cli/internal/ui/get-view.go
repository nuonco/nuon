package ui

import (
	"github.com/pterm/pterm"
)

type GetView struct {
	headerColumn []string
}

func NewGetView(headerColumn []string) *GetView {
	return &GetView{
		headerColumn: headerColumn,
	}
}

func (v *GetView) Render(item []string) {
	data := [][]string{}
	for i, v := range v.headerColumn {
		row := []string{
			v,
			item[i],
		}
		data = append(data, row)
	}
	pterm.DefaultTable.
		WithData(data).
		Render()
}

func (v *GetView) Error(err error) {
	pterm.Error.Println(err)
}
