package errs

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/errbase"
	"github.com/cockroachdb/errors/withstack"
)

func WithUserFacing(err error, format string, args ...any) error {
	// TODO (sdboyer) is it worth creating a stack trace only if there isn't already one in the tree?
	// return errors.WithHint(errors.WithStackDepth(err, 1), msg)
	return errors.WithStackDepth(errors.WithHint(err, fmt.Sprintf(format, args...)), 1)
}

func NewUserFacing(format string, args ...any) error {
	return errors.WithHint(errors.NewWithDepthf(1, format, args...), fmt.Sprintf(format, args...))
}

func HasNuonStackTrace(err error) bool {
	var stacks []*withstack.ReportableStackTrace
	visitAllMulti(err, func(c error) {
		st := withstack.GetReportableStackTrace(c)
		if st != nil {
			stacks = append(stacks, st)
		}
	})

	for _, st := range stacks {
		for _, fr := range st.Frames {
			if strings.Contains(fr.Module, "powertoolsdev") || strings.Contains(fr.Module, "nuonco") {
				return true
			}
		}
	}
	return false
}

func visitAllMulti(err error, f func(error)) {
	f(err)
	if e := errbase.UnwrapOnce(err); e != nil {
		visitAllMulti(e, f)
	}
	for _, e := range errbase.UnwrapMulti(err) {
		visitAllMulti(e, f)
	}
}

func visitAllMultiPostOrder(err error, f func(error)) {
	if e := errbase.UnwrapOnce(err); e != nil {
		visitAllMulti(e, f)
	}
	for _, e := range errbase.UnwrapMulti(err) {
		visitAllMulti(e, f)
	}
	f(err)
}
