package ui

import (
	"fmt"

	"github.com/cockroachdb/errors/withstack"
	"github.com/powertoolsdev/mono/pkg/errs"
)

type DeleteView struct {
	SpinnerView
	model string
	id    string
}

func NewDeleteView(model string, id string) *DeleteView {
	return &DeleteView{
		*NewSpinnerView(false),
		id,
		model,
	}
}

func (v *DeleteView) Start() {
	v.SpinnerView.Start(fmt.Sprintf("deleting %s %s", v.model, v.id))
}

func (v *DeleteView) Success() {
	v.SpinnerView.Success(fmt.Sprintf("successfully deleted %s %s", v.model, v.id))
}

func (v *DeleteView) SuccessQueued() {
	v.SpinnerView.Success(fmt.Sprintf("successfully queued %s to be deleted %s", v.id, v.model))
}

func (v *DeleteView) Fail(err error) error {
	if !errs.HasNuonStackTrace(err) {
		err = withstack.WithStackDepth(err, 1)
	}
	v.SpinnerView.Fail(fmt.Errorf("failed to delete %s: %w", v.model, err))
	return err
}
