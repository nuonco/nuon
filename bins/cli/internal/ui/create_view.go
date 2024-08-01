package ui

import (
	"fmt"

	"github.com/cockroachdb/errors/withstack"
	"github.com/powertoolsdev/mono/pkg/errs"
)

type CreateView struct {
	SpinnerView
	model string
}

func NewCreateView(model string, json bool) *CreateView {
	return &CreateView{
		*NewSpinnerView(json),
		model,
	}
}

func (v *CreateView) Start() {
	v.SpinnerView.Start(fmt.Sprintf("creating %s", v.model))
}

func (v *CreateView) Success(id string) {
	v.SpinnerView.Success(fmt.Sprintf("successfully created %s %s", v.model, id))
}

func (v *CreateView) Fail(err error) error {
	if !errs.HasNuonStackTrace(err) {
		err = withstack.WithStackDepth(err, 1)
	}
	v.SpinnerView.Fail(fmt.Errorf("failed to create %s: %w", v.model, err))
	return err
}
