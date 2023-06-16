package output

import (
	"bytes"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
)

type Dual interface {
	Bytes() ([]byte, error)
	Writer() (io.Writer, error)
}

var _ Dual = (*dual)(nil)

type dual struct {
	v *validator.Validate

	Logger hclog.Logger `validate:"required"`
	buf    *bytes.Buffer
}

type dualOption func(*dual) error

func New(v *validator.Validate, opts ...dualOption) (*dual, error) {
	d := &dual{
		v:   v,
		buf: new(bytes.Buffer),
	}
	for _, opt := range opts {
		if err := opt(d); err != nil {
			return nil, err
		}
	}

	if err := d.v.Struct(d); err != nil {
		return nil, err
	}

	return d, nil
}

// WithLogger specifies the log that will be used to output to
func WithLogger(lg hclog.Logger) dualOption {
	return func(d *dual) error {
		d.Logger = lg
		return nil
	}
}
