package zaphclog

import (
	"bufio"
	"bytes"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
)

type zaphclogWriter struct {
	l hclog.Logger
}

func (z *zaphclogWriter) Write(byts []byte) (int, error) {
	buf := bytes.NewBuffer(byts)

	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		err := z.writeTerraform(scanner.Bytes())
		if err == nil {
			continue
		}

		z.l.Info(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return 0, errors.Wrap(err, "unable to scan output")
	}

	return len(byts), nil
}
