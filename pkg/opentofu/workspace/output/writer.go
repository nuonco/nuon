package output

import (
	"io"

	"github.com/hashicorp/go-hclog"
)

func (d *dual) Writer() (io.Writer, error) {
	lgWriter := d.Logger.StandardWriter(&hclog.StandardLoggerOptions{})
	writer := io.MultiWriter(d.buf, lgWriter)
	return writer, nil
}
