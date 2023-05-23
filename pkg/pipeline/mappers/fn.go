package mappers

import (
	"context"
	"log"

	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

// ExecFn is a function used to execute a step
type ExecFn func(context.Context, *log.Logger, terminal.UI) ([]byte, error)

// CallbackFn is a function used to send the outputs of an exec, as a callback
type CallbackFn func(context.Context, *log.Logger, terminal.UI, []byte) error

// Mapper is an interface that a pipeline uses to map a step's function into the correct functions within a pipeline
type Mapper interface {
	// GetExecFn returns the correct function to execute, based on the input type
	GetExecFn(context.Context, interface{}) (ExecFn, error)

	// GetCallbackFn returns the correct function to execute, based on the input
	GetCallbackFn(context.Context, interface{}) (CallbackFn, error)
}

type defaultMapper struct{}

// NewDefaultMapper returns a new mapper.
func NewDefaultMapper() *defaultMapper {
	return nil
}
